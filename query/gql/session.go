package gql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/cayleygraph/cayley/graph"
	"github.com/cayleygraph/cayley/query"
	"github.com/cayleygraph/cayley/query/gql/diagnostic"
	"github.com/cayleygraph/cayley/query/gql/parser"
	"github.com/cayleygraph/cayley/query/gql/semantic"
)

const Name = "gql"

var ErrNotImplemented = errors.New("gql: execution is not yet implemented")

func init() {
	query.RegisterLanguage(query.Language{
		Name: Name,
		Session: func(qs graph.QuadStore) query.Session {
			return NewSession(qs)
		},
		HTTPError: httpError,
	})
}

type Session struct {
	qs           graph.QuadStore
	catalog      semantic.Catalog
	validator    *semantic.Validator
	role         string
	defaultGraph string
}

type SessionOption func(*Session)

const (
	defaultGraphName = "default"
	defaultRoleName  = "anonymous"
)

func NewSession(qs graph.QuadStore, opts ...SessionOption) *Session {
	s := &Session{qs: qs, role: defaultRoleName, defaultGraph: defaultGraphName}
	cat := semantic.NewInMemoryCatalog()
	cat.RegisterGraph(semantic.Graph{
		Name:          defaultGraphName,
		DefaultSchema: "",
		Roles: map[string]semantic.GraphRole{
			defaultRoleName: {
				Name:     defaultRoleName,
				CanRead:  true,
				CanWrite: false,
			},
		},
	})
	cat.SetDefaultGraph(defaultGraphName)
	s.catalog = cat
	for _, opt := range opts {
		opt(s)
	}
	if s.catalog == nil {
		s.catalog = cat
	}
	if s.role == "" {
		s.role = defaultRoleName
	}
	if s.validator == nil {
		s.validator = semantic.NewValidator(s.catalog)
	}
	return s
}

func (s *Session) Execute(ctx context.Context, input string, opt query.Options) (query.Iterator, error) {
	switch opt.Collation {
	case query.Raw, query.REPL, query.JSON, query.JSONLD:
		// supported collations
	default:
		if opt.Collation != 0 {
			return nil, &query.ErrUnsupportedCollation{Collation: opt.Collation}
		}
	}

	if opt.Limit < 0 {
		return nil, fmt.Errorf("gql: limit must be non-negative, got %d", opt.Limit)
	}

	if strings.TrimSpace(input) == "" {
		return nil, query.ErrParseMore
	}

	script, err := parser.ParseScript(input)
	if err != nil {
		return nil, err
	}

	if s.validator == nil {
		s.validator = semantic.NewValidator(s.catalog)
	}

	_, err = s.validator.Validate(ctx, script, semantic.Options{
		Role:         s.role,
		DefaultGraph: s.defaultGraph,
	})
	if err != nil {
		return nil, err
	}

	return nil, ErrNotImplemented
}

func httpError(w query.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	enc := json.NewEncoder(w)
	if derr, ok := diagnostic.As(err); ok {
		_ = enc.Encode(derr)
		return
	}
	_ = enc.Encode(struct {
		Error string `json:"error"`
	}{Error: err.Error()})
}

func WithCatalog(cat semantic.Catalog) SessionOption {
	return func(s *Session) {
		s.catalog = cat
		s.validator = semantic.NewValidator(cat)
	}
}

func WithRole(role string) SessionOption {
	return func(s *Session) {
		s.role = role
	}
}

func WithDefaultGraph(name string) SessionOption {
	return func(s *Session) {
		s.defaultGraph = name
	}
}
