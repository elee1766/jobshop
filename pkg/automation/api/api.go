// Package api defines the shared contract between the automation engine and its
// host capabilities. It does not depend on the engine or capability packages.
package api

import (
	"context"
	"net/http"
	"time"

	"go.starlark.net/starlark"
)

// Limits are host ceilings. Zero values select bounded defaults in the engine.
type Limits struct {
	Duration    time.Duration
	Steps       uint64
	OutputBytes int
	SourceBytes int
	Modules     int
	DataBytes   int
	Actions     int
	Calls       int
}

// Credential is host-owned and never made available as a script value.
type Credential struct {
	Origins []string
	Headers map[string]string
	Secrets []string
}

// JevOptions names a full Decisions endpoint, not a service base URL.
type JevOptions struct {
	Enabled  bool
	Endpoint string
	APIKey   string
	Model    string
}

// Options grants host capabilities. Shell runs as the controller OS user and is
// not an agent sandbox. HTTP permits private networks unless HTTPOrigins limits it.
type Options struct {
	Limits      Limits
	HTTP        *http.Client
	AllowHTTP   bool
	HTTPOrigins []string
	Auth        map[string]Credential
	AllowShell  bool
	ShellDir    string
	ShellEnv    []string
	Jev         JevOptions
	Observe     Observer
	// Replay may reuse a completed, unredacted result. The host must bind source,
	// invocation, capability, index and InputDigest, and reject unknown outcomes.
	Replay func(context.Context, Observation) (any, bool, error)
}

// Decision is local audit evidence. Reason describes validation or fallback, not
// an invented upstream model explanation. Jev never authorizes privileged work.
type Decision struct {
	Choice     string   `json:"choice"`
	Reason     string   `json:"reason"`
	Defaulted  bool     `json:"defaulted"`
	Confidence *float64 `json:"confidence"`
}

// Observation records an ordered capability attempt, not proof of an external
// effect. Index is invocation-local. An unknown outcome must be reconciled before
// retrying. Inputs are redacted, while InputDigest identifies their original data.
type Observation struct {
	Index       int            `json:"index"`
	Phase       string         `json:"phase"`
	Capability  string         `json:"capability"`
	InputDigest string         `json:"input_digest"`
	Inputs      map[string]any `json:"inputs"`
	Result      any            `json:"result,omitempty"`
	Error       string         `json:"error,omitempty"`
	StartedAt   time.Time      `json:"started_at"`
	FinishedAt  time.Time      `json:"finished_at,omitzero"`
	Duration    time.Duration  `json:"duration_ns,omitempty"`
	Redacted    bool           `json:"redacted,omitempty"`
}

// Observer is called synchronously outside any controller lock. It may reject a
// call before execution. Failure after execution does not undo external effects.
type Observer func(context.Context, Observation) error

// Host supplies invocation state without exposing the engine to capabilities.
// Options are normalized by the engine and must not be mutated by a capability.
// MarkStarted and MarkCompleted delimit external work for unknown-outcome tracking.
type Host interface {
	Context() context.Context
	Options() Options
	MarkStarted()
	MarkCompleted()
	RecordDecision(Decision)
}

// Call is a capability implementation invoked after the engine's admission checks.
type Call func(Host, *starlark.Builtin, starlark.Tuple, []starlark.Tuple) (starlark.Value, error)
