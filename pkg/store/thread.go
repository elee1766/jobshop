package store

import "context"

// Helpers over Thread and Changes (#579: helpers wrap the interface in
// package store; they are not Store methods).

// Edges returns the edges in either direction touching id.
func Edges(ctx context.Context, s Store, id string) ([]Edge, error) {
	th, err := s.Thread(ctx, id, Parts{Edges: true})
	return th.Edges, err
}

// Comments returns id's comments oldest first.
func Comments(ctx context.Context, s Store, id string) ([]Comment, error) {
	th, err := s.Thread(ctx, id, Parts{Comments: true})
	return th.Comments, err
}

// History returns id's audit log oldest first.
func History(ctx context.Context, s Store, id string) ([]HistoryEntry, error) {
	th, err := s.Thread(ctx, id, Parts{History: true})
	return th.History, err
}

// LastSeq is the newest history seq, 0 when there is none.
func LastSeq(ctx context.Context, s Store) (int64, error) {
	p, err := s.Changes(ctx, 0, -1)
	return p.Last, err
}

// LastClaimant is the actor of id's last claim, "" when none.
func LastClaimant(ctx context.Context, s Store, id string) string {
	hist, err := History(ctx, s, id)
	if err != nil {
		return ""
	}
	for i := len(hist) - 1; i >= 0; i-- {
		if hist[i].Op == "claim" {
			return hist[i].Actor
		}
	}
	return ""
}
