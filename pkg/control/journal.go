package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	host "github.com/elee1766/acphost/pkg/acphost"

	"github.com/elee1766/jobshop/pkg/hostspec"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// session.journal (#529): read-only, paged, redacted access to a
// session's ACP journal. It reads the host's journal.ndjson through
// acphost's reader, so it works for live and ended sessions while the
// session dir exists. It takes the session id only, so a later
// multi-controller setup can forward it to the owner unchanged.

const (
	journalDefaultLimit = 500
	journalMaxLimit     = 5000
	// grantSessionJournal lets a pool read journals in its project.
	grantSessionJournal = "session-journal"
)

func (c *Controller) sessionJournal(ctx context.Context, sc scope, p rpc.SessionJournalParams) (any, error) {
	if p.ID == "" {
		return nil, fmt.Errorf("%w: session id is required", store.ErrInvalid)
	}
	if p.From != "" && p.From != "agent" && p.From != "controller" {
		return nil, fmt.Errorf("%w: from must be agent or controller", store.ErrInvalid)
	}
	if p.SinceSeq < 0 {
		return nil, fmt.Errorf("%w: since_seq must be >= 0", store.ErrInvalid)
	}
	t, err := c.store.Get(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	if t.Kind != store.KindSession {
		return nil, fmt.Errorf("%w: %s is a %s, not a session", store.ErrInvalid, p.ID, t.Kind)
	}
	if err := c.canReadJournal(sc, t); err != nil {
		return nil, err
	}
	limit := p.Limit
	if limit <= 0 {
		limit = journalDefaultLimit
	}
	limit = min(limit, journalMaxLimit)
	dir := c.root.SessionDir(t.ID)
	frames, _, err := host.ReadJournal(filepath.Join(dir, host.JournalFile), p.SinceSeq)
	if errors.Is(err, fs.ErrNotExist) {
		frames, err = nil, nil
	}
	if err != nil {
		return nil, err
	}
	red := c.redactor(journalEnv(dir))
	res := rpc.SessionJournalResult{Frames: []rpc.JournalFrame{}, NextSeq: p.SinceSeq, Done: true}
	for _, f := range frames {
		if len(res.Frames) == limit {
			res.Done = false
			break
		}
		// Frames that do not match still advance the cursor, so a
		// filtered caller does not rescan them.
		res.NextSeq = f.Seq
		if (p.From != "" && f.From != p.From) || (len(p.Kinds) > 0 && !slices.ContainsFunc(frameKinds(f.Msg), func(k string) bool { return slices.Contains(p.Kinds, k) })) {
			continue
		}
		res.Frames = append(res.Frames, rpc.JournalFrame{Seq: f.Seq, T: f.T, From: f.From, Msg: red.JSON(f.Msg)})
	}
	return res, nil
}

// canReadJournal: admin; a session for itself; a pool with the
// session-journal grant for sessions in its project.
func (c *Controller) canReadJournal(sc scope, t store.Ticket) error {
	if sc.admin || sc.session == t.ID {
		return nil
	}
	if p, ok := c.config().Pools[sc.pool]; ok && p.Allows(grantSessionJournal) {
		if t.Project != sc.project {
			return forbidden("session %s is not in project %s", t.ID, sc.project)
		}
		return nil
	}
	return forbidden("reading another session's journal needs the admin socket or pool admin grant %q", grantSessionJournal)
}

// frameKinds are msg.method and params.update.sessionUpdate, the two
// things callers filter on.
func frameKinds(msg json.RawMessage) []string {
	var m struct {
		Method string `json:"method"`
		Params struct {
			Update struct {
				SessionUpdate string `json:"sessionUpdate"`
			} `json:"update"`
		} `json:"params"`
	}
	if json.Unmarshal(msg, &m) != nil {
		return nil
	}
	return []string{m.Method, m.Params.Update.SessionUpdate}
}

// journalEnv is the env the session's agent ran with, for redacting its
// secret values: the host's inherited env (this controller's) plus the
// spec's.
func journalEnv(dir string) []string {
	env := os.Environ()
	if s, err := hostspec.Read(dir); err == nil {
		env = append(env, s.Env...)
	}
	return env
}
