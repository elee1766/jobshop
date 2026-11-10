// Package gateway is the A2A gateway (#547): it mirrors escalation
// tickets to another agent. The Matrix and Discord chat gateways were
// removed on 2026-11-06; Desk is the operator UI.
package gateway

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// EscalationLabel marks escalation tickets.
const EscalationLabel = "needs-human"

// MetaPack is the ticket meta pack the gateway writes.
const MetaPack = "gateway"

// Controller is the part of the admin API the gateway uses.
type Controller interface {
	Subscribe(ctx context.Context, p rpc.SessionSubscribeParams, fn func(rpc.SessionEvent)) (rpc.SessionSubscribeResult, error)
	Send(ctx context.Context, p rpc.SessionSendParams) (rpc.SessionSendResult, error)
	// List, Show and Update read and mark escalation tickets.
	List(ctx context.Context, p rpc.ListParams) ([]domain.Ticket, error)
	Show(ctx context.Context, id string) (rpc.ShowResult, error)
	Update(ctx context.Context, p rpc.UpdateParams) (domain.Ticket, error)
	// Events follows events.subscribe; conversations are driven by it.
	Events(ctx context.Context, p rpc.EventsSubscribeParams, fn func(rpc.Event)) (rpc.EventsSubscribeResult, error)
	// Comment records A2A failures on the ticket.
	Comment(ctx context.Context, p rpc.CommentParams) error
}

// Escalation is meta.gateway on an escalation ticket.
type Escalation struct {
	// Gateway, Room, Root, Key, Session and Final were written by the
	// removed chat gateways. Kept so a write here does not drop them from
	// existing tickets.
	Gateway string `json:"gateway,omitempty"`
	Room    string `json:"room,omitempty"`
	Root    string `json:"root,omitempty"`
	Key     string `json:"key,omitempty"`
	Session string `json:"session,omitempty"`
	Final   bool   `json:"final,omitempty"`
	// A2A is each A2A gateway's conversation on this ticket, by gateway
	// name (#547).
	A2A map[string]*A2AState `json:"a2a,omitempty"`
}

func ticketMeta(t domain.Ticket) map[string]json.RawMessage {
	m := map[string]json.RawMessage{}
	if len(t.Meta) > 0 {
		_ = json.Unmarshal(t.Meta, &m)
	}
	return m
}

// chatMeta is meta.escalation, the controller's escalation chat (#379).
type chatMeta struct {
	Key     string `json:"key"`
	Pool    string `json:"pool"`
	Session string `json:"session"`
}

func metaChat(t domain.Ticket) (chatMeta, bool) {
	var c chatMeta
	raw, ok := ticketMeta(t)["escalation"]
	if !ok || json.Unmarshal(raw, &c) != nil || c.Key == "" || c.Session == "" {
		return c, false
	}
	return c, true
}

// updateGatewayMeta reads the ticket, lets fn change its meta.gateway, and
// writes it conditional on the revision read, retrying on conflicts (the
// session claiming or commenting moves the revision). e holds what was
// read even when ok is false, so a writer keeps fields it does not own.
func updateGatewayMeta(ctx context.Context, ctl Controller, id string, fn func(t domain.Ticket, e *Escalation, ok bool) error) (domain.Ticket, error) {
	for attempt := 0; ; attempt++ {
		r, err := ctl.Show(ctx, id)
		if err != nil {
			return domain.Ticket{}, err
		}
		t := r.Ticket
		e, ok := metaEscalationAny(t)
		if err := fn(t, &e, ok); err != nil {
			return t, err
		}
		m := ticketMeta(t)
		m[MetaPack], _ = json.Marshal(e)
		mb, err := json.Marshal(m)
		if err != nil {
			return t, err
		}
		t, err = ctl.Update(ctx, rpc.UpdateParams{ID: id, Meta: mb, IfRevision: t.Revision})
		if errors.Is(err, domain.ErrConflict) && attempt < 5 {
			continue
		}
		return t, err
	}
}
