package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/gfx-labs/jobshop/pkg/domain"
)

// AddLabel adds a label; an existing one is a no-op. A helper over
// Update (#579).
func AddLabel(ctx context.Context, s Store, actor, id, label string) error {
	_, err := s.Update(ctx, actor, id, Update{AddLabels: []string{label}})
	return err
}

// RemoveLabel removes a label; a missing one is a no-op. A helper over
// Update.
func RemoveLabel(ctx context.Context, s Store, actor, id, label string) error {
	_, err := s.Update(ctx, actor, id, Update{RemoveLabels: []string{label}})
	return err
}

// Reopen returns a closed ticket to open. A helper over Update.
func Reopen(ctx context.Context, s Store, actor, id string) (domain.Ticket, error) {
	return s.Update(ctx, actor, id, Update{Reopen: true})
}

// AddComment appends a comment and returns it. A helper over Update:
// the comment is read back as the newest one by author with this body.
func AddComment(ctx context.Context, s Store, author, id, body string) (domain.Comment, error) {
	if strings.TrimSpace(body) == "" {
		return domain.Comment{}, fmt.Errorf("%w: empty comment", domain.ErrInvalid)
	}
	if _, err := s.Update(ctx, author, id, Update{Comment: body}); err != nil {
		return domain.Comment{}, err
	}
	th, err := s.Thread(ctx, id, Parts{Comments: true})
	if err != nil {
		return domain.Comment{}, err
	}
	for i := len(th.Comments) - 1; i >= 0; i-- {
		if c := th.Comments[i]; c.Author == author && c.Body == body {
			return c, nil
		}
	}
	return domain.Comment{}, fmt.Errorf("%w: comment on %s not found after write", domain.ErrNotFound, id)
}
