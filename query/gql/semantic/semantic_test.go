package semantic_test

import (
	"context"
	"errors"
	"testing"

	"github.com/cayleygraph/cayley/query/gql/diagnostic"
	"github.com/cayleygraph/cayley/query/gql/parser"
	"github.com/cayleygraph/cayley/query/gql/semantic"
)

func setupCatalog() *semantic.InMemoryCatalog {
	cat := semantic.NewInMemoryCatalog()
	cat.RegisterGraph(semantic.Graph{
		Name:          "main",
		DefaultSchema: "public",
		Roles: map[string]semantic.GraphRole{
			"reader": {Name: "reader", CanRead: true, CanWrite: false},
			"writer": {Name: "writer", CanRead: true, CanWrite: true},
		},
	})
	cat.SetDefaultGraph("main")
	return cat
}

func TestValidateMatchRequiresGraph(t *testing.T) {
	cat := setupCatalog()
	validator := semantic.NewValidator(cat)
	script, err := parser.ParseScript("MATCH (n) RETURN n")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	_, err = validator.Validate(context.Background(), script, semantic.Options{Role: "reader"})
	if err != nil {
		t.Fatalf("expected validation success, got %v", err)
	}
}

func TestValidateUseUnknownGraph(t *testing.T) {
	cat := setupCatalog()
	validator := semantic.NewValidator(cat)
	script, err := parser.ParseScript("USE GRAPH missing; MATCH (n) RETURN n")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	_, err = validator.Validate(context.Background(), script, semantic.Options{Role: "reader"})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	var derr *diagnostic.Error
	if !errors.As(err, &derr) {
		t.Fatalf("expected diagnostic error, got %T", err)
	}
}

func TestValidateAuthorization(t *testing.T) {
	cat := setupCatalog()
	validator := semantic.NewValidator(cat)
	script, err := parser.ParseScript("MATCH (n) RETURN n; DELETE GRAPH main")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	_, err = validator.Validate(context.Background(), script, semantic.Options{Role: "reader"})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	var derr *diagnostic.Error
	if !errors.As(err, &derr) {
		t.Fatalf("expected diagnostic error, got %T", err)
	}
}

func TestValidateReturnVariables(t *testing.T) {
	cat := setupCatalog()
	validator := semantic.NewValidator(cat)
	script, err := parser.ParseScript("MATCH (n) RETURN n, m")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	_, err = validator.Validate(context.Background(), script, semantic.Options{Role: "reader"})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	var derr *diagnostic.Error
	if !errors.As(err, &derr) {
		t.Fatalf("expected diagnostic error, got %T", err)
	}
}
