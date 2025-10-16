package parser

import (
	"strings"
	"unicode"

	"github.com/cayleygraph/cayley/query/gql/diagnostic"
)

type Position struct {
	Offset int
	Line   int
	Column int
}

type Statement interface {
	Pos() Position
	Text() string
	statement()
}

type Script struct {
	Statements []Statement
}

type baseStatement struct {
	start Position
	text  string
}

func (b baseStatement) Pos() Position { return b.start }
func (b baseStatement) Text() string  { return b.text }
func (baseStatement) statement()      {}

type UseGraphStatement struct {
	baseStatement
	Graph string
}

type MatchStatement struct {
	baseStatement
	Pattern string
	Where   string
	Return  []string
}

type CommandStatement struct {
	baseStatement
	Keyword string
	Body    string
}

func ParseScript(input string) (*Script, error) {
	segments, err := splitStatements(input)
	if err != nil {
		return nil, err
	}
	if len(segments) == 0 {
		return &Script{}, nil
	}
	script := &Script{Statements: make([]Statement, 0, len(segments))}
	for _, seg := range segments {
		stmt, err := parseStatement(seg)
		if err != nil {
			return nil, err
		}
		script.Statements = append(script.Statements, stmt)
	}
	return script, nil
}

type segment struct {
	text   string
	start  int
	line   int
	column int
}

func splitStatements(input string) ([]segment, error) {
	var (
		segments  []segment
		buf       strings.Builder
		quote     rune
		escape    bool
		line      = 1
		column    = 1
		startOff  = -1
		startLine = 1
		startCol  = 1
		stack     []rune
	)
	for i, r := range input {
		if startOff == -1 {
			if unicode.IsSpace(r) {
				if r == '\n' {
					line++
					column = 1
				} else {
					column++
				}
				continue
			}
			startOff = i
			startLine = line
			startCol = column
		}
		buf.WriteRune(r)
		if quote != 0 {
			if escape {
				escape = false
			} else if r == '\\' {
				escape = true
			} else if r == quote {
				quote = 0
			}
		} else {
			switch r {
			case '\'', '"':
				quote = r
			case '(', '[', '{':
				stack = append(stack, r)
			case ')', ']', '}':
				if len(stack) > 0 {
					top := stack[len(stack)-1]
					if matchesDelimiter(top, r) {
						stack = stack[:len(stack)-1]
					}
				}
			case ';':
				if len(stack) == 0 {
					text := strings.TrimSpace(buf.String()[:buf.Len()-1])
					if text != "" {
						segments = append(segments, segment{
							text:   text,
							start:  startOff,
							line:   startLine,
							column: startCol,
						})
					}
					buf.Reset()
					startOff = -1
					quote = 0
					stack = stack[:0]
				}
			}
		}
		if r == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	if quote != 0 {
		return nil, diagnostic.NewError("gql: unterminated string literal", diagnostic.Diagnostic{
			Severity:  diagnostic.SeverityError,
			Message:   "unterminated string literal",
			Statement: "",
			Line:      line,
			Column:    column,
		})
	}
	if buf.Len() > 0 {
		text := strings.TrimSpace(buf.String())
		if text != "" {
			segments = append(segments, segment{
				text:   text,
				start:  startOff,
				line:   startLine,
				column: startCol,
			})
		}
	}
	return segments, nil
}

func matchesDelimiter(open, close rune) bool {
	switch open {
	case '(':
		return close == ')'
	case '[':
		return close == ']'
	case '{':
		return close == '}'
	default:
		return false
	}
}

func parseStatement(seg segment) (Statement, error) {
	leading := firstWord(seg.text)
	if leading == "" {
		return nil, diagnostic.NewError("gql: empty statement", diagnostic.Diagnostic{
			Severity:  diagnostic.SeverityError,
			Message:   "statement is empty",
			Statement: "",
			Line:      seg.line,
			Column:    seg.column,
		})
	}
	keyword := strings.ToUpper(leading)
	switch keyword {
	case "USE":
		return parseUse(seg)
	case "MATCH":
		return parseMatch(seg)
	default:
		return parseCommand(keyword, seg)
	}
}

func parseUse(seg segment) (Statement, error) {
	rest := strings.TrimSpace(seg.text[len(firstWord(seg.text)):])
	if rest == "" {
		return nil, diagnostic.NewError("gql: invalid USE statement", diagnostic.Diagnostic{
			Severity:  diagnostic.SeverityError,
			Message:   "USE statement must specify a graph name",
			Statement: "USE",
			Line:      seg.line,
			Column:    seg.column,
		})
	}
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return nil, diagnostic.NewError("gql: invalid USE statement", diagnostic.Diagnostic{
			Severity:  diagnostic.SeverityError,
			Message:   "USE statement must specify a graph name",
			Statement: seg.text,
			Line:      seg.line,
			Column:    seg.column,
		})
	}
	nameIdx := 0
	if len(fields) > 1 && strings.EqualFold(fields[0], "GRAPH") {
		nameIdx = 1
	}
	if nameIdx >= len(fields) {
		return nil, diagnostic.NewError("gql: invalid USE statement", diagnostic.Diagnostic{
			Severity:  diagnostic.SeverityError,
			Message:   "USE statement must specify a graph name",
			Statement: seg.text,
			Line:      seg.line,
			Column:    seg.column,
		})
	}
	graph := normalizeIdentifier(fields[nameIdx])
	if graph == "" {
		return nil, diagnostic.NewError("gql: invalid USE statement", diagnostic.Diagnostic{
			Severity:  diagnostic.SeverityError,
			Message:   "graph name must be a quoted string or identifier",
			Statement: seg.text,
			Line:      seg.line,
			Column:    seg.column,
		})
	}
	return &UseGraphStatement{
		baseStatement: baseStatement{start: Position{Offset: seg.start, Line: seg.line, Column: seg.column}, text: seg.text},
		Graph:         graph,
	}, nil
}

func parseMatch(seg segment) (Statement, error) {
	body := strings.TrimSpace(seg.text[len(firstWord(seg.text)):])
	if body == "" {
		return nil, diagnostic.NewError("gql: invalid MATCH statement", diagnostic.Diagnostic{
			Severity:  diagnostic.SeverityError,
			Message:   "MATCH requires a pattern",
			Statement: seg.text,
			Line:      seg.line,
			Column:    seg.column,
		})
	}
	upper := strings.ToUpper(body)
	returnIdx := strings.Index(upper, " RETURN ")
	if returnIdx == -1 {
		return nil, diagnostic.NewError("gql: invalid MATCH statement", diagnostic.Diagnostic{
			Severity:  diagnostic.SeverityError,
			Message:   "RETURN clause is required",
			Statement: seg.text,
			Line:      seg.line,
			Column:    seg.column,
		})
	}
	whereIdx := strings.Index(upper[:returnIdx], " WHERE ")
	var (
		pattern string
		where   string
	)
	if whereIdx != -1 {
		pattern = strings.TrimSpace(body[:whereIdx])
		where = strings.TrimSpace(body[whereIdx+len(" WHERE ") : returnIdx])
	} else {
		pattern = strings.TrimSpace(body[:returnIdx])
	}
	if pattern == "" {
		return nil, diagnostic.NewError("gql: invalid MATCH statement", diagnostic.Diagnostic{
			Severity:  diagnostic.SeverityError,
			Message:   "MATCH pattern cannot be empty",
			Statement: seg.text,
			Line:      seg.line,
			Column:    seg.column,
		})
	}
	projection := strings.TrimSpace(body[returnIdx+len(" RETURN "):])
	if projection == "" {
		return nil, diagnostic.NewError("gql: invalid MATCH statement", diagnostic.Diagnostic{
			Severity:  diagnostic.SeverityError,
			Message:   "RETURN clause cannot be empty",
			Statement: seg.text,
			Line:      seg.line,
			Column:    seg.column,
		})
	}
	items := splitProjection(projection)
	if len(items) == 0 {
		return nil, diagnostic.NewError("gql: invalid MATCH statement", diagnostic.Diagnostic{
			Severity:  diagnostic.SeverityError,
			Message:   "RETURN clause must project at least one item",
			Statement: seg.text,
			Line:      seg.line,
			Column:    seg.column,
		})
	}
	return &MatchStatement{
		baseStatement: baseStatement{start: Position{Offset: seg.start, Line: seg.line, Column: seg.column}, text: seg.text},
		Pattern:       pattern,
		Where:         where,
		Return:        items,
	}, nil
}

func parseCommand(keyword string, seg segment) (Statement, error) {
	body := strings.TrimSpace(seg.text[len(firstWord(seg.text)):])
	return &CommandStatement{
		baseStatement: baseStatement{start: Position{Offset: seg.start, Line: seg.line, Column: seg.column}, text: seg.text},
		Keyword:       keyword,
		Body:          body,
	}, nil
}

func firstWord(s string) string {
	s = strings.TrimLeftFunc(s, unicode.IsSpace)
	for i, r := range s {
		if unicode.IsSpace(r) {
			return s[:i]
		}
	}
	return s
}

func normalizeIdentifier(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if (strings.HasPrefix(raw, "\"") && strings.HasSuffix(raw, "\"")) ||
		(strings.HasPrefix(raw, "'") && strings.HasSuffix(raw, "'")) {
		return raw[1 : len(raw)-1]
	}
	for _, r := range raw {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' {
			return ""
		}
	}
	return raw
}

func splitProjection(projection string) []string {
	var (
		parts []string
		buf   strings.Builder
		depth int
		quote rune
	)
	for _, r := range projection {
		if quote != 0 {
			buf.WriteRune(r)
			if r == quote {
				quote = 0
			}
			continue
		}
		switch r {
		case '\'', '"':
			quote = r
			buf.WriteRune(r)
		case '(', '[', '{':
			depth++
			buf.WriteRune(r)
		case ')', ']', '}':
			if depth > 0 {
				depth--
			}
			buf.WriteRune(r)
		case ',':
			if depth == 0 {
				part := strings.TrimSpace(buf.String())
				if part != "" {
					parts = append(parts, part)
				}
				buf.Reset()
				continue
			}
			buf.WriteRune(r)
		default:
			buf.WriteRune(r)
		}
	}
	if buf.Len() > 0 {
		part := strings.TrimSpace(buf.String())
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
}
