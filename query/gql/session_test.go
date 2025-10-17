package gql_test

import (
	"context"
	"errors"
	"testing"

	"github.com/cayleygraph/cayley/query"
	"github.com/cayleygraph/cayley/query/gql"
	"github.com/cayleygraph/cayley/query/gql/diagnostic"
	"github.com/cayleygraph/cayley/query/gql/parser"
	"github.com/cayleygraph/cayley/query/gql/semantic"
)

func TestSessionRequiresInput(t *testing.T) {
	ses := gql.NewSession(nil)
	_, err := ses.Execute(context.Background(), "   ", query.Options{})
	if err != query.ErrParseMore {
		t.Fatalf("expected ErrParseMore, got %v", err)
	}
}

func TestLanguageRegistered(t *testing.T) {
	lang := query.GetLanguage(gql.Name)
	if lang == nil {
		t.Fatalf("expected language %q to be registered", gql.Name)
	}
	if lang.Session == nil {
		t.Fatalf("registered language %q is missing a session factory", gql.Name)
	}
}

func TestSessionParseDiagnostics(t *testing.T) {
	ses := gql.NewSession(nil)
	_, err := ses.Execute(context.Background(), "MATCH (n)", query.Options{})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	var derr *diagnostic.Error
	if !errors.As(err, &derr) {
		t.Fatalf("expected diagnostic error, got %T", err)
	}
}

func TestSessionSemanticValidation(t *testing.T) {
	cat := semantic.NewInMemoryCatalog()
	cat.RegisterGraph(semantic.Graph{
		Name: "main",
		Roles: map[string]semantic.GraphRole{
			"reader": {Name: "reader", CanRead: true, CanWrite: false},
		},
	})
	cat.SetDefaultGraph("main")
	ses := gql.NewSession(nil, gql.WithCatalog(cat), gql.WithRole("reader"), gql.WithDefaultGraph("main"))
	_, err := ses.Execute(context.Background(), "MATCH (n) RETURN n", query.Options{})
	var merr *gql.MilestoneError
	if !errors.As(err, &merr) {
		t.Fatalf("expected milestone error, got %v", err)
	}
	if merr.Capability != gql.CapabilityExecution {
		t.Fatalf("expected execution capability error, got %v", merr.Capability)
	}
}

func TestSessionMilestoneParsingDisabled(t *testing.T) {
	ses := gql.NewSession(nil, gql.WithMilestone(gql.Milestone1Readiness))
	_, err := ses.Execute(context.Background(), "MATCH (n) RETURN n", query.Options{})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	var merr *gql.MilestoneError
	if !errors.As(err, &merr) {
		t.Fatalf("expected milestone error, got %v", err)
	}
	if merr.Capability != gql.CapabilityParsing {
		t.Fatalf("expected parsing capability error, got %v", merr.Capability)
	}
}

func TestSessionSemanticAuthorizationError(t *testing.T) {
	cat := semantic.NewInMemoryCatalog()
	cat.RegisterGraph(semantic.Graph{
		Name: "main",
		Roles: map[string]semantic.GraphRole{
			"reader": {Name: "reader", CanRead: true, CanWrite: false},
		},
	})
	cat.SetDefaultGraph("main")
	ses := gql.NewSession(nil, gql.WithCatalog(cat), gql.WithRole("reader"), gql.WithDefaultGraph("main"))
	_, err := ses.Execute(context.Background(), "DELETE GRAPH main", query.Options{})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	var derr *diagnostic.Error
	if !errors.As(err, &derr) {
		t.Fatalf("expected diagnostic error, got %T", err)
	}
}

type captureValidator struct {
	inner *semantic.Validator
	last  semantic.Options
}

func (c *captureValidator) Validate(ctx context.Context, script *parser.Script, opt semantic.Options) (*semantic.Result, error) {
	c.last = opt
	if c.inner != nil {
		return c.inner.Validate(ctx, script, opt)
	}
	return &semantic.Result{}, nil
}

func TestSessionDefaultSchemaOption(t *testing.T) {
	cat := semantic.NewInMemoryCatalog()
	cat.RegisterGraph(semantic.Graph{
		Name:          "main",
		DefaultSchema: "public",
		Roles: map[string]semantic.GraphRole{
			"reader": {Name: "reader", CanRead: true, CanWrite: false},
		},
	})
	cat.SetDefaultGraph("main")

	cap := &captureValidator{inner: semantic.NewValidator(cat)}
	ses := gql.NewSession(nil,
		gql.WithCatalog(cat),
		gql.WithRole("reader"),
		gql.WithDefaultGraph("main"),
		gql.WithDefaultSchema("analytics"),
		gql.WithValidator(cap),
	)

	_, err := ses.Execute(context.Background(), "MATCH (n) RETURN n", query.Options{})
	var merr *gql.MilestoneError
	if !errors.As(err, &merr) {
		t.Fatalf("expected milestone error, got %v", err)
	}

	if cap.last.DefaultGraph != "main" {
		t.Fatalf("expected default graph to be %q, got %q", "main", cap.last.DefaultGraph)
	}
	if cap.last.DefaultSchema != "analytics" {
		t.Fatalf("expected default schema to be %q, got %q", "analytics", cap.last.DefaultSchema)
	}
	if cap.last.Role != "reader" {
		t.Fatalf("expected role to be %q, got %q", "reader", cap.last.Role)
	}
}
