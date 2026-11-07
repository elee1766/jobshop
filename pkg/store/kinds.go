package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Kind is an engine-defined ticket kind. Packs cannot add kinds.
type Kind string

// Engine kinds.
const (
	KindTask    Kind = "task"
	KindSession Kind = "session"
	KindMessage Kind = "message"
	KindRun     Kind = "run"
	KindStep    Kind = "step"
	KindBatch   Kind = "batch"
)

// ClaimableKinds are the kinds ClaimNext and Ready consider by default.
// Messages are consumed through the inbox, sessions are never claimed,
// runs have no pool and are tracked by the controller.
var ClaimableKinds = []Kind{KindTask, KindStep}

// Fields is implemented by each kind's typed field struct.
type Fields interface {
	Validate() error
}

// TaskFields is the engine state of a task.
type TaskFields struct {
	// Wait parks the task until outside state holds (#381).
	Wait *Wait `json:"wait,omitempty"`
}

// Validate implements Fields.
func (f TaskFields) Validate() error { return f.Wait.Validate() }

// SessionFields is the engine state of a session ticket.
type SessionFields struct {
	Key string `json:"key,omitempty"` // e.g. matrix:<room>; unique while open
	// DisplayName is a human label for lists (#626). Not unique, never
	// used for lookup or dedup. Empty means a default is shown.
	DisplayName string     `json:"display_name,omitempty"`
	Named       bool       `json:"named,omitempty"`
	Archived    *time.Time `json:"archived,omitempty"`
	Agent       string     `json:"agent"`
	ACPSession  string     `json:"acp_session,omitempty"`
	Current     string     `json:"current,omitempty"` // ticket id being worked
	// TraceParent is the session trace's W3C traceparent, kept so lifecycle
	// spans after a controller restart join the same trace.
	TraceParent string `json:"traceparent,omitempty"`
	// Cordoned takes the session out of rotation (#281): it claims no new
	// work. Nil means in rotation.
	Cordoned *Cordon `json:"cordoned,omitempty"`
	// Model and Effort are the agent's current values as it last reported
	// them (#304), for session list, Desk and telemetry.
	Model  string `json:"model,omitempty"`
	Effort string `json:"effort,omitempty"`
	// AgentName and AgentVersion are the agentInfo the agent reported
	// from ACP initialize (e.g. "jcode"), so clients can switch display.
	AgentName    string `json:"agent_name,omitempty"`
	AgentVersion string `json:"agent_version,omitempty"`
	// Options are the ACP config options the controller set on the
	// conversation; ConfigIDs the option ids the agent advertised. Both
	// survive a controller restart, so an adopted host is not set again.
	Options   map[string]string `json:"options,omitempty"`
	ConfigIDs []string          `json:"config_ids,omitempty"`
	// Override is `session config` (#304): option values on top of the
	// pool's until the session ends or the override is reset.
	Override map[string]string `json:"override,omitempty"`
	// ShippedSeq is the last journal seq shipped to VictoriaLogs (#530).
	// The controller saves it every few seconds, so after a crash it
	// resends a little; rows are deduped by (session, seq) on query.
	ShippedSeq int64 `json:"shipped_seq,omitempty"`
	// Slot is the session's index in its pool's isolation per_session id
	// range (#540): it runs as map_uid+slot. Nil without per_session.
	Slot *int `json:"slot,omitempty"`
	// Prestart marks an attention chat the controller started before
	// anyone clicked it (#606). Idle, it is suspended rather than ended.
	Prestart bool `json:"prestart,omitempty"`
	// ReadOnly limits the session to read methods and read tool calls
	// until a human sends it a message (#606).
	ReadOnly bool `json:"read_only,omitempty"`
	// Suspended: the host was stopped while idle and the session stays
	// open; the next message resumes it with session/load (#606).
	Suspended bool `json:"suspended,omitempty"`
	// Operator is the last human who sent the session a message, so a
	// question a chat answers for them records them as from (JOB-52).
	Operator string `json:"operator,omitempty"`
	// Follow is a question chat waiting on the asking worker's reaction
	// to the answer it relayed (JOB-52).
	Follow *QuestionFollow `json:"follow,omitempty"`
	// Ended is the host state read from the session directory when the
	// session ended (JOB-32), so session list does not read the
	// directory of every closed session.
	Ended *SessionEnded `json:"ended,omitempty"`
}

// QuestionFollow is an answered question whose asker's reaction the chat
// reports: Ticket is the work ticket, Asker the asking session, Message
// the answer delivered to it (the turn that delivers it is the reaction).
type QuestionFollow struct {
	Question string    `json:"question"`
	Ticket   string    `json:"ticket"`
	Asker    string    `json:"asker,omitempty"`
	Message  string    `json:"message,omitempty"`
	Since    time.Time `json:"since"`
}

// SessionEnded is what session list shows for a closed session's host.
type SessionEnded struct {
	HostPid int        `json:"host_pid,omitempty"`
	Seq     int64      `json:"seq,omitempty"`
	SeqTime *time.Time `json:"seq_time,omitempty"`
	Exit    string     `json:"exit,omitempty"`
}

// Cordon reasons (#281).
const (
	// CordonScale: cordoned by the controller to scale the pool down; the
	// session ends scaled_down once it holds nothing.
	CordonScale = "scale"
	// CordonManual: `session cordon`; the session stays open and idle.
	CordonManual = "manual"
	// CordonDrain: `session drain`; the session ends drained once idle.
	CordonDrain = "drain"
)

// Cordon records why and by whom a session was cordoned.
type Cordon struct {
	Reason string    `json:"reason"`
	By     string    `json:"by,omitempty"`
	At     time.Time `json:"at"`
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
	// Context is delivered with the message as a prompt block for the
	// model only (session.send context).
	Context string `json:"context,omitempty"`
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

// Run states (RunFields.State).
const (
	RunRunning    = "running"
	RunNeedsHuman = "needs_human"
	RunDone       = "done"
	RunCancelled  = "cancelled"
)

// ResolutionRebase closes a step to route its run back for a rebase
// (#424): like "failed" the step reopens, but no failure is counted,
// since master moving under a PR is not the step's fault. On a
// non-step ticket it is an ordinary close.
const ResolutionRebase = "rebase"

// ResolutionRequestChanges closes a reviewing step to send its findings
// back to the step it reviews (JOB-98). In one transaction the reviewer
// counts a review round and reopens, blocked behind the target, and the
// target reopens with the findings in fields.changes. No failure is
// counted. Past the reviewer's round cap the reviewer stays closed with
// this resolution, the target is left alone and the run needs a human.
const ResolutionRequestChanges = "request_changes"

// DefaultReviewRounds is a step's review round cap when its workflow
// sets no max_review_rounds.
const DefaultReviewRounds = 3

// RunFields is the engine state of a workflow run. Its steps are its
// children (parent_id).
type RunFields struct {
	Workflow string            `json:"workflow"`
	Pack     string            `json:"pack,omitempty"` // file that declared the workflow
	Vars     map[string]string `json:"vars,omitempty"`
	Key      string            `json:"key,omitempty"` // dedup: one open run per workflow and key
	State    string            `json:"state"`
	// Escalations counts needs_human transitions; each files one
	// escalation ticket keyed on it.
	Escalations int `json:"escalations,omitempty"`
}

// Validate implements Fields.
func (f RunFields) Validate() error {
	if f.Workflow == "" {
		return fmt.Errorf("%w: run: workflow is required", ErrInvalid)
	}
	switch f.State {
	case RunRunning, RunNeedsHuman, RunDone, RunCancelled:
		return nil
	}
	return fmt.Errorf("%w: run: unknown state %q", ErrInvalid, f.State)
}

// Batch states (BatchFields.State).
const (
	BatchOpen       = "open"
	BatchRunning    = "running"
	BatchNeedsHuman = "needs_human"
	BatchDone       = "done"
	BatchFailed     = "failed"
)

// BatchFields is an ordered set of runs started together (#148, #169).
// The member runs are its children. A batch has no pool and is never
// claimed.
type BatchFields struct {
	Key    string      `json:"key"` // one open batch per project and key
	Issues []string    `json:"issues"`
	Order  [][2]string `json:"order,omitempty"` // [a, b]: b starts after a's last step
	State  string      `json:"state"`
}

// Validate implements Fields.
func (f BatchFields) Validate() error {
	if f.Key == "" {
		return fmt.Errorf("%w: batch: key is required", ErrInvalid)
	}
	switch f.State {
	case BatchOpen, BatchRunning, BatchNeedsHuman, BatchDone, BatchFailed:
		return nil
	}
	return fmt.Errorf("%w: batch: unknown state %q", ErrInvalid, f.State)
}

// StepFields places a step ticket in its run.
type StepFields struct {
	Run      string `json:"run"`
	Step     string `json:"step"`
	Index    int    `json:"index"`
	Retries  int    `json:"retries"`
	Failures int    `json:"failures"`
	// Wait parks the step until outside state holds (#381).
	Wait *Wait `json:"wait,omitempty"`
	// Resources are the resource keys the step holds while in progress
	// (#400). It is ready only when each has a free slot.
	Resources []string `json:"resources,omitempty"`
	// Type is command or http for steps the controller runs (#396);
	// empty is an agent step.
	Type string `json:"type,omitempty"`
	// Command is the last attempt's filled command line or "METHOD url",
	// Output its bounded, redacted stdout or response body, Outcome the
	// outcome word it set, and Result ok, not_yet or failed.
	Command string `json:"command,omitempty"`
	Output  string `json:"output,omitempty"`
	Outcome string `json:"outcome,omitempty"`
	Result  string `json:"result,omitempty"`
	// Needs are the step ids this step waits on (#385), for run show.
	Needs []string `json:"needs,omitempty"`
	// When gates the step on another step's outcome (#385). A gated step
	// is created deferred; the controller opens or skips it once its
	// needs are finished.
	When *StepWhen `json:"when,omitempty"`
	// Placement is the step's resolved placement (#409): its options by
	// weight and which one the ticket is on now.
	Placement *StepPlacement `json:"placement,omitempty"`
	// RetryPool is the resolved retry_pool (JOB-59): the controller moves
	// the step here after a failed attempt or a review resume.
	RetryPool string `json:"retry_pool,omitempty"`
	// OnRequestChanges is the step a request_changes close of this step
	// reopens (JOB-98), MaxReviewRounds how many it may make (nil:
	// DefaultReviewRounds) and ReviewRounds how many it made.
	OnRequestChanges string `json:"on_request_changes,omitempty"`
	MaxReviewRounds  *int   `json:"max_review_rounds,omitempty"`
	ReviewRounds     int    `json:"review_rounds,omitempty"`
	// Changes is the open change request against this step: set when a
	// reviewer's request_changes reopens it, cleared when it closes done.
	Changes *StepChanges `json:"changes,omitempty"`
}

// ReviewRoundLimit is MaxReviewRounds or the default.
func (f StepFields) ReviewRoundLimit() int {
	if f.MaxReviewRounds == nil {
		return DefaultReviewRounds
	}
	return *f.MaxReviewRounds
}

// StepChanges is a reviewer's change request (JOB-98), shown to the
// session that claims the reopened step.
type StepChanges struct {
	Step     string `json:"step"`   // the reviewer's step name
	Ticket   string `json:"ticket"` // the reviewer's ticket id
	Round    int    `json:"round"`
	Max      int    `json:"max"`
	Findings string `json:"findings"`
}

// StepPlacement records a step's placement options. The ticket's pool
// and Resources are the chosen option's; the controller moves the step
// to another option while it is open and unclaimed.
type StepPlacement struct {
	Options []PlacementOption `json:"options"`
	// Chosen indexes Options. Fallback is Chosen > 0.
	Chosen int `json:"chosen"`
	// PreferMS is prefer_for in milliseconds.
	PreferMS int64 `json:"prefer_ms,omitempty"`
	// Since is when the step first could start (its blockers finished),
	// set by the controller; prefer_for counts from it.
	Since time.Time `json:"since,omitzero"`
}

// PlacementOption is one resolved option: a pool and every resource key
// it needs (required keys included).
type PlacementOption struct {
	Pool      string   `json:"pool"`
	Resources []string `json:"resources,omitempty"`
	Weight    int      `json:"weight"`
}

// Fallback reports whether the step is on an option below the top one.
func (p *StepPlacement) Fallback() bool { return p != nil && p.Chosen > 0 }

// StepWhen is a step condition: step closed with outcome.
type StepWhen struct {
	Step    string `json:"step"`
	Outcome string `json:"outcome"`
}

// StepSkipped is the resolution of a step whose when did not match. It
// unblocks dependents like done.
const StepSkipped = "skipped"

// StepFinished reports whether a step ticket closed done or skipped,
// the resolutions that let the steps after it run.
func StepFinished(t Ticket) bool {
	return t.Status == StatusClosed && (t.Resolution == "done" || t.Resolution == StepSkipped)
}

// Auto reports whether the controller runs the step itself.
func (f StepFields) Auto() bool { return f.Type == "command" || f.Type == "http" }

// StepFailure counts one failed attempt on a step ticket: a close
// "failed", a release while unclosed, or an expired lease. It bumps
// fields.failures and reports whether the retry limit is reached (the
// step then closes failed instead of reopening). Backends call it inside
// the transaction of the failing event. Non-step tickets are unchanged.
func StepFailure(t *Ticket) (f StepFields, exhausted, isStep bool, err error) {
	if t.Kind != KindStep {
		return f, false, false, nil
	}
	if f, err = DecodeFields[StepFields](*t); err != nil {
		return f, false, true, err
	}
	f.Failures++
	b, err := json.Marshal(f)
	if err != nil {
		return f, false, true, err
	}
	t.Fields = b
	return f, f.Failures >= f.Retries, true, nil
}

// StepFailureComment is the comment a backend adds for a counted failure.
func StepFailureComment(f StepFields, how string, exhausted bool) string {
	if exhausted {
		return fmt.Sprintf("Attempt %d failed (%s). Retry limit %d reached: step closed failed, run needs a human.", f.Failures, how, f.Retries)
	}
	return fmt.Sprintf("Attempt %d of %d failed (%s); step reopened for another attempt.", f.Failures, f.Retries, how)
}

// Validate implements Fields.
func (f StepFields) Validate() error {
	if f.Run == "" || f.Step == "" {
		return fmt.Errorf("%w: step: run and step are required", ErrInvalid)
	}
	if f.Index < 0 || f.Retries < 0 || f.Failures < 0 || f.ReviewRounds < 0 || (f.MaxReviewRounds != nil && *f.MaxReviewRounds < 0) {
		return fmt.Errorf("%w: step: index, retries, failures and review rounds must be >= 0", ErrInvalid)
	}
	for _, r := range f.Resources {
		if r == "" {
			return fmt.Errorf("%w: step: empty resource key", ErrInvalid)
		}
	}
	if p := f.Placement; p != nil {
		if len(p.Options) == 0 || p.Chosen < 0 || p.Chosen >= len(p.Options) {
			return fmt.Errorf("%w: step: placement chosen out of range", ErrInvalid)
		}
	}
	return f.Wait.Validate()
}

func (k Kind) newFields() (Fields, bool) {
	switch k {
	case KindTask:
		return &TaskFields{}, true
	case KindSession:
		return &SessionFields{}, true
	case KindMessage:
		return &MessageFields{}, true
	case KindRun:
		return &RunFields{}, true
	case KindStep:
		return &StepFields{}, true
	case KindBatch:
		return &BatchFields{}, true
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
