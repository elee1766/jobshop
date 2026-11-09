// Package chatlife owns the rules of managed conversations: Desk named
// chats, escalation chats (#379) and attention chats (#535, #606). It was
// split out of pkg/control by JOB-147.
//
// A managed chat is a session ticket in a chat pool whose key says what
// it is for: desk:<id> (named, started from Desk), escalation:<ticket>
// (the chat of a ticket on hold) or attention:<kind>:<id> (the chat that
// resolves an attention item). This package decides what each kind may
// do (new, rename, convert, archive, resolve, restore), which key and
// prompt a chat gets, which methods and tools a read-only chat may use,
// and which attention chats to end. It reads tickets through store.Store
// or from values the caller passes, and it never writes.
//
// The controller keeps every write and everything that needs a live
// session: starting, resuming and stopping runners, sending messages,
// closing the ticket an escalation chat answers for, the revision-checked
// session and meta updates, the supervisor pass and the grants. Those
// paths are listed in docs/notes/JOB-147-attention-chat.md.
package chatlife

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	acp "github.com/coder/acp-go-sdk"

	"github.com/elee1766/jobshop/pkg/control/attention"
	"github.com/elee1766/jobshop/pkg/control/question"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Key prefixes of managed chats.
const (
	DeskPrefix       = "desk:"
	EscalationPrefix = "escalation:"
	AttentionPrefix  = attention.KeyPrefix
)

// MaxDisplayName caps a chat's display name, in runes.
const MaxDisplayName = 120

// MaxText bounds a first message and its context, in bytes.
const MaxText = 32 << 10

// chat.manage actions.
const (
	ActionNew     = "new"
	ActionRename  = "rename"
	ActionConvert = "convert"
	ActionArchive = "archive"
	ActionResolve = "resolve"
	ActionRestore = "restore"
)

// Managed reports whether key is a managed chat's key.
func Managed(key string) bool {
	return strings.HasPrefix(key, DeskPrefix) || strings.HasPrefix(key, EscalationPrefix) || strings.HasPrefix(key, AttentionPrefix)
}

// Named reports whether f is a chat Desk lists by name: started from
// Desk, or converted from an escalation chat.
func Named(f store.SessionFields) bool {
	return f.Named || strings.HasPrefix(f.Key, DeskPrefix)
}

// EscalationTicket is the ticket an escalation chat answers for.
func EscalationTicket(f store.SessionFields) (string, bool) {
	return strings.CutPrefix(f.Key, EscalationPrefix)
}

// NeedsName reports whether action takes a chat name.
func NeedsName(action string) bool {
	return action == ActionNew || action == ActionRename || action == ActionConvert
}

// NormalizeName collapses whitespace and checks the length.
func NormalizeName(name string) (string, error) {
	name = strings.Join(strings.Fields(name), " ")
	if name == "" || utf8.RuneCountInString(name) > MaxDisplayName {
		return "", fmt.Errorf("%w: give a chat name of 1 to %d characters", store.ErrInvalid, MaxDisplayName)
	}
	return name, nil
}

// CheckNameFree fails when another named chat in sessions (every session
// ticket) already uses name, ignoring case. self is the chat being
// renamed or converted.
func CheckNameFree(sessions []store.Ticket, self, name string) error {
	for _, t := range sessions {
		f, _ := store.DecodeFields[store.SessionFields](t)
		if t.ID != self && Named(f) && strings.EqualFold(f.DisplayName, name) {
			return fmt.Errorf("%w: chat name %q is already used", store.ErrConflict, name)
		}
	}
	return nil
}

// CheckNew checks the first message of a new chat. Attachments are
// checked by the caller, which owns their storage.
func CheckNew(p rpc.ChatManageParams) error {
	if len(p.Context) > MaxText {
		return fmt.Errorf("%w: context is too long", store.ErrInvalid)
	}
	if (strings.TrimSpace(p.Text) == "" && len(p.Attachments) == 0) || len(p.Text) > MaxText {
		return fmt.Errorf("%w: first message must be 1 to %d bytes", store.ErrInvalid, MaxText)
	}
	return nil
}

// Target decodes the session t that action targets and checks that it is
// a managed chat.
func Target(t store.Ticket) (store.SessionFields, error) {
	f, err := store.DecodeFields[store.SessionFields](t)
	if err != nil || t.Kind != store.KindSession || !Managed(f.Key) {
		return f, fmt.Errorf("%w: not a managed chat", store.ErrInvalid)
	}
	return f, nil
}

// CheckConvert fails unless t is an open escalation chat, the only kind that converts to a named chat.
func CheckConvert(t store.Ticket, f store.SessionFields) error {
	if !strings.HasPrefix(f.Key, EscalationPrefix) || t.Status == store.StatusClosed {
		return fmt.Errorf("%w: convert requires an escalation chat", store.ErrInvalid)
	}
	return nil
}

// Convert makes the escalation chat a named chat with an independent
// lifecycle. The key stays, so pending messages and source references
// remain valid.
func Convert(f *store.SessionFields, name string) {
	f.Named, f.DisplayName, f.Prestart, f.ReadOnly = true, name, false, false
}

// ResolveTarget is the ticket that resolving the escalation chat f
// closes. Only an unconverted escalation chat resolves.
func ResolveTarget(f store.SessionFields) (string, error) {
	id, ok := EscalationTicket(f)
	if !ok || f.Named {
		return "", fmt.Errorf("%w: resolve requires an escalation chat", store.ErrInvalid)
	}
	return id, nil
}

// CheckCurrent fails when chat is no longer the session of target's
// escalation, so resolving it would close work another chat answers for.
func CheckCurrent(target store.Ticket, chat string) error {
	if e, ok := EscalationOf(target); !ok || e.Session != chat {
		return fmt.Errorf("%w: this chat is no longer the escalation's current session", store.ErrConflict)
	}
	return nil
}

// Restore reopens f for a send: the ACP conversation id and journal are
// kept and the session loads on the next message.
func Restore(f *store.SessionFields) {
	f.Archived, f.Ended, f.Suspended = nil, nil, true
}

// View commands (desk.command) only steer the view of the caller's own
// chat.
var viewArgs = map[string][]string{
	"navigate": {"view", "ref"}, "filter": {"window", "repo", "step", "waiting_on"},
	"select": {"refs"}, "attach": {"refs"}, "expand": {"window"}, "collapse": {"window"},
}

// CheckView checks the action and arguments of a view command.
func CheckView(p rpc.DeskCommandParams) error {
	keys, ok := viewArgs[p.Action]
	if !ok {
		return fmt.Errorf("%w: unknown view action %q", store.ErrInvalid, p.Action)
	}
	for k, v := range p.Args {
		if !slices.Contains(keys, k) || len(v) > 2048 {
			return fmt.Errorf("%w: invalid %s argument %q", store.ErrInvalid, p.Action, k)
		}
	}
	if p.Action == "navigate" && !slices.Contains([]string{"home", "trace", "improvements", "system"}, p.Args["view"]) {
		return fmt.Errorf("%w: unknown view", store.ErrInvalid)
	}
	if slices.Contains([]string{"filter", "expand", "collapse"}, p.Action) && !slices.Contains([]string{"waiting", "not-moving", "inflight"}, strings.TrimPrefix(p.Args["window"], "/")) {
		return fmt.Errorf("%w: unknown window", store.ErrInvalid)
	}
	if p.Action == "filter" && strings.TrimPrefix(p.Args["window"], "/") != "inflight" {
		return fmt.Errorf("%w: filters currently apply to inflight only", store.ErrInvalid)
	}
	if slices.Contains([]string{"select", "attach"}, p.Action) && strings.TrimSpace(p.Args["refs"]) == "" {
		return fmt.Errorf("%w: refs are required", store.ErrInvalid)
	}
	return nil
}

// CheckViewTarget fails unless f is a named chat, the only target of a view command.
func CheckViewTarget(f store.SessionFields) error {
	if !Named(f) {
		return fmt.Errorf("%w: desk commands require a named chat", store.ErrInvalid)
	}
	return nil
}

// ReadOnlyMethods are what a read-only session may call (#606).
var ReadOnlyMethods = []string{
	rpc.MethodWhoami, rpc.MethodVersion, rpc.MethodTicketShow, rpc.MethodTicketTrace, rpc.MethodTicketGraph, rpc.MethodTicketTimeline, rpc.MethodTicketList,
	rpc.MethodReady, rpc.MethodRecall, rpc.MethodRunShow, rpc.MethodRunList, rpc.MethodBatchShow,
	rpc.MethodBatchList, rpc.MethodAttention, rpc.MethodConfigShow, rpc.MethodPoolList, rpc.MethodSessionList,
	rpc.MethodSessionJournal, rpc.MethodJournalSearch, rpc.MethodTriggerList, rpc.MethodActivityList,
	rpc.MethodLearnList, rpc.MethodMemoryShow, rpc.MethodReportIncidents, rpc.MethodLinearIntake,
	// Naming itself changes only its label (#626).
	rpc.MethodSessionRename,
}

// ReadOnlyMethod reports whether a read-only session may call method.
func ReadOnlyMethod(method string) bool { return slices.Contains(ReadOnlyMethods, method) }

// ReadOnlyTool reports whether a read-only session may run a tool call of kind k. Execute
// stays allowed so the agent can run the jobshop CLI; its writes are
// refused at the session socket. Edit, delete and move are refused.
func ReadOnlyTool(k *acp.ToolKind) bool {
	if k == nil {
		return true
	}
	return !slices.Contains([]acp.ToolKind{acp.ToolKindEdit, acp.ToolKindDelete, acp.ToolKindMove}, *k)
}

// QuestionChatPrefix starts the key of a question's attention chat.
const QuestionChatPrefix = AttentionPrefix + rpc.AttentionQuestion + ":"

// AttentionPrompt is the context of a resolving chat's first turn. t is
// the item's ticket (zero for an item that is not one), run its result
// when it is a run, actions its choices.
type AttentionPrompt struct {
	Item    rpc.AttentionItem
	Why     string
	Ticket  store.Ticket
	Run     *rpc.RunResult
	Actions []attention.Action
}

// maxRunJSON bounds the run result in a prompt.
const maxRunJSON = 16 << 10

// Text renders the prompt.
func (p AttentionPrompt) Text() string {
	it, t := p.Item, p.Ticket
	var b strings.Builder
	fmt.Fprintf(&b, "Desk attention item: %s %s\n", it.Kind, it.ID)
	title := it.Title
	if title == "" {
		title = t.Title
	}
	if title != "" {
		fmt.Fprintf(&b, "Title: %s\n", title)
	}
	if p.Why != "" && p.Why != it.Kind {
		fmt.Fprintf(&b, "Why: %s\n", p.Why)
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
	if it.Kind == rpc.AttentionRun && t.Kind == store.KindRun && p.Run != nil {
		out, _ := json.MarshalIndent(p.Run, "", "  ")
		if len(out) > maxRunJSON {
			out = append(out[:maxRunJSON], "\n[cut]"...)
		}
		fmt.Fprintf(&b, "\n`jobshop run show %s`:\n```json\n%s\n```\n", it.ID, out)
	}
	b.WriteString("\nThe operator opens this chat from Desk. Read the item with the jobshop CLI (read-only commands only: this session cannot write until the operator sends a message), then reply in a few lines: what blocks it, and the concrete actions available, as Desk's resolve buttons (approve, reject or dismiss for a decision; close, resume, unquarantine, clear wait) and the exact CLI commands. Do not act until the operator agrees.")
	for _, action := range p.Actions {
		fmt.Fprintf(&b, "\nOption %d: %s. %s", action.Option.ID, action.Option.Title, action.Option.Description)
	}
	b.WriteString("\nIf one option is supported by the evidence, end with a line exactly 'Recommended option: N', replacing N with its number. Otherwise do not recommend one.")
	if it.Kind == rpc.AttentionQuestion {
		qm, _ := question.Of(t)
		b.WriteString(question.ChatPrompt(it.ID, qm))
	}
	return b.String()
}

// Summary is what a read-only attention chat's first reply records on
// its item (#607): the reply as context and row summary, and the option
// it recommended, "" when none.
type Summary struct {
	Context, Row, Recommend string
}

// Summarize reads a chat reply. Kind and id name the item from the
// chat's key; ok is false when the reply is empty or the item is not a
// ticket, so nothing is recorded.
func Summarize(key, said string) (id string, s Summary, ok bool) {
	said = strings.TrimSpace(said)
	if said == "" {
		return "", s, false
	}
	kind, id, isAttention := attention.ParseKey(key)
	if !isAttention || !slices.Contains(attention.TicketKinds, kind) {
		return "", s, false
	}
	for _, line := range strings.Split(said, "\n") {
		if value, found := strings.CutPrefix(strings.TrimSpace(line), "Recommended option: "); found {
			if n, err := strconv.Atoi(value); err == nil && n >= 1 && n <= 5 {
				s.Recommend = value
			}
		}
	}
	// Keep authored Markdown intact. The row preview may clip a link destination.
	s.Context, s.Row = said, attention.SummaryOf(said)
	return id, s, true
}

// Left is an attention chat whose item left attention.
type Left struct {
	Session store.Ticket
	// Item is kind:id, for the log.
	Item string
}

// Sweep reads the open sessions against the listed items. It returns the
// attention chats whose item left attention, which the controller closes,
// and how many of the rest were pre-started. A ticket listed under
// another kind (a question that became a wait) still counts as in
// attention. followed says an answered question's chat is not settled yet:
// it follows the asker (JOB-52).
func Sweep(items []rpc.AttentionItem, open []store.Ticket, followed func(store.Ticket, store.SessionFields) bool) (left []Left, prestarted int) {
	listed := map[string]bool{}
	for _, it := range items {
		listed[it.Kind+":"+it.ID] = true
		if slices.Contains(attention.TicketKinds, it.Kind) {
			listed["ticket:"+it.ID] = true
		}
	}
	for _, s := range open {
		f, err := store.DecodeFields[store.SessionFields](s)
		if err != nil {
			continue
		}
		kind, id, ok := attention.ParseKey(f.Key)
		if !ok {
			continue
		}
		inList := listed[kind+":"+id] || (slices.Contains(attention.TicketKinds, kind) && listed["ticket:"+id])
		if !inList && kind == rpc.AttentionQuestion && followed(s, f) {
			inList = true
		}
		if !inList {
			left = append(left, Left{Session: s, Item: kind + ":" + id})
			continue
		}
		if f.Prestart {
			prestarted++
		}
	}
	return left, prestarted
}
