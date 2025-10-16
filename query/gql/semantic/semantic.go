package semantic

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/cayleygraph/cayley/query/gql/diagnostic"
	"github.com/cayleygraph/cayley/query/gql/parser"
)

type GraphRole struct {
	Name     string
	CanRead  bool
	CanWrite bool
}

type Schema struct {
	Name  string
	Types map[string]struct{}
}

type Graph struct {
	Name          string
	DefaultSchema string
	Schemas       map[string]Schema
	Roles         map[string]GraphRole
}

type Catalog interface {
	LookupGraph(ctx context.Context, name string) (*Graph, error)
	DefaultGraph(ctx context.Context) (*Graph, error)
}

type Options struct {
	Role          string
	DefaultGraph  string
	DefaultSchema string
}

type CheckedStatement struct {
	Statement parser.Statement
	Graph     *Graph
	Schema    string
	Variables []string
}

type Result struct {
	Statements []CheckedStatement
}

type Validator struct {
	catalog Catalog
}

func NewValidator(cat Catalog) *Validator {
	return &Validator{catalog: cat}
}

func (v *Validator) Validate(ctx context.Context, script *parser.Script, opt Options) (*Result, error) {
	if script == nil || len(script.Statements) == 0 {
		return &Result{}, nil
	}
	var (
		currentGraph  *Graph
		currentSchema = opt.DefaultSchema
		role          = opt.Role
		diags         []diagnostic.Diagnostic
		checked       []CheckedStatement
	)
	if opt.DefaultGraph != "" {
		g, err := v.catalog.LookupGraph(ctx, opt.DefaultGraph)
		if err == nil && g != nil {
			currentGraph = g
		}
	}
	if currentGraph == nil {
		g, err := v.catalog.DefaultGraph(ctx)
		if err == nil {
			currentGraph = g
		}
	}
	if currentGraph != nil && currentSchema == "" {
		currentSchema = currentGraph.DefaultSchema
	}
	for _, stmt := range script.Statements {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		switch s := stmt.(type) {
		case *parser.UseGraphStatement:
			g, err := v.catalog.LookupGraph(ctx, s.Graph)
			if err != nil || g == nil {
				diags = append(diags, diagnostic.Diagnostic{
					Severity:  diagnostic.SeverityError,
					Message:   fmt.Sprintf("graph %q not found", s.Graph),
					Statement: s.Text(),
					Line:      s.Pos().Line,
					Column:    s.Pos().Column,
					Code:      "CATALOG_GRAPH_UNDEFINED",
				})
				continue
			}
			currentGraph = g
			currentSchema = g.DefaultSchema
			checked = append(checked, CheckedStatement{Statement: s, Graph: currentGraph, Schema: currentSchema})
		case *parser.MatchStatement:
			if currentGraph == nil {
				diags = append(diags, diagnostic.Diagnostic{
					Severity:  diagnostic.SeverityError,
					Message:   "no active graph selected",
					Statement: s.Text(),
					Line:      s.Pos().Line,
					Column:    s.Pos().Column,
					Code:      "CATALOG_GRAPH_MISSING",
				})
				continue
			}
			if !hasRolePermission(currentGraph, role, true) {
				diags = append(diags, diagnostic.Diagnostic{
					Severity:  diagnostic.SeverityError,
					Message:   fmt.Sprintf("role %q is not authorized to query graph %q", role, currentGraph.Name),
					Statement: s.Text(),
					Line:      s.Pos().Line,
					Column:    s.Pos().Column,
					Code:      "AUTHORIZATION_DENIED",
				})
				continue
			}
			vars := extractVariables(s.Pattern)
			missing := unresolvedVariables(vars, s.Return)
			if len(missing) > 0 {
				diags = append(diags, diagnostic.Diagnostic{
					Severity:  diagnostic.SeverityError,
					Message:   fmt.Sprintf("RETURN references undefined variables: %s", strings.Join(missing, ", ")),
					Statement: s.Text(),
					Line:      s.Pos().Line,
					Column:    s.Pos().Column,
					Code:      "SEMANTIC_UNKNOWN_VARIABLE",
				})
				continue
			}
			checked = append(checked, CheckedStatement{
				Statement: s,
				Graph:     currentGraph,
				Schema:    currentSchema,
				Variables: vars,
			})
		case *parser.CommandStatement:
			if currentGraph == nil {
				diags = append(diags, diagnostic.Diagnostic{
					Severity:  diagnostic.SeverityError,
					Message:   "no active graph selected",
					Statement: s.Text(),
					Line:      s.Pos().Line,
					Column:    s.Pos().Column,
					Code:      "CATALOG_GRAPH_MISSING",
				})
				continue
			}
			requiresWrite := isWriteCommand(s.Keyword)
			if requiresWrite && !hasRolePermission(currentGraph, role, false) {
				diags = append(diags, diagnostic.Diagnostic{
					Severity:  diagnostic.SeverityError,
					Message:   fmt.Sprintf("role %q is not authorized to modify graph %q", role, currentGraph.Name),
					Statement: s.Text(),
					Line:      s.Pos().Line,
					Column:    s.Pos().Column,
					Code:      "AUTHORIZATION_DENIED",
				})
				continue
			}
			checked = append(checked, CheckedStatement{Statement: s, Graph: currentGraph, Schema: currentSchema})
		default:
			checked = append(checked, CheckedStatement{Statement: stmt, Graph: currentGraph, Schema: currentSchema})
		}
	}
	if len(diags) > 0 {
		sort.SliceStable(diags, func(i, j int) bool {
			if diags[i].Line == diags[j].Line {
				return diags[i].Column < diags[j].Column
			}
			return diags[i].Line < diags[j].Line
		})
		return nil, diagnostic.NewError("gql: semantic analysis failed", diags...)
	}
	return &Result{Statements: checked}, nil
}

func hasRolePermission(g *Graph, role string, read bool) bool {
	if g == nil {
		return false
	}
	if role == "" {
		return true
	}
	perm, ok := g.Roles[role]
	if !ok {
		return false
	}
	if read {
		return perm.CanRead
	}
	return perm.CanWrite
}

var varPattern = regexp.MustCompile(`(?i)([\(\[\{]\s*)([A-Za-z_][A-Za-z0-9_]*)`)

func extractVariables(pattern string) []string {
	if pattern == "" {
		return nil
	}
	matches := varPattern.FindAllStringSubmatch(pattern, -1)
	seen := make(map[string]struct{}, len(matches))
	var vars []string
	for _, m := range matches {
		if len(m) < 3 {
			continue
		}
		name := m[2]
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		vars = append(vars, name)
	}
	return vars
}

func unresolvedVariables(known []string, projections []string) []string {
	if len(projections) == 0 {
		return nil
	}
	knownSet := make(map[string]struct{}, len(known))
	for _, k := range known {
		knownSet[k] = struct{}{}
	}
	var missing []string
	for _, item := range projections {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" || trimmed == "*" {
			continue
		}
		aliasIdx := strings.IndexAny(trimmed, " ()")
		candidate := trimmed
		if aliasIdx > 0 {
			candidate = trimmed[:aliasIdx]
		}
		candidate = strings.Trim(candidate, "`\"")
		if _, ok := knownSet[candidate]; !ok {
			if !strings.Contains(candidate, "(") && !strings.Contains(candidate, ".") {
				missing = append(missing, candidate)
			}
		}
	}
	if len(missing) > 1 {
		sort.Strings(missing)
	}
	return missing
}

func isWriteCommand(keyword string) bool {
	switch strings.ToUpper(keyword) {
	case "INSERT", "UPDATE", "DELETE", "MERGE", "CREATE", "DROP", "ASSERT":
		return true
	default:
		return false
	}
}
