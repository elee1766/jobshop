package store

import "context"

// Create inserts one ticket. A helper over Apply (#579), not a Store
// method.
func Create(ctx context.Context, s Store, actor string, nt NewTicket) (Ticket, error) {
	out, err := s.Apply(ctx, actor, Graph{Tickets: []NewTicket{nt}})
	if err != nil {
		return Ticket{}, err
	}
	return out[0], nil
}

// AddEdge records one relation. Blocking edges that would form a cycle
// fail with ErrCycle. A helper over Apply.
func AddEdge(ctx context.Context, s Store, actor string, e Edge) error {
	_, err := s.Apply(ctx, actor, Graph{Edges: []Edge{e}})
	return err
}

// RemoveEdge deletes one relation; a missing edge is ErrNotFound. A
// helper over Apply.
func RemoveEdge(ctx context.Context, s Store, actor string, e Edge) error {
	_, err := s.Apply(ctx, actor, Graph{RemoveEdges: []Edge{e}})
	return err
}
