// Package api defines the shared contract between the automation engine and its
// host capabilities. It does not depend on the engine or capability packages.
package api

import (
	"context"
	"io"
	"net/http"
	"time"

	"go.starlark.net/starlark"

	"github.com/gfx-labs/jobshop/pkg/integrations/github"
	"github.com/gfx-labs/jobshop/pkg/integrations/linear"
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
	// GitHub, when set, enables the github built-in: it returns the host's
	// client for the identity covering repo.
	GitHub GitHubClient
	// Linear, when set, enables the linear built-in: it returns the host's
	// client for the identity covering a team key.
	Linear LinearClient
	// Tickets, when set, enables the read-only tickets, runs and sessions
	// built-ins over the controller's store.
	Tickets TicketReader
	// Now is time.now() in scripts; nil is the wall clock. A host pins it per
	// invocation so a retried callback sees the same time.
	Now func(context.Context) time.Time
	// Exec, when set, runs shell's argv instead of a direct child process: a
	// step function's shell runs in its pool's sandbox and worktree.
	Exec    Exec
	Observe Observer
	// Replay may reuse a completed, unredacted result. The host must bind source,
	// invocation, capability, index and InputDigest, and reject unknown outcomes.
	Replay func(context.Context, Observation) (any, bool, error)
}

// TicketQuery is a tickets() read. Meta keys are "pack.key".
type TicketQuery struct {
	Kind     string
	Statuses []string
	Project  string
	Pool     string
	Label    string
	Workflow string
	Key      string
	State    string
	Parent   string
	Holder   string
	Meta     map[string]string
	Order    string
	Limit    int
	// Comments adds a "comments" list, oldest first, to every result.
	Comments bool
}

// TicketReader returns matching tickets as JSON-shaped maps.
type TicketReader func(ctx context.Context, q TicketQuery) ([]any, error)

// GitHubClient returns the client for owner/repo. An empty repo asks for
// any identity, which a github-token identity answers. The client mints its
// token per request and reports it through the context's apicall.Observer.
type GitHubClient func(ctx context.Context, repo string) (*github.Client, error)

// LinearClient returns the client for a Linear team key. An empty team asks
// for the only Linear identity. The client mints its token per request and
// reports it through the context's apicall.Observer.
type LinearClient func(ctx context.Context, team string) (*linear.Client, error)

// ExecRequest is one shell call after argument validation. Dir is empty when
// the script set no cwd, so the host picks its own default.
type ExecRequest struct {
	Argv   []string
	Dir    string
	Env    map[string]string
	Stdout io.Writer
	Stderr io.Writer
}

// Exec runs one command and returns its exit code. A nonzero exit is not an
// error; an error means the command did not run or did not finish.
type Exec func(context.Context, ExecRequest) (int, error)

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
	// AddSecret redacts a credential obtained during the call from logs,
	// errors and observations of the rest of the invocation.
	AddSecret(string)
}

// Call is a capability implementation invoked after the engine's admission checks.
type Call func(Host, *starlark.Builtin, starlark.Tuple, []starlark.Tuple) (starlark.Value, error)
