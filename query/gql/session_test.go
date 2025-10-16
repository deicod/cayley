package gql_test

import (
	"context"
	"errors"
	"testing"

	"github.com/cayleygraph/cayley/query"
	"github.com/cayleygraph/cayley/query/gql"
	"github.com/cayleygraph/cayley/query/gql/diagnostic"
	"github.com/cayleygraph/cayley/query/gql/semantic"
)

func TestSessionRequiresInput(t *testing.T) {
	ses := gql.NewSession(nil)
	_, err := ses.Execute(context.Background(), "   ", query.Options{})
	if err != query.ErrParseMore {
		t.Fatalf("expected ErrParseMore, got %v", err)
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
	if err != gql.ErrNotImplemented {
		t.Fatalf("expected ErrNotImplemented, got %v", err)
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
