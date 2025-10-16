package semantic

import (
	"context"
	"errors"
	"sync"
)

type InMemoryCatalog struct {
	mu           sync.RWMutex
	graphs       map[string]*Graph
	defaultGraph string
}

func NewInMemoryCatalog() *InMemoryCatalog {
	return &InMemoryCatalog{graphs: make(map[string]*Graph)}
}

func (c *InMemoryCatalog) SetDefaultGraph(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.defaultGraph = name
}

func (c *InMemoryCatalog) RegisterGraph(g Graph) {
	c.mu.Lock()
	defer c.mu.Unlock()
	copy := g
	if copy.Schemas == nil {
		copy.Schemas = make(map[string]Schema)
	}
	if copy.Roles == nil {
		copy.Roles = make(map[string]GraphRole)
	}
	c.graphs[g.Name] = &copy
}

func (c *InMemoryCatalog) LookupGraph(ctx context.Context, name string) (*Graph, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if ctx != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}
	g, ok := c.graphs[name]
	if !ok {
		return nil, errors.New("graph not found")
	}
	return cloneGraph(g), nil
}

func (c *InMemoryCatalog) DefaultGraph(ctx context.Context) (*Graph, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if ctx != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}
	if c.defaultGraph == "" {
		for _, g := range c.graphs {
			return cloneGraph(g), nil
		}
		return nil, errors.New("no default graph configured")
	}
	g, ok := c.graphs[c.defaultGraph]
	if !ok {
		return nil, errors.New("default graph not registered")
	}
	return cloneGraph(g), nil
}

func cloneGraph(g *Graph) *Graph {
	if g == nil {
		return nil
	}
	cp := *g
	if g.Schemas != nil {
		cp.Schemas = make(map[string]Schema, len(g.Schemas))
		for k, v := range g.Schemas {
			cp.Schemas[k] = v
		}
	}
	if g.Roles != nil {
		cp.Roles = make(map[string]GraphRole, len(g.Roles))
		for k, v := range g.Roles {
			cp.Roles[k] = v
		}
	}
	return &cp
}
