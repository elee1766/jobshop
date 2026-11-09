package control

import (
	"context"

	"github.com/elee1766/jobshop/pkg/control/question"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/store"
)

// questionContext is the asker's context followed by the links and
// recent comments the controller fills in (JOB-101). The text is built by
// pkg/control/question; the comments are read here.
func (c *Controller) questionContext(ctx context.Context, sc scope, t domain.Ticket, asked string) string {
	cs, _ := store.Comments(ctx, c.store, t.ID)
	return question.Context(asked, t, question.Asker{Session: sc.session, Actor: sc.actor}, cs)
}
