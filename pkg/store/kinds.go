package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// Kind is an engine-defined ticket kind. Packs cannot add kinds.
type Kind string

// Engine kinds.
const (
	KindTask    Kind = "task"
	KindSession Kind = "session"
	KindMessage Kind = "message"
)

// ClaimableKinds are the kinds ClaimNext and Ready consider by default.
// Messages are consumed through the inbox, sessions are never claimed.
var ClaimableKinds = []Kind{KindTask}

// Fields is implemented by each kind's typed field struct.
type Fields interface {
	Validate() error
}

// TaskFields has no engine fields yet.
type TaskFields struct{}

// Validate implements Fields.
func (TaskFields) Validate() error { return nil }

// SessionFields is the engine state of a session ticket.
type SessionFields struct {
	Key        string `json:"key,omitempty"` // e.g. matrix:<room>; unique while open
	Agent      string `json:"agent"`
	ACPSession string `json:"acp_session,omitempty"`
	Current    string `json:"current,omitempty"` // ticket id being worked
}

// Validate implements Fields.
func (f SessionFields) Validate() error {
	if f.Agent == "" {
		return fmt.Errorf("%w: session: agent is required", ErrInvalid)
	}
	return nil
}

// MessageFields addresses a message ticket.
type MessageFields struct {
	To        string `json:"to"` // session id, "pool:<name>", or session key
	From      string `json:"from"`
	Interrupt bool   `json:"interrupt,omitempty"`
}

// Validate implements Fields.
func (f MessageFields) Validate() error {
	if f.To == "" {
		return fmt.Errorf("%w: message: to is required", ErrInvalid)
	}
	if name, ok := strings.CutPrefix(f.To, "pool:"); ok && name == "" {
		return fmt.Errorf("%w: message: empty pool name", ErrInvalid)
	}
	return nil
}

func (k Kind) newFields() (Fields, bool) {
	switch k {
	case KindTask:
		return &TaskFields{}, true
	case KindSession:
		return &SessionFields{}, true
	case KindMessage:
		return &MessageFields{}, true
	}
	return nil, false
}

// Valid reports whether k is an engine kind.
func (k Kind) Valid() bool { _, ok := k.newFields(); return ok }

// NormalizeFields decodes raw strictly into the kind's struct, validates it
// and returns canonical JSON. Empty input means the zero struct.
func NormalizeFields(k Kind, raw json.RawMessage) (json.RawMessage, error) {
	f, ok := k.newFields()
	if !ok {
		return nil, fmt.Errorf("%w: unknown kind %q", ErrInvalid, k)
	}
	if len(bytes.TrimSpace(raw)) > 0 {
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.DisallowUnknownFields()
		if err := dec.Decode(f); err != nil {
			return nil, fmt.Errorf("%w: %s fields: %w", ErrInvalid, k, err)
		}
	}
	if err := f.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(f)
}

// DecodeFields returns the typed fields of t. T must match t.Kind.
func DecodeFields[T Fields](t Ticket) (T, error) {
	var f T
	if len(t.Fields) == 0 {
		return f, nil
	}
	err := json.Unmarshal(t.Fields, &f)
	return f, err
}

// NormalizeMeta checks that meta is an object whose values are objects, one
// per pack. Which pack names are allowed is the controller's (config) concern.
func NormalizeMeta(raw json.RawMessage) (json.RawMessage, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return json.RawMessage("{}"), nil
	}
	var m map[string]map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("%w: meta must be {\"<pack>\": {...}}: %w", ErrInvalid, err)
	}
	for name, v := range m {
		if name == "" || v == nil {
			return nil, fmt.Errorf("%w: meta: pack %q must map to an object", ErrInvalid, name)
		}
	}
	return json.Marshal(m)
}
