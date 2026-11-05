package desk

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Attention opens the chat that resolves it (#535). A click on an item
// asks Desk for its chat: the controller's escalation chat when the
// ticket has one, else a chat-pool session keyed attention:<kind>:<id>,
// started with a context prompt. The key is kept on the ticket
// (meta.desk.attention_chat) or, for items that are not tickets, in
// <root>/desk/attention.json, which also holds acknowledged items.

// ticketKinds are attention kinds whose ID is a ticket.
var ticketKinds = []string{
	rpc.AttentionRun, rpc.AttentionTicket, rpc.AttentionQuestion, rpc.AttentionDecision,
	rpc.AttentionStale, rpc.AttentionWait,
}

// AttentionChat is POST /api/attention/chat's body and result.
type AttentionChat struct {
	Kind    string `json:"kind"`
	ID      string `json:"id"`
	Key     string `json:"key,omitempty"`
	Session string `json:"session,omitempty"`
	// Pool is where the chat runs; sends by Key go there.
	Pool string `json:"pool,omitempty"`
	// Started is set when this call started the chat.
	Started bool `json:"started,omitempty"`
}

// attentionState is <root>/desk/attention.json.
type attentionState struct {
	// Chats maps kind:id to the chat key, for items that are not tickets.
	Chats map[string]string `json:"chats,omitempty"`
	// Acked are kind:id keys acknowledged in Desk; hidden while listed.
	Acked []string `json:"acked,omitempty"`
}

type attentionFile struct {
	mu   sync.Mutex
	path string
}

func (f *attentionFile) load() attentionState {
	var st attentionState
	if f.path == "" {
		return st
	}
	b, err := os.ReadFile(f.path)
	if err == nil {
		_ = json.Unmarshal(b, &st)
	}
	return st
}

func (f *attentionFile) update(fn func(*attentionState)) error {
	if f.path == "" {
		return fmt.Errorf("%w: Desk has no root to keep attention state in", errBadRequest)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	st := f.load()
	fn(&st)
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(f.path), 0o700); err != nil {
		return err
	}
	tmp := f.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, f.path)
}

// dropAcked hides acknowledged items and forgets acks for items gone.
func (s *Server) dropAcked(out []AttentionItem) []AttentionItem {
	st := s.attn.load()
	if len(st.Acked) == 0 {
		return out
	}
	listed := map[string]bool{}
	for _, it := range out {
		listed[attentionKey(it)] = true
	}
	kept := slices.DeleteFunc(slices.Clone(st.Acked), func(k string) bool { return !listed[k] })
	if len(kept) != len(st.Acked) {
		_ = s.attn.update(func(st *attentionState) { st.Acked = kept })
	}
	return slices.DeleteFunc(out, func(it AttentionItem) bool { return slices.Contains(kept, attentionKey(it)) })
}

// attentionAck is POST /api/attention/ack {kind, id}: hide an item that
// has no other resolve action (rollback, config, ingest, resource).
func (s *Server) attentionAck(r *http.Request) (any, error) {
	var b AttentionChat
	if err := s.decode(r, &b); err != nil {
		return nil, err
	}
	if b.Kind == "" || b.ID == "" {
		return nil, fmt.Errorf("%w: kind and id are required", errBadRequest)
	}
	k := b.Kind + ":" + b.ID
	err := s.attn.update(func(st *attentionState) {
		if !slices.Contains(st.Acked, k) {
			st.Acked = append(st.Acked, k)
		}
	})
	return map[string]bool{"ok": err == nil}, err
}

// attentionChat is POST /api/attention/chat {kind, id, title}: the
// session that resolves the item, started when there is none.
func (s *Server) attentionChat(r *http.Request) (any, error) {
	var b struct {
		Kind  string `json:"kind"`
		ID    string `json:"id"`
		Title string `json:"title"`
		Why   string `json:"why"`
	}
	if err := s.decode(r, &b); err != nil {
		return nil, err
	}
	if b.Kind == "" || b.ID == "" || strings.ContainsAny(b.Kind+b.ID, " \n") {
		return nil, fmt.Errorf("%w: kind and id are required", errBadRequest)
	}
	ctx := r.Context()
	res := AttentionChat{Kind: b.Kind, ID: b.ID, Key: "attention:" + b.Kind + ":" + b.ID, Pool: s.cfg.ChatPool}
	isTicket := slices.Contains(ticketKinds, b.Kind)
	var t store.Ticket
	if isTicket {
		var sh rpc.ShowResult
		if err := s.ctl.Call(ctx, rpc.MethodTicketShow, rpc.IDParams{ID: b.ID}, &sh); err != nil {
			return nil, err
		}
		t = sh.Ticket
		// The controller's escalation chat answers for the ticket (#379).
		if e, ok := escalationOf(t); ok && b.Kind != rpc.AttentionQuestion {
			var m struct {
				Escalation struct {
					Pool string `json:"pool"`
				} `json:"escalation"`
			}
			_ = json.Unmarshal(t.Meta, &m)
			res.Key, res.Session, res.Pool = e.Key, e.Session, m.Escalation.Pool
			return res, nil
		}
		if k := deskMeta(t.Meta).AttentionChat; k != "" {
			res.Key = k
		}
	} else if k := s.attn.load().Chats[b.Kind+":"+b.ID]; k != "" {
		res.Key = k
	}
	if s.cfg.ChatPool == "" {
		return nil, fmt.Errorf("%w: no chat pool configured (--chat-pool)", errBadRequest)
	}
	if id := s.openChat(r, res.Key); id != "" {
		res.Session = id
		return res, nil
	}
	prompt := s.attentionPrompt(r, b.Kind, b.ID, b.Title, b.Why, t)
	var sent rpc.SessionSendResult
	p := rpc.SessionSendParams{Pool: s.cfg.ChatPool, Key: res.Key, Text: prompt}
	if err := s.ctl.Call(ctx, rpc.MethodSessionSend, p, &sent); err != nil {
		return nil, err
	}
	res.Session, res.Started = sent.Session, true
	if err := s.recordChat(r, isTicket, t, b.Kind+":"+b.ID, res.Key); err != nil {
		s.log.Warn("desk: record attention chat", "item", b.Kind+":"+b.ID, "err", err)
	}
	return res, nil
}

// openChat is the open session for key in the chat pool, or "".
func (s *Server) openChat(r *http.Request, key string) string {
	var list []rpc.SessionInfo
	if err := s.ctl.Call(r.Context(), rpc.MethodSessionList, rpc.SessionListParams{}, &list); err != nil {
		return ""
	}
	for _, si := range list {
		if si.Ticket == nil || si.Ticket.Status == store.StatusClosed {
			continue
		}
		if f, _ := store.DecodeFields[store.SessionFields](*si.Ticket); f.Key == key {
			return si.ID
		}
	}
	return ""
}

type deskMetaPack struct {
	AttentionChat string `json:"attention_chat,omitempty"`
}

func deskMeta(raw json.RawMessage) deskMetaPack {
	var m struct {
		Desk deskMetaPack `json:"desk"`
	}
	_ = json.Unmarshal(raw, &m)
	return m.Desk
}

// recordChat keeps the key so a second click finds the same chat. On a
// ticket it goes in meta.desk; ticket.update replaces meta, so the
// write is conditional on the revision read.
func (s *Server) recordChat(r *http.Request, isTicket bool, t store.Ticket, item, key string) error {
	if !isTicket {
		return s.attn.update(func(st *attentionState) {
			if st.Chats == nil {
				st.Chats = map[string]string{}
			}
			st.Chats[item] = key
		})
	}
	if deskMeta(t.Meta).AttentionChat == key {
		return nil
	}
	m := map[string]json.RawMessage{}
	if len(t.Meta) > 0 {
		if err := json.Unmarshal(t.Meta, &m); err != nil {
			return err
		}
	}
	d, err := json.Marshal(deskMetaPack{AttentionChat: key})
	if err != nil {
		return err
	}
	m["desk"] = d
	raw, err := json.Marshal(m)
	if err != nil {
		return err
	}
	var out store.Ticket
	err = s.ctl.Call(r.Context(), rpc.MethodTicketUpdate, rpc.UpdateParams{ID: t.ID, Meta: raw, IfRevision: t.Revision}, &out)
	if errors.Is(err, store.ErrConflict) {
		return nil // changed meanwhile; the key is deterministic, so the next click still finds the chat
	}
	return err
}

// attentionPrompt is the first message of a resolving chat: what the
// item is, its links, and for a run the run show output.
func (s *Server) attentionPrompt(r *http.Request, kind, id, title, why string, t store.Ticket) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Desk attention item: %s %s\n", kind, id)
	if t.ID != "" && title == "" {
		title = t.Title
	}
	if title != "" {
		fmt.Fprintf(&b, "Title: %s\n", title)
	}
	if why != "" && why != kind {
		fmt.Fprintf(&b, "Why: %s\n", why)
	}
	if t.ID != "" {
		var m struct {
			GitHub struct {
				Repo  string `json:"repo"`
				Issue any    `json:"issue"`
				PR    any    `json:"pr"`
			} `json:"github"`
		}
		_ = json.Unmarshal(t.Meta, &m)
		if m.GitHub.Repo != "" && m.GitHub.Issue != nil {
			fmt.Fprintf(&b, "Issue: https://github.com/%s/issues/%v\n", m.GitHub.Repo, m.GitHub.Issue)
		}
		if m.GitHub.Repo != "" && m.GitHub.PR != nil {
			fmt.Fprintf(&b, "PR: https://github.com/%s/pull/%v\n", m.GitHub.Repo, m.GitHub.PR)
		}
	}
	if kind == rpc.AttentionRun {
		var run json.RawMessage
		if err := s.ctl.Call(r.Context(), rpc.MethodRunShow, rpc.IDParams{ID: id}, &run); err == nil {
			out, _ := json.MarshalIndent(run, "", "  ")
			if len(out) > 16<<10 {
				out = append(out[:16<<10], "\n[cut]"...)
			}
			fmt.Fprintf(&b, "\n`jobshop run show %s`:\n```json\n%s\n```\n", id, out)
		}
	}
	b.WriteString("\nHelp the operator resolve this item. Look at it with the jobshop CLI, explain what blocks it, and propose the action. Do not act on it until the operator agrees.")
	return b.String()
}

// SessionEvents is GET /api/session/<id>/events?before=<seq>&limit=<n>:
// decoded events before a seq, for the chat's "Load older" (#535).
// First is the lowest seq read; 1 or less means the start was reached.
type SessionEvents struct {
	Events []rpc.SessionEvent `json:"events"`
	First  int64              `json:"first"`
}

const maxOlderFrames = 2000

func (s *Server) sessionEvents(r *http.Request) (any, error) {
	q := r.URL.Query()
	before, err := strconv.ParseInt(q.Get("before"), 10, 64)
	if err != nil || before <= 1 {
		return nil, fmt.Errorf("%w: before must be a seq > 1", errBadRequest)
	}
	limit := 400
	if v := q.Get("limit"); v != "" {
		if limit, err = strconv.Atoi(v); err != nil || limit <= 0 {
			return nil, fmt.Errorf("%w: limit must be > 0", errBadRequest)
		}
	}
	limit = min(limit, maxOlderFrames)
	since := max(before-1-int64(limit), 0)
	p := rpc.SessionJournalParams{ID: r.PathValue("id"), SinceSeq: since, Limit: int(before - 1 - since)}
	var page rpc.SessionJournalResult
	if err := s.ctl.Call(r.Context(), rpc.MethodSessionJournal, p, &page); err != nil {
		return nil, err
	}
	out := SessionEvents{Events: []rpc.SessionEvent{}, First: since + 1}
	var dec rpc.EventDecoder
	for _, f := range page.Frames {
		if f.Seq >= before {
			break
		}
		if ev, ok := dec.Decode(f.Seq, f.T, f.From, f.Msg); ok {
			ev.Session = p.ID
			out.Events = append(out.Events, ev)
		}
	}
	return out, nil
}
