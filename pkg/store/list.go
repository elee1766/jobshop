package store

import (
	"context"
	"fmt"
	"strings"
)

// Helpers over List for common shapes (#579: helpers wrap the interface
// in package store; they are not Store methods).

// OpenStatuses are the statuses of work not yet closed.
var OpenStatuses = []Status{StatusOpen, StatusInProgress}

// QueryTerms bounds text search and escapes literal terms for SQL LIKE with ESCAPE '!'.
// A qualified reference's # separates its repository and number, also matching metadata.
func QueryTerms(query string) ([]string, error) {
	terms := strings.Fields(strings.ReplaceAll(query, "#", " "))
	if len(query) > 512 || len(terms) > 16 {
		return nil, fmt.Errorf("%w: query exceeds 512 bytes or 16 terms", ErrInvalid)
	}
	escape := strings.NewReplacer("!", "!!", "%", "!%", "_", "!_")
	for i, term := range terms {
		terms[i] = "%" + escape.Replace(strings.ToLower(term)) + "%"
	}
	return terms, nil
}

// Children returns id's children in fields.index order (a run's steps).
func Children(ctx context.Context, s Store, id string) ([]Ticket, error) {
	return s.List(ctx, Filter{ParentID: id, Order: OrderIndex, Limit: NoLimit})
}

// FindOne returns the first ticket matching f, or ErrNotFound.
func FindOne(ctx context.Context, s Store, f Filter) (Ticket, error) {
	f.Limit = 1
	ts, err := s.List(ctx, f)
	if err != nil {
		return Ticket{}, err
	}
	if len(ts) == 0 {
		return Ticket{}, fmt.Errorf("%w: no %s ticket with key %q", ErrNotFound, f.Kind, f.Key)
	}
	return ts[0], nil
}

// OpenRun returns the open run of workflow with key, or ErrNotFound.
// The controller serializes starts, so there is at most one.
func OpenRun(ctx context.Context, s Store, workflow, key string) (Ticket, error) {
	if workflow == "" || key == "" {
		// An empty Filter field matches all; no run has an empty key.
		return Ticket{}, fmt.Errorf("%w: open run %s key %q", ErrNotFound, workflow, key)
	}
	return FindOne(ctx, s, Filter{Kind: KindRun, Statuses: OpenStatuses, Workflow: workflow, Key: key, Order: OrderOldest})
}

// OpenBatch returns the open batch of project with key, or ErrNotFound.
func OpenBatch(ctx context.Context, s Store, project, key string) (Ticket, error) {
	if key == "" {
		return Ticket{}, fmt.Errorf("%w: open batch %s key %q", ErrNotFound, project, key)
	}
	return FindOne(ctx, s, Filter{Kind: KindBatch, Statuses: OpenStatuses, Project: project, Key: key, Order: OrderOldest})
}
