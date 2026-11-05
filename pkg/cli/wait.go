package cli

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// wait (#381): park a ticket until outside state holds.
type waitCmd struct {
	Add   waitAddCmd   `cmd:"" help:"Park a ticket until conditions hold; a held ticket is released (not a failed attempt). End the turn after."`
	List  waitListCmd  `cmd:"" help:"List waiting tickets."`
	Clear waitClearCmd `cmd:"" help:"Remove a ticket's wait; it becomes ready."`
}

type waitAddCmd struct {
	Conds   []string `arg:"" name:"cond" help:"TYPE:key=val,...  Types: github.review (pr, login, head, state), github.checks (pr or repo+sha, state=success|failure, required), github.pr (pr, state=merged|closed|head_changed), github.label (issue, label; met when the issue has the label or is closed), github.tag (repo, tag), deploy (repo, sha; met when the controller's build commit contains sha), run (id, state=done|closed), linear.state (issue=KEY, state name or type), linear.label (issue=KEY, label, present=false waits for removal), thread.reply (thread, from), time (at). pr is owner/name#n or a PR URL; at is a duration or RFC3339."`
	Reason  string   `short:"r" required:"" help:"Why the ticket waits."`
	Owner   string   `help:"Who must act (a login, team or person)."`
	Any     bool     `help:"Clear when any condition holds (default: all)."`
	Resume  bool     `help:"Give the ticket back to this session when the wait clears."`
	Timeout string   `help:"Duration or RFC3339 time; past it the wait is labeled needs-human and listed in attention."`
	ID      string   `short:"t" name:"ticket" help:"Ticket id (default: the session's current ticket)."`
}

func (c waitAddCmd) Run(g *Globals) error {
	now := time.Now()
	p := rpc.WaitAddParams{ID: c.ID, Reason: c.Reason, Owner: c.Owner, Any: c.Any, Resume: c.Resume}
	for _, s := range c.Conds {
		cd, err := parseWaitCond(s, now)
		if err != nil {
			return err
		}
		p.Conds = append(p.Conds, cd)
	}
	if c.Timeout != "" {
		t, err := parseUntil(c.Timeout, now)
		if err != nil {
			return err
		}
		p.Timeout = &t
	}
	return g.call(rpc.MethodWaitAdd, p, nil)
}

// linearKeyRe is a Linear issue key (ENG-12).
var linearKeyRe = regexp.MustCompile(`^[A-Z][A-Z0-9]*-\d+$`)

// parseWaitCond reads TYPE:key=val,key=val.
func parseWaitCond(s string, now time.Time) (store.WaitCond, error) {
	typ, rest, _ := strings.Cut(s, ":")
	cd := store.WaitCond{Type: typ}
	if rest == "" {
		return cd, fmt.Errorf("condition %q: want TYPE:key=val[,key=val]", s)
	}
	for kv := range strings.SplitSeq(rest, ",") {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			return cd, fmt.Errorf("condition %q: %q is not key=val", s, kv)
		}
		var err error
		switch k {
		case "pr":
			var pr bool
			cd.Repo, cd.PR, pr, err = parseThread(v)
			if err == nil && strings.Contains(v, "/issues/") && !pr {
				err = fmt.Errorf("%s is an issue, not a pull request", v)
			}
		case "repo":
			cd.Repo = v
		case "issue":
			if linearKeyRe.MatchString(v) {
				cd.Thread = v // linear.state, linear.label
				break
			}
			cd.Repo, cd.Issue, _, err = parseThread(v)
		case "tag", "release":
			cd.Tag = v
		case "id":
			cd.Run = v
		case "present":
			var present bool
			present, err = strconv.ParseBool(v)
			cd.Absent = !present
		case "label":
			cd.Label = v
		case "head":
			cd.Head = v
		case "sha":
			cd.SHA = v
		case "login":
			cd.Login = v
		case "state":
			cd.State = v
		case "required":
			cd.Required, err = strconv.ParseBool(v)
		case "thread":
			cd.Thread = v
			if repo, n, _, terr := parseThread(v); terr == nil {
				cd.Thread = fmt.Sprintf("%s#%d", repo, n)
			}
		case "from":
			cd.From = v
		case "at":
			cd.At, err = parseUntil(v, now)
		default:
			err = fmt.Errorf("unknown key %q", k)
		}
		if err != nil {
			return cd, fmt.Errorf("condition %q: %w", s, err)
		}
	}
	return cd, nil
}

type waitListCmd struct {
	Project string `short:"p" help:"Project (admin; default all)."`
}

func (c waitListCmd) Run(g *Globals) error {
	return g.call(rpc.MethodWaitList, rpc.WaitListParams{Project: c.Project}, nil)
}

type waitClearCmd struct {
	ID     string `arg:"" help:"Ticket id."`
	Reason string `short:"r" help:"Why the wait is cleared."`
}

func (c waitClearCmd) Run(g *Globals) error {
	return g.call(rpc.MethodWaitClear, rpc.WaitClearParams{ID: c.ID, Reason: c.Reason}, nil)
}
