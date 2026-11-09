package hooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
)

// GuardEvent is the JSON a guardrail hook receives on stdin.
type GuardEvent struct {
	On      string          `json:"on"`
	Action  string          `json:"action,omitempty"`
	Target  string          `json:"target,omitempty"`
	Detail  string          `json:"detail,omitempty"`
	Session string          `json:"session,omitempty"`
	Pool    string          `json:"pool,omitempty"`
	Project string          `json:"project,omitempty"`
	Ticket  *domain.Ticket  `json:"ticket,omitempty"`
	Extra   json.RawMessage `json:"extra,omitempty"`
}

// Answer is a guardrail decision, interpreted by the host's admission policy.
type Answer struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

// CommandResult separates process outcomes from host-specific sandbox execution.
type CommandResult struct {
	Output   []byte
	Stderr   string
	ExitCode int
	Signal   string
}

func (r CommandResult) failure() error {
	if r.ExitCode == 0 && r.Signal == "" {
		return nil
	}
	exit := fmt.Sprint(r.ExitCode)
	if r.Signal != "" {
		exit = "signal " + r.Signal
	}
	return fmt.Errorf("exited %s: %s", exit, strings.TrimSpace(r.Stderr))
}

// Executor runs a command using host-owned process groups and pool isolation.
type Executor interface {
	ExecuteHook(context.Context, string, string, []byte) (CommandResult, error)
}

// Runner handles legacy hook protocols independently of sandbox mechanics.
type Runner struct {
	Dir      string
	After    func(time.Duration) <-chan time.Time
	Executor Executor
}

// Run fails closed on a timeout, command failure, or unreadable answer.
func (r Runner) Run(ctx context.Context, h config.Hook, ev GuardEvent) Answer {
	in, err := json.Marshal(ev)
	if err != nil {
		return Answer{domain.GuardDeny, "hook event: " + err.Error()}
	}
	timeout := config.Duration(h.Timeout, 30*time.Second)
	if h.URL != "" {
		p := hookPost{url: h.URL, secret: h.SecretFile.In(r.Dir), timeout: timeout, retry: h.Retry, after: r.After}
		out, err := p.do(ctx, in)
		if err != nil {
			return Answer{domain.GuardDeny, "hook " + h.URL + ": " + err.Error()}
		}
		return parseHookAnswer(out)
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	result, err := r.Executor.ExecuteHook(ctx, h.Command, ev.Pool, in)
	if ctx.Err() != nil {
		return Answer{domain.GuardDeny, "hook timed out: " + h.Command}
	}
	if err != nil {
		return Answer{domain.GuardDeny, "hook: " + err.Error()}
	}
	if err := result.failure(); err != nil {
		return Answer{domain.GuardDeny, "hook " + err.Error()}
	}
	return parseHookAnswer(result.Output)
}

// Admission runs one admitted event's command or signed HTTP transport.
func (r Runner) Admission(ctx context.Context, h config.AdmissionHook, body []byte, delivery string) ([]byte, error) {
	timeout := config.Duration(h.Timeout, 30*time.Second)
	if h.URL != "" {
		p := hookPost{url: h.URL, secret: h.SecretFile.In(r.Dir), timeout: timeout, retry: h.Retry, after: r.After, headers: map[string]string{"X-Jobshop-Delivery": delivery}}
		out, err := p.do(ctx, body)
		return bytes.TrimSpace(out), err
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	result, err := r.Executor.ExecuteHook(ctx, h.Command, "", body)
	if ctx.Err() != nil {
		return nil, fmt.Errorf("timed out after %s", timeout)
	}
	if err != nil {
		return nil, err
	}
	if err := result.failure(); err != nil {
		return nil, err
	}
	return bytes.TrimSpace(result.Output), nil
}

func parseHookAnswer(out []byte) Answer {
	out = bytes.TrimSpace(out)
	if len(out) == 0 {
		return Answer{Decision: domain.GuardAllow}
	}
	var a Answer
	if err := json.Unmarshal(out, &a); err != nil {
		return Answer{domain.GuardDeny, "hook answer is not JSON: " + err.Error()}
	}
	switch a.Decision {
	case domain.GuardAllow, domain.GuardDeny, domain.GuardAskHuman:
		return a
	}
	return Answer{domain.GuardDeny, fmt.Sprintf("hook decision %q is not allow, deny or ask_human", a.Decision)}
}
