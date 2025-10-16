package gql

import (
	"context"
	"errors"
	"fmt"

	"github.com/cayleygraph/cayley/graph"
	"github.com/cayleygraph/cayley/query"
)

const Name = "gql"

var ErrNotImplemented = errors.New("gql: execution is not yet implemented")

func init() {
	query.RegisterLanguage(query.Language{
		Name: Name,
		Session: func(qs graph.QuadStore) query.Session {
			return NewSession(qs)
		},
	})
}

type Session struct {
	qs graph.QuadStore
}

func NewSession(qs graph.QuadStore) *Session {
	return &Session{qs: qs}
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

	return nil, ErrNotImplemented
}
