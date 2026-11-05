package desk

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Chat (#203): Desk sends to sessions with session.send, the same call
// the gateway makes, and reads replies from /api/session/<id>/stream
// (session.subscribe). Escalation panels send by meta.escalation.key, the
// controller's escalation chat (#379), in the pool it recorded.

// maxChatText bounds one message; the agent prompt has its own limits.
const maxChatText = 32 << 10

// ChatSend is the body of POST /api/chat/send. Name the session by ID
// (rw attach, refused unless its pool has attach: rw), or by Key in the
// chat pool (an escalation thread or a desk:<id> session).
type ChatSend struct {
	Session   string `json:"session,omitempty"`
	Key       string `json:"key,omitempty"`
	Text      string `json:"text"`
	Interrupt bool   `json:"interrupt,omitempty"`
}

// ChatAccess is GET /api/session/<id>/access: whether Desk can send to it.
type ChatAccess struct {
	Write  bool   `json:"write"`
	Reason string `json:"reason,omitempty"`
}

// ChatNew is the result of POST /api/chat/new.
type ChatNew struct {
	Key     string `json:"key"`
	Session string `json:"session"`
}

// errBadRequest marks client errors (400).
var errBadRequest = errors.New("bad request")

func (s *Server) decode(r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, maxChatText+4096)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return fmt.Errorf("%w: %w", errBadRequest, err)
	}
	return nil
}

func (s *Server) chatSend(r *http.Request) (any, error) {
	var b ChatSend
	if err := s.decode(r, &b); err != nil {
		return nil, err
	}
	if strings.TrimSpace(b.Text) == "" || len(b.Text) > maxChatText {
		return nil, fmt.Errorf("%w: text must be 1 to %d bytes", errBadRequest, maxChatText)
	}
	p := rpc.SessionSendParams{Text: b.Text, Interrupt: b.Interrupt}
	switch {
	case b.Session != "" && b.Key == "":
		// Attach to an arbitrary session: the controller enforces attach: rw.
		p.ID, p.Attach = b.Session, true
	case b.Key != "" && b.Session == "":
		pool, err := s.keyPool(r, b.Key)
		if err != nil {
			return nil, err
		}
		p.Pool, p.Key = pool, b.Key
	default:
		return nil, fmt.Errorf("%w: give session or key", errBadRequest)
	}
	var res rpc.SessionSendResult
	if err := s.ctl.Call(r.Context(), rpc.MethodSessionSend, p, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// chatCancel is POST /api/chat/cancel {session}: Stop. It cancels the
// running turn through session.cancel (#535) and sends nothing.
func (s *Server) chatCancel(r *http.Request) (any, error) {
	var b struct {
		Session string `json:"session"`
	}
	if err := s.decode(r, &b); err != nil {
		return nil, err
	}
	if b.Session == "" {
		return nil, fmt.Errorf("%w: session is required", errBadRequest)
	}
	var res rpc.SessionCancelResult
	if err := s.ctl.Call(r.Context(), rpc.MethodSessionCancel, rpc.IDParams{ID: b.Session}, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// escalationKeyPrefix keys the controller's escalation chats (#379).
const escalationKeyPrefix = "escalation:"

// keyPool is the pool a keyed send goes to: an escalation:<ticket> key
// goes to the chat pool recorded on the ticket, others to Desk's chat
// pool.
func (s *Server) keyPool(r *http.Request, key string) (string, error) {
	if id, ok := strings.CutPrefix(key, escalationKeyPrefix); ok {
		var res rpc.ShowResult
		if err := s.ctl.Call(r.Context(), rpc.MethodTicketShow, rpc.IDParams{ID: id}, &res); err != nil {
			return "", err
		}
		var m struct {
			Escalation struct {
				Key  string `json:"key"`
				Pool string `json:"pool"`
			} `json:"escalation"`
		}
		_ = json.Unmarshal(res.Ticket.Meta, &m)
		if m.Escalation.Key != key || m.Escalation.Pool == "" {
			return "", fmt.Errorf("%w: %s has no escalation chat", errBadRequest, id)
		}
		return m.Escalation.Pool, nil
	}
	if s.cfg.ChatPool == "" {
		return "", fmt.Errorf("%w: no chat pool configured (--chat-pool)", errBadRequest)
	}
	return s.cfg.ChatPool, nil
}

// chatNew starts a fresh operator session keyed desk:<id> with a first
// message, so the session exists (and streams) from the first click.
func (s *Server) chatNew(r *http.Request) (any, error) {
	var b struct {
		Text string `json:"text"`
	}
	if err := s.decode(r, &b); err != nil {
		return nil, err
	}
	if strings.TrimSpace(b.Text) == "" || len(b.Text) > maxChatText {
		return nil, fmt.Errorf("%w: text must be 1 to %d bytes", errBadRequest, maxChatText)
	}
	if s.cfg.ChatPool == "" {
		return nil, fmt.Errorf("%w: no chat pool configured (--chat-pool)", errBadRequest)
	}
	key := "desk:" + strings.ToLower(rand.Text()[:12])
	var res rpc.SessionSendResult
	p := rpc.SessionSendParams{Pool: s.cfg.ChatPool, Key: key, Text: b.Text}
	if err := s.ctl.Call(r.Context(), rpc.MethodSessionSend, p, &res); err != nil {
		return nil, err
	}
	return ChatNew{Key: key, Session: res.Session}, nil
}

// chatAccess probes rw attach with an empty attach send, which the
// controller checks without delivering anything.
func (s *Server) chatAccess(r *http.Request) (any, error) {
	p := rpc.SessionSendParams{ID: r.PathValue("id"), Attach: true}
	var res rpc.SessionSendResult
	if err := s.ctl.Call(r.Context(), rpc.MethodSessionSend, p, &res); err != nil {
		if errors.Is(err, store.ErrInvalid) || errors.Is(err, store.ErrClosed) || errors.Is(err, store.ErrNotFound) {
			return ChatAccess{Reason: err.Error()}, nil
		}
		return nil, err
	}
	return ChatAccess{Write: true}, nil
}

// ChatSession is one open session of the chat pool, for GET /api/chat.
type ChatSession struct {
	ID    string `json:"id"`
	Key   string `json:"key"`
	Title string `json:"title,omitempty"`
	Alive bool   `json:"alive"`
}

func (s *Server) chatList(r *http.Request) (any, error) {
	var list []rpc.SessionInfo
	if err := s.ctl.Call(r.Context(), rpc.MethodSessionList, rpc.SessionListParams{}, &list); err != nil {
		return nil, err
	}
	out := []ChatSession{}
	for _, si := range list {
		if si.Ticket == nil {
			continue
		}
		f, _ := store.DecodeFields[store.SessionFields](*si.Ticket)
		// Escalation chats live in the pack's chat pool, which may not be
		// Desk's (#379).
		if si.Ticket.Pool != s.cfg.ChatPool && !strings.HasPrefix(f.Key, escalationKeyPrefix) {
			continue
		}
		out = append(out, ChatSession{ID: si.ID, Key: f.Key, Title: si.Ticket.Title, Alive: si.Alive})
	}
	return map[string]any{"pool": s.cfg.ChatPool, "sessions": out}, nil
}

// sameOrigin rejects a cross-site write whose browser still sent the
// custom header (it cannot without CORS, which Desk never grants); it
// is a second check behind CSRFHeader.
func (s *Server) sameOrigin(r *http.Request) bool {
	o := r.Header.Get("Origin")
	if o == "" {
		return true
	}
	u, err := url.Parse(o)
	return err == nil && u.Host == r.Host
}
