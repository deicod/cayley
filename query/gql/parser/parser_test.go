package parser_test

import (
	"errors"
	"testing"

	"github.com/cayleygraph/cayley/query/gql/diagnostic"
	"github.com/cayleygraph/cayley/query/gql/parser"
)

func TestParseMultiStatement(t *testing.T) {
	input := "USE GRAPH main; MATCH (n)-[e]->(m) RETURN n, m, e; INSERT VERTEX person;"
	script, err := parser.ParseScript(input)
	if err != nil {
		t.Fatalf("ParseScript returned error: %v", err)
	}
	if script == nil {
		t.Fatalf("expected script, got nil")
	}
	if len(script.Statements) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(script.Statements))
	}
	if _, ok := script.Statements[0].(*parser.UseGraphStatement); !ok {
		t.Fatalf("expected first statement to be UseGraphStatement, got %T", script.Statements[0])
	}
	match, ok := script.Statements[1].(*parser.MatchStatement)
	if !ok {
		t.Fatalf("expected second statement to be MatchStatement, got %T", script.Statements[1])
	}
	if len(match.Return) != 3 {
		t.Fatalf("expected 3 return items, got %d", len(match.Return))
	}
	if _, ok := script.Statements[2].(*parser.CommandStatement); !ok {
		t.Fatalf("expected third statement to be CommandStatement, got %T", script.Statements[2])
	}
}

func TestParseMatchMissingReturn(t *testing.T) {
	input := "MATCH (n) WHERE n.age > 30"
	_, err := parser.ParseScript(input)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	var derr *diagnostic.Error
	if !errors.As(err, &derr) {
		t.Fatalf("expected diagnostic error, got %T", err)
	}
	if len(derr.DiagnosticsList()) == 0 {
		t.Fatalf("expected diagnostics, got none")
	}
}

func TestParseSplitWithQuotedSemicolon(t *testing.T) {
	input := "MATCH (n) RETURN n.name, '\\';'"
	script, err := parser.ParseScript(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(script.Statements) != 1 {
		t.Fatalf("expected one statement, got %d", len(script.Statements))
	}
}

func TestParseSplitNestedSemicolons(t *testing.T) {
	input := "CALL { MATCH (n) RETURN n; }; MATCH (m) RETURN m"
	script, err := parser.ParseScript(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(script.Statements) != 2 {
		t.Fatalf("expected two statements, got %d", len(script.Statements))
	}
	if _, ok := script.Statements[0].(*parser.CommandStatement); !ok {
		t.Fatalf("expected first statement to be CommandStatement, got %T", script.Statements[0])
	}
	if _, ok := script.Statements[1].(*parser.MatchStatement); !ok {
		t.Fatalf("expected second statement to be MatchStatement, got %T", script.Statements[1])
	}
}
