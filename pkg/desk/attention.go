package desk

import (
	"encoding/json"
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
// asks the controller for its chat (attention.chat, #606): the escalation
// chat when the ticket has one, else the chat-pool session keyed
// attention:<kind>:<id>, which the controller usually started when the
// item arrived. <root>/desk/attention.json holds acknowledged items.

// AttentionChat is POST /api/attention/chat's body and result.
type AttentionChat struct {
	Kind    string `json:"kind"`
	ID      string `json:"id"`
	Key     string `json:"key,omitempty"`
	Session string `json:"session,omitempty"`
	// Pool is where the chat runs; sends by Key go there.
	Pool string `json:"pool,omitempty"`
	// Started is set when this call started the chat; a pre-started chat
	// is only attached to.
	Started bool `json:"started,omitempty"`
}

// attentionState is <root>/desk/attention.json.
type attentionState struct {
	// Chats mapped kind:id to a chat key before #606 moved attention
	// chats to the controller. Kept so an old file round-trips.
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
	if err == nil {
		// An acknowledged item's chat is done (#606).
		s.closeAttentionChat(r, "attention:"+k)
	}
	return map[string]bool{"ok": err == nil}, err
}

// closeAttentionChat closes the open session keyed key, if any.
func (s *Server) closeAttentionChat(r *http.Request, key string) {
	var list []rpc.SessionInfo
	if err := s.ctl.Call(r.Context(), rpc.MethodSessionList, rpc.SessionListParams{}, &list); err != nil {
		return
	}
	for _, si := range list {
		if si.Ticket == nil || si.Ticket.Status == store.StatusClosed {
			continue
		}
		if f, _ := store.DecodeFields[store.SessionFields](*si.Ticket); f.Key == key {
			var out store.Ticket
			if err := s.ctl.Call(r.Context(), rpc.MethodTicketClose, rpc.CloseParams{ID: si.ID, Resolution: "acknowledged"}, &out); err != nil {
				s.log.WarnContext(r.Context(), "desk: close acknowledged chat", "session", si.ID, "err", err)
			}
		}
	}
}

// attentionChat is POST /api/attention/chat {kind, id, title, why}: the
// session that resolves the item. The controller owns it (attention.chat,
// #606): usually the chat it pre-started when the item arrived, so this
// only attaches and the pane replays the journal with the first reply.
func (s *Server) attentionChat(r *http.Request) (any, error) {
	var b rpc.AttentionChatParams
	if err := s.decode(r, &b); err != nil {
		return nil, err
	}
	if b.Kind == "" || b.ID == "" || strings.ContainsAny(b.Kind+b.ID, " \n") {
		return nil, fmt.Errorf("%w: kind and id are required", errBadRequest)
	}
	var res rpc.AttentionChatResult
	if err := s.ctl.Call(r.Context(), rpc.MethodAttentionChat, b, &res); err != nil {
		return nil, err
	}
	return AttentionChat{Kind: res.Kind, ID: res.ID, Key: res.Key, Session: res.Session, Pool: res.Pool, Started: res.Started}, nil
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
