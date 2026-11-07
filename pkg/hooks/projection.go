package hooks

import (
	"encoding/json"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/elee1766/jobshop/pkg/store"
)

const (
	hookAuditInlineBytes   = 4 << 10
	hookAuditArtifactBytes = 1 << 20
	hookAuditAttempts      = 50
	hookAuditCalls         = 200
)

type hookPublicEvent struct {
	Source     string `json:"source"`
	Event      string `json:"event"`
	Delivery   string `json:"delivery"`
	Repository string `json:"repository,omitempty"`
}

type hookPublicInvocation struct {
	ResultMetadata   any              `json:"result_metadata,omitempty"`
	ResultOutput     any              `json:"result_output,omitempty"`
	AcceptedAt       time.Time        `json:"accepted_at"`
	ParentInvocation string           `json:"parent_invocation,omitempty"`
	Targets          any              `json:"targets,omitempty"`
	Trigger          string           `json:"trigger,omitempty"`
	Run              string           `json:"run,omitempty"`
	Step             string           `json:"step,omitempty"`
	StepRevision     int64            `json:"step_revision,omitempty"`
	Entry            string           `json:"entry"`
	Digest           string           `json:"digest"`
	Event            *hookPublicEvent `json:"event,omitempty"`
	Error            string           `json:"error,omitempty"`
	Output           any              `json:"output,omitempty"`
	Decisions        any              `json:"decisions,omitempty"`
	Actions          any              `json:"actions,omitempty"`
}

type hookCallReuse struct {
	Key           string `json:"key"`
	Index         int    `json:"index"`
	SourceAttempt int64  `json:"source_attempt"`
}

type hookPublicAttempt struct {
	ResultMetadata any       `json:"result_metadata,omitempty"`
	ResultOutput   any       `json:"result_output,omitempty"`
	Invocation     string    `json:"invocation"`
	Attempt        int64     `json:"attempt"`
	StartedAt      time.Time `json:"started_at"`
	FinishedAt     time.Time `json:"finished_at,omitzero"`
	Error          string    `json:"error,omitempty"`
	Output         any       `json:"output,omitempty"`
	Decisions      any       `json:"decisions,omitempty"`
	Actions        any       `json:"actions,omitempty"`
	Targets        any       `json:"targets,omitempty"`
	ReusedCalls    any       `json:"reused_calls,omitempty"`
}

type hookPublicObservation struct {
	Index      int           `json:"index"`
	Phase      string        `json:"phase"`
	Capability string        `json:"capability"`
	Inputs     any           `json:"inputs,omitempty"`
	Result     any           `json:"result,omitempty"`
	Error      string        `json:"error,omitempty"`
	StartedAt  time.Time     `json:"started_at"`
	FinishedAt time.Time     `json:"finished_at,omitzero"`
	Duration   time.Duration `json:"duration_ns,omitempty"`
	Redacted   bool          `json:"redacted,omitempty"`
}

type hookPublicCall struct {
	Invocation  string                `json:"invocation"`
	Attempt     int64                 `json:"attempt"`
	StartedAt   time.Time             `json:"started_at"`
	FinishedAt  time.Time             `json:"finished_at,omitzero"`
	ElapsedNS   int64                 `json:"elapsed_ns,omitempty"`
	Observation hookPublicObservation `json:"observation"`
}

// Public evidence is saved before credentials can rotate or be removed.
func (c *Engine) hookInvocationPublic(in Invocation) json.RawMessage {
	out := hookPublicInvocation{AcceptedAt: in.AcceptedAt, ParentInvocation: in.ParentInvocation, Trigger: in.Trigger, Run: in.Run, Step: in.Step, StepRevision: in.StepRevision, Entry: in.Entry, Digest: in.Digest, Error: c.hookAuditText(in.Error)}
	if in.Event != nil {
		out.Event = &hookPublicEvent{Source: in.Event.Source, Event: in.Event.Event, Delivery: in.Event.Delivery, Repository: in.Event.Repository}
	}
	out.Targets, out.Output, out.Decisions, out.Actions = c.hookAuditField(in.Targets), c.hookAuditField(in.Output), c.hookAuditField(in.Decisions), c.hookAuditField(in.Actions)
	out.ResultMetadata, out.ResultOutput = c.hookAuditField(in.ResultMetadata), c.hookAuditField(in.ResultOutput)
	return c.hookAuditJSON(out)
}

func (c *Engine) hookAttemptPublic(in hookAttempt, evidence json.RawMessage) json.RawMessage {
	out := hookPublicAttempt{Invocation: in.Invocation, Attempt: in.Attempt, StartedAt: in.StartedAt, FinishedAt: in.FinishedAt}
	var accepted hookPublicInvocation
	if DecodeJSON(evidence, &accepted) == nil {
		out.Error, out.Targets, out.Output, out.Decisions, out.Actions = accepted.Error, accepted.Targets, accepted.Output, accepted.Decisions, accepted.Actions
		out.ResultMetadata, out.ResultOutput = accepted.ResultMetadata, accepted.ResultOutput
	}
	out.ReusedCalls = c.hookAuditField(in.ReusedCalls)
	return c.hookAuditJSON(out)
}

func (c *Engine) hookPublicError(evidence json.RawMessage, message string) json.RawMessage {
	var out hookPublicInvocation
	if evidence != nil && DecodeJSON(evidence, &out) != nil {
		out = hookPublicInvocation{}
	}
	out.Error = c.hookAuditText(message)
	return c.hookAuditJSON(out)
}

func (c *Engine) hookCallPublic(in hookCallRecord) json.RawMessage {
	o := in.Observation
	out := hookPublicCall{
		Invocation: in.Invocation, Attempt: in.Attempt, StartedAt: in.StartedAt, FinishedAt: in.FinishedAt, ElapsedNS: in.ElapsedNS,
		Observation: hookPublicObservation{Index: o.Index, Phase: o.Phase, Capability: o.Capability, Inputs: c.hookAuditField(o.Inputs), Result: c.hookAuditField(o.Result), Error: c.hookAuditText(o.Error), StartedAt: o.StartedAt, FinishedAt: o.FinishedAt, Duration: o.Duration, Redacted: o.Redacted},
	}
	return c.hookAuditJSON(out)
}

func (c *Engine) hookAuditText(text string) string {
	text = c.hookAuditSanitize(text).(string)
	if len(text) > 1024 {
		return strings.ToValidUTF8(text[:1024], "") + "…"
	}
	return text
}

func (c *Engine) hookAuditJSON(value any) json.RawMessage {
	data, err := json.Marshal(c.hookAuditSanitize(value))
	if err != nil {
		return json.RawMessage(`{"omitted":true}`)
	}
	return data
}

func (c *Engine) hookAuditField(value any) any {
	if value == nil {
		return nil
	}
	value = c.hookAuditSanitize(value)
	data, err := json.Marshal(value)
	if err != nil {
		return map[string]any{"omitted": true}
	}
	if len(data) <= hookAuditInlineBytes {
		return value
	}
	originalBytes := len(data)
	truncated := originalBytes > hookAuditArtifactBytes
	if truncated {
		data, _ = json.Marshal(map[string]any{"omitted": true, "bytes": originalBytes, "reason": "audit artifact byte limit"})
	}
	digest := hookKey(string(data))
	relative := filepath.Join("automation", "audit", digest+".json")
	path := filepath.Join(c.root, relative)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return map[string]any{"omitted": true, "bytes": originalBytes}
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return map[string]any{"omitted": true, "bytes": originalBytes}
	}
	return map[string]any{"artifact": relative, "bytes": originalBytes, "digest": digest, "truncated": truncated}
}

func sensitiveAuditKey(key string) bool {
	key = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, key)
	for _, sensitive := range []string{"secret", "password", "passwd", "privatekey", "apikey", "authorization", "cookie", "credential", "accesskey", "signingkey"} {
		if strings.Contains(key, sensitive) {
			return true
		}
	}
	return key == "token" || strings.HasSuffix(key, "token")
}

func (c *Engine) hookAuditSanitize(value any) any {
	data, err := json.Marshal(value)
	if err != nil {
		return "[OMITTED]"
	}
	if err := DecodeJSON(data, &value); err != nil {
		return "[OMITTED]"
	}
	c.hookAuditMu.Lock()
	secrets := slices.Clone(c.hookAuditSecrets)
	c.hookAuditMu.Unlock()
	slices.SortFunc(secrets, func(a, b string) int { return len(b) - len(a) })
	nodes := 0
	var clean func(any, int) any
	clean = func(value any, depth int) any {
		nodes++
		if nodes > 10000 || depth > 32 {
			return "[OMITTED]"
		}
		switch v := value.(type) {
		case map[string]any:
			out := make(map[string]any, min(len(v), 10000))
			for _, key := range slices.Sorted(maps.Keys(v)) {
				if nodes > 10000 {
					out["audit_omitted"] = true
					break
				}
				name := clean(key, depth+1).(string)
				if sensitiveAuditKey(key) {
					out[name] = "[REDACTED]"
				} else {
					out[name] = clean(v[key], depth+1)
				}
			}
			return out
		case []any:
			out := make([]any, 0, min(len(v), 10000))
			for _, item := range v {
				if nodes > 10000 {
					out = append(out, "[OMITTED]")
					break
				}
				out = append(out, clean(item, depth+1))
			}
			return out
		case string:
			for _, secret := range secrets {
				if secret != "" {
					v = strings.ReplaceAll(v, secret, "[REDACTED]")
				}
			}
			trimmed := strings.TrimSpace(v)
			if depth < 32 && len(trimmed) > 1 && (trimmed[0] == '{' || trimmed[0] == '[') {
				var nested any
				if DecodeJSON([]byte(trimmed), &nested) == nil {
					encoded, err := json.Marshal(clean(nested, depth+1))
					if err == nil {
						return string(encoded)
					}
				}
			}
			return v
		default:
			return value
		}
	}
	return clean(value, 0)
}

// Remember retains accepted secrets so credential rotation cannot expose old evidence.
func (c *Engine) Remember(runtime *Runtime) {
	c.hookAuditMu.Lock()
	defer c.hookAuditMu.Unlock()
	for _, program := range runtime.Programs {
		for _, secret := range program.secrets {
			if !slices.Contains(c.hookAuditSecrets, secret) {
				c.hookAuditSecrets = append(c.hookAuditSecrets, secret)
			}
		}
	}
}

// The record envelope never exposes the private worker holder or lease.
func publicHookRecord(rec store.Record, data json.RawMessage) store.Record {
	return store.Record{Collection: rec.Collection, Key: rec.Key, State: rec.State, Value: data, Revision: rec.Revision, Attempts: rec.Attempts, CreatedAt: rec.CreatedAt, UpdatedAt: rec.UpdatedAt}
}

func (c *Engine) publicHookInvocation(rec store.Record, summary bool) (store.Record, error) {
	var in Invocation
	if DecodeJSON(rec.Value, &in) != nil {
		return store.Record{}, errors.New("invalid automation invocation evidence")
	}
	out := hookPublicInvocation{AcceptedAt: in.AcceptedAt, ParentInvocation: in.ParentInvocation, Trigger: in.Trigger, Run: in.Run, Step: in.Step, StepRevision: in.StepRevision, Entry: in.Entry, Digest: in.Digest}
	if in.Public != nil && DecodeJSON(in.Public, &out) != nil {
		return store.Record{}, errors.New("invalid public automation invocation evidence")
	}
	if in.Public == nil && in.Event != nil {
		out.Event = &hookPublicEvent{Source: in.Event.Source, Event: in.Event.Event, Delivery: in.Event.Delivery, Repository: in.Event.Repository}
	}
	if summary {
		out.Actions, out.Decisions, out.Output, out.Targets = nil, nil, nil, nil
		out.ResultMetadata, out.ResultOutput = nil, nil
	}
	texts := []*string{&out.ParentInvocation, &out.Trigger, &out.Run, &out.Step, &out.Entry, &out.Digest, &out.Error}
	if out.Event != nil {
		texts = append(texts, &out.Event.Source, &out.Event.Event, &out.Event.Delivery, &out.Event.Repository)
	}
	for _, text := range texts {
		*text = c.hookAuditText(*text)
	}
	return publicHookRecord(rec, c.hookAuditJSON(out)), nil
}

func (c *Engine) publicHookAttempt(rec store.Record) (store.Record, error) {
	var in hookAttempt
	if DecodeJSON(rec.Value, &in) != nil {
		return store.Record{}, errors.New("invalid automation attempt evidence")
	}
	out := hookPublicAttempt{Invocation: in.Invocation, Attempt: in.Attempt, StartedAt: in.StartedAt, FinishedAt: in.FinishedAt}
	if in.Public != nil && DecodeJSON(in.Public, &out) != nil {
		return store.Record{}, errors.New("invalid public automation attempt evidence")
	}
	out.Invocation, out.Error = c.hookAuditText(out.Invocation), c.hookAuditText(out.Error)
	return publicHookRecord(rec, c.hookAuditJSON(out)), nil
}

func (c *Engine) publicHookCall(rec store.Record) (store.Record, error) {
	var in hookCallRecord
	if DecodeJSON(rec.Value, &in) != nil {
		return store.Record{}, errors.New("invalid automation call evidence")
	}
	out := hookPublicCall{
		Invocation: in.Invocation, Attempt: in.Attempt, StartedAt: in.StartedAt, FinishedAt: in.FinishedAt, ElapsedNS: in.ElapsedNS,
		Observation: hookPublicObservation{Index: in.Observation.Index, Phase: in.Observation.Phase, Capability: in.Observation.Capability},
	}
	if in.Public != nil && DecodeJSON(in.Public, &out) != nil {
		return store.Record{}, errors.New("invalid public automation call evidence")
	}
	out.Invocation = c.hookAuditText(out.Invocation)
	out.Observation.Capability, out.Observation.Error = c.hookAuditText(out.Observation.Capability), c.hookAuditText(out.Observation.Error)
	return publicHookRecord(rec, c.hookAuditJSON(out)), nil
}
