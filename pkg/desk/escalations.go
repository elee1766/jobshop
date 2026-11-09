package desk

import (
	"cmp"
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// Escalation sessions (#379): the chats the controller opened for work
// on hold (meta.escalation), listed with their ticket so Desk shows them
// as escalations and opens the session panel from a click.

// Escalation is one escalation chat with the ticket it is about.
type Escalation struct {
	Ticket string `json:"ticket"`
	Title  string `json:"title"`
	// Cause is meta.escalation.cause: run, pool, wait, guard or label.
	Cause string `json:"cause,omitempty"`
	// Why is the first line of the ticket body.
	Why     string `json:"why,omitempty"`
	Held    string `json:"held,omitempty"`
	Key     string `json:"key"`
	Session string `json:"session"`
	// Open is true while the ticket is not closed.
	Open       bool      `json:"open"`
	Resolution string    `json:"resolution,omitempty"`
	Since      time.Time `json:"since"`
	Closed     time.Time `json:"closed,omitzero"`
}

// maxClosedEscalations bounds the closed escalations listed.
const (
	maxOpenEscalations   = 200
	maxClosedEscalations = 50
)

func escalationOf(t domain.Ticket) (Escalation, bool) {
	var m struct {
		Escalation struct {
			Key     string `json:"key"`
			Session string `json:"session"`
			Cause   string `json:"cause"`
			Held    string `json:"held"`
		} `json:"escalation"`
	}
	if json.Unmarshal(t.Meta, &m) != nil || m.Escalation.Session == "" {
		return Escalation{}, false
	}
	why, _, _ := strings.Cut(strings.TrimSpace(t.Body), "\n")
	if len(why) > 200 {
		why = why[:200] + "..."
	}
	return Escalation{
		Ticket: t.ID, Title: t.Title, Cause: m.Escalation.Cause, Why: why, Held: m.Escalation.Held,
		Key: m.Escalation.Key, Session: m.Escalation.Session, Open: t.Status != domain.StatusClosed,
		Resolution: t.Resolution, Since: t.CreatedAt, Closed: t.ClosedAt,
	}, true
}

// escalations lists escalation chats: open ones oldest first, then the
// latest closed ones.
func (s *Server) escalations(r *http.Request) (any, error) {
	var out []Escalation
	for _, f := range []rpc.ListParams{
		{Label: NeedsHuman, Status: domain.StatusOpen, Limit: maxOpenEscalations},
		{Label: NeedsHuman, Status: domain.StatusInProgress, Limit: maxOpenEscalations},
		{Label: NeedsHuman, Status: domain.StatusClosed, Newest: true, Limit: maxClosedEscalations},
	} {
		var ts []domain.Ticket
		if err := s.ctl.Call(r.Context(), rpc.MethodTicketList, f, &ts); err != nil {
			return nil, err
		}
		for _, t := range ts {
			if e, ok := escalationOf(t); ok {
				out = append(out, e)
			}
		}
	}
	slices.SortStableFunc(out, func(a, b Escalation) int {
		if a.Open != b.Open {
			if a.Open {
				return -1
			}
			return 1
		}
		if a.Open {
			return a.Since.Compare(b.Since)
		}
		return cmp.Or(b.Closed.Compare(a.Closed), strings.Compare(a.Ticket, b.Ticket))
	})
	if out == nil {
		out = []Escalation{}
	}
	return out, nil
}
