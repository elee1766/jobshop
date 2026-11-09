package control

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/conversation"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/knowledge/memory"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/storage/store"
)

// Prompt rules (#205): operator guidance from config added to the prompts
// of matching sessions. A prompt is assembled from labeled parts, top to
// bottom:
//
//	system rules  (first turn of a host, again after a respawn)
//	before rules  (where the pool prompt goes)
//	pool prompt   (every ticket turn; a chat session's first message turn)
//	after rules   (where the pool prompt goes)
//	turn rules    (every ticket and message turn)
//	ticket or messages
//
// Within a position, rules keep config order (packs first, host last).

// PromptPart is one labeled piece of a turn's prompt (rpc.PromptPart).
type PromptPart = rpc.PromptPart

// Turn kinds (RuleEvent.Turn).
const (
	turnTicket  = "ticket"
	turnMessage = "message"
)

// RuleEvent is the JSON a rule command reads on stdin.
type RuleEvent struct {
	Rule    string `json:"rule"`
	At      string `json:"at"`
	Session string `json:"session"`
	Key     string `json:"key,omitempty"` // the session key, e.g. improve:<slug>
	Pool    string `json:"pool,omitempty"`
	Agent   string `json:"agent,omitempty"`
	Project string `json:"project,omitempty"`
	Turn    string `json:"turn"` // ticket or message
	Ticket  string `json:"ticket,omitempty"`
	Run     string `json:"run,omitempty"`
	Step    string `json:"step,omitempty"`
	Repo    string `json:"repo,omitempty"`
}

// turnSpec says which turn a prompt is assembled for.
type turnSpec struct {
	session, pool, agent, project string
	kind                          string         // turnTicket or turnMessage
	ticket                        *domain.Ticket // the claimed ticket, ticket turns only
	// system: the first turn of this host. poolPrompt: the pool prompt
	// (and before/after rules) belong in this turn.
	system, poolPrompt bool
}

// ruleSubject fills the match subject for a turn. Run, workflow and step
// come from the step ticket's fields and its run's; repo from
// meta.github.repo (#183).
func (c *Controller) ruleSubject(ctx context.Context, cfg *config.Config, ts turnSpec) (config.RuleSubject, RuleEvent) {
	p := cfg.Pools[ts.pool]
	s := config.RuleSubject{
		Pool: ts.pool, Agent: ts.agent, Project: ts.project,
		Packs: cfg.PoolPacks(ts.pool), Chat: p.Chat(),
	}
	ev := RuleEvent{Session: ts.session, Pool: ts.pool, Agent: ts.agent, Project: ts.project, Turn: ts.kind}
	if ts.session != "" {
		if st, err := c.store.Get(ctx, ts.session); err == nil {
			if f, err := domain.DecodeFields[domain.SessionFields](st); err == nil {
				s.Key, ev.Key = f.Key, f.Key
			}
		}
	}
	if ts.ticket == nil {
		return s, ev
	}
	t := *ts.ticket
	s.Ticket, s.Labels = true, t.Labels
	ev.Ticket = t.ID
	s.Repo = metaRepo(t)
	ev.Repo = s.Repo
	if t.Kind == domain.KindStep {
		if f, err := domain.DecodeFields[domain.StepFields](t); err == nil {
			s.Step, ev.Step, ev.Run = f.Step, f.Step, f.Run
			if run, err := c.store.Get(ctx, f.Run); err == nil {
				if rf, err := domain.DecodeFields[domain.RunFields](run); err == nil {
					s.Workflow = rf.Workflow
				}
			}
		}
	}
	return s, ev
}

// metaRepo is meta.github.repo ("owner/name"), or "".
func metaRepo(t domain.Ticket) string {
	var m struct {
		GitHub struct {
			Repo string `json:"repo"`
		} `json:"github"`
	}
	if len(t.Meta) == 0 || json.Unmarshal(t.Meta, &m) != nil {
		return ""
	}
	return m.GitHub.Repo
}

// assemble builds the labeled parts of a turn's prompt. body is the
// ticket or messages part. It runs rule commands.
func (c *Controller) assemble(ctx context.Context, ts turnSpec, body PromptPart) []PromptPart {
	cfg := c.config()
	subj, ev := c.ruleSubject(ctx, cfg, ts)
	byAt := map[string][]PromptPart{}
	for _, r := range cfg.PromptRules {
		switch r.At {
		case config.RuleSystem:
			if !ts.system {
				continue
			}
		case config.RuleBefore, config.RuleAfter:
			if !ts.poolPrompt {
				continue
			}
		}
		if !r.Match.Matches(subj) {
			continue
		}
		text := strings.TrimSpace(r.Text)
		if r.Command != "" {
			e := ev
			e.Rule, e.At = r.Name, r.At
			if out := c.runRuleCommand(ctx, r, e); out != "" {
				text = strings.TrimSpace(text + "\n\n" + out)
			}
		}
		if r.Memory != nil {
			if out := c.ruleMemory(r, subj); out != "" {
				text = strings.TrimSpace(text + "\n\n" + out)
			}
		}
		if text == "" {
			continue
		}
		byAt[r.At] = append(byAt[r.At], PromptPart{Source: "rule:" + r.Name, Text: text})
	}
	var parts []PromptPart
	parts = append(parts, byAt[config.RuleSystem]...)
	parts = append(parts, byAt[config.RuleBefore]...)
	if pp := strings.TrimSpace(cfg.Pools[ts.pool].Prompt); ts.poolPrompt && pp != "" {
		parts = append(parts, PromptPart{Source: "pool", Text: pp})
	}
	parts = append(parts, byAt[config.RuleAfter]...)
	parts = append(parts, byAt[config.RuleTurn]...)
	return append(parts, body)
}

// runRuleCommand runs a rule's hook with ev on stdin and returns its
// trimmed stdout, capped. A failure or timeout returns "" and logs WARN.
func (c *Controller) runRuleCommand(ctx context.Context, r config.PromptRule, ev RuleEvent) string {
	ctx, cancel := context.WithTimeout(ctx, config.Duration(r.Timeout, config.DefaultRuleTimeout))
	defer cancel()
	in, err := json.Marshal(ev)
	if err != nil {
		return ""
	}
	// The session's pool isolates the command like its agent (#810).
	cfg := c.config()
	ic := isolatedCmd{argv: shArgv(r.Command), pool: ev.Pool, stdin: bytes.NewReader(in), name: "rule", cfg: cfg}
	if !c.isolates(cfg, ev.Pool) {
		ic.pool, ic.dir = "", c.root.Dir
	}
	out := &capped{max: config.MaxRuleOutput}
	var stderr strings.Builder
	ic.stdout, ic.stderr = out, &stderr
	ex, err := c.runIsolated(ctx, ic)
	if err == nil && (ex.ExitCode != 0 || ex.Signal != "") {
		err = fmt.Errorf("exit %s", exitText(ex))
	}
	switch {
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		c.log.WarnContext(ctx, "prompt rule command timed out", "rule", r.Name, "session", ev.Session)
		return ""
	case err != nil:
		c.log.WarnContext(ctx, "prompt rule command failed", "rule", r.Name, "session", ev.Session,
			"err", err, "stderr", clipLog(stderr.String()))
		return ""
	}
	if out.over {
		c.log.WarnContext(ctx, "prompt rule command output truncated", "rule", r.Name, "max_bytes", config.MaxRuleOutput)
	}
	return strings.TrimSpace(strings.ToValidUTF8(out.b.String(), ""))
}

// capped keeps the first max bytes written and drops the rest.
type capped struct {
	b    bytes.Buffer
	max  int
	over bool
}

func (w *capped) Write(p []byte) (int, error) {
	if room := w.max - w.b.Len(); room < len(p) {
		w.over = true
		w.b.Write(p[:max(room, 0)])
		return len(p), nil
	}
	return w.b.Write(p)
}

func clipLog(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 500 {
		return s[:500] + "..."
	}
	return s
}

// joinParts is the prompt text sent to the agent.
func joinParts(parts []PromptPart) string {
	s := make([]string, 0, len(parts))
	for _, p := range parts {
		if p.Text != "" {
			s = append(s, p.Text)
		}
	}
	return strings.Join(s, "\n\n")
}

// ruleNames lists the rules among parts, for the prompt's _meta.
func ruleNames(parts []PromptPart) []string {
	var out []string
	for _, p := range parts {
		if n, ok := strings.CutPrefix(p.Source, "rule:"); ok {
			out = append(out, n)
		}
	}
	return out
}

// renderParts is `session prompt` output: each part under a label, with
// its injection status; a skipped part shows no text.
func renderParts(parts []PromptPart) string {
	var b strings.Builder
	for i, p := range parts {
		if i > 0 {
			b.WriteString("\n")
		}
		switch p.Status {
		case "":
			fmt.Fprintf(&b, "=== %s ===\n%s\n", p.Source, p.Text)
		case injectUnchanged:
			fmt.Fprintf(&b, "=== %s (unchanged, not sent) ===\n", p.Source)
		default:
			fmt.Fprintf(&b, "=== %s (%s) ===\n%s\n", p.Source, p.Status, p.Text)
		}
	}
	return b.String()
}

// sessionPrompt shows the prompt a session's next turn would get, each
// part labeled by source (#205). The turn is the ticket the session
// holds, else the next ready ticket in its pool, else a message turn.
// It runs rule commands but changes nothing: the ticket stays
// unclaimed and a pending system turn stays pending.
func (c *Controller) sessionPrompt(ctx context.Context, sc scope, p rpc.SessionPromptParams) (any, error) {
	if err := requireAdmin(sc, "session prompt"); err != nil {
		return nil, err
	}
	if p.ID == "" {
		return nil, fmt.Errorf("%w: session id is required", domain.ErrInvalid)
	}
	sess, err := c.openSessionTicket(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	f, err := domain.DecodeFields[domain.SessionFields](sess)
	if err != nil {
		return nil, err
	}
	ts := turnSpec{session: sess.ID, pool: sess.Pool, agent: f.Agent, project: sess.Project, kind: turnMessage}
	c.mu.Lock()
	r := c.runners[sess.ID]
	c.mu.Unlock()
	var delivered time.Time
	if r != nil {
		r.mu.Lock()
		ts.system = r.systemDue
		delivered = r.delivered
		r.mu.Unlock()
	}
	pool, havePool := c.config().Pools[sess.Pool]
	var t *domain.Ticket
	if held, err := c.store.List(ctx, store.Filter{Holder: sess.ID, Status: domain.StatusInProgress}); err == nil {
		for _, h := range held {
			if h.Kind != domain.KindMessage {
				t = &h
				break
			}
		}
	}
	if t == nil && havePool && !pool.Chat() {
		if ready, err := c.store.List(ctx, store.Filter{Project: sess.Project, Pool: sess.Pool, Ready: true, Limit: 1}); err == nil && len(ready) > 0 {
			t = &ready[0]
		}
	}
	var body PromptPart
	if t != nil {
		ts.kind, ts.ticket, ts.poolPrompt = turnTicket, t, true
		comments, err := store.Comments(ctx, c.store, t.ID)
		if err != nil {
			return nil, err
		}
		body = PromptPart{Source: "ticket", Text: ticketPromptText("", *t, comments)}
	} else {
		ts.poolPrompt = pool.Chat() && delivered.IsZero()
		body = PromptPart{Source: "messages", Text: conversation.EnvelopePlaceholder}
	}
	parts := c.assemble(ctx, ts, body)
	var sent map[string]string
	if r != nil {
		r.mu.Lock()
		sent = r.injected
		r.mu.Unlock()
	}
	send, all, _ := dedupe(c.config(), sent, parts)
	res := rpc.SessionPromptResult{
		Session: sess.ID, Turn: ts.kind, Rules: ruleNames(send), Parts: all, Text: renderParts(all),
	}
	if res.Rules == nil {
		res.Rules = []string{}
	}
	if t != nil {
		res.Ticket = t.ID
	}
	return res, nil
}

// ruleMemory renders a memory rule for the turn's repo and pool (#266).
// An empty or missing dir adds nothing; files over the cap are cut and
// logged at WARN so `jobshop memory check` finds them.
func (c *Controller) ruleMemory(r config.PromptRule, subj config.RuleSubject) string {
	dir := r.Memory.Dir
	if dir == "" {
		return ""
	}
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(c.root.Dir, dir)
	}
	out, over, err := memory.Render(dir, memory.Scope{Repo: subj.Repo, Pool: subj.Pool}, r.Memory.Cap)
	if err != nil {
		c.log.Warn("prompt rule memory", "rule", r.Name, "dir", dir, "err", err)
		return ""
	}
	for _, f := range over {
		c.log.Warn("memory file over cap, truncated", "rule", r.Name, "file", f.Path, "chars", f.Chars, "cap", r.Memory.Cap)
	}
	return out
}
