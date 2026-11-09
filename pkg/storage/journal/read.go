// Reading session journals for the controller's session.journal,
// journal.search and recall paths (split out of pkg/control in JOB-140,
// moved here in JOB-173). The controller decides who may read what
// (grants, project scope, canReadJournal) and hands in only the result of
// that decision, an Access for one session or a project filter for a
// search. Nothing here imports pkg/control, looks at a scope or reads
// config.
//
// Journal files are read through acphost's readers, which serve the
// plain journal.ndjson and the compressed journal.ndjson.zst (#631) the
// same way, so paging by seq works on both.

package journal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"time"

	host "github.com/elee1766/acphost/pkg/acphost"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

const (
	// PageDefaultLimit is a session.journal page's size when none is asked.
	PageDefaultLimit = 500
	// PageMaxLimit caps a session.journal page.
	PageMaxLimit = 5000

	// SearchDefaultLimit is journal.search's row limit when none is asked.
	SearchDefaultLimit = 1000
	// SearchMaxLimit caps journal.search rows.
	SearchMaxLimit = 10000

	// RecallBytes bounds the journal text indexed per session. Summarize
	// keeps the first prompt and the newest items when a journal is longer.
	RecallBytes = 64 << 10
)

// Redactor redacts one journal message. The controller builds it from
// its rules and the session's env.
type Redactor interface {
	JSON(json.RawMessage) json.RawMessage
}

// Access is read access to one session's journal that the controller has
// already authorized.
type Access struct {
	// Dir is the session dir holding journal.ndjson or its .zst.
	Dir string
	// Redact is applied to every frame returned.
	Redact Redactor
}

// ValidatePage checks the caller's paging parameters.
func ValidatePage(p rpc.SessionJournalParams) error {
	if p.From != "" && p.From != "agent" && p.From != "controller" {
		return fmt.Errorf("%w: from must be agent or controller", domain.ErrInvalid)
	}
	if p.SinceSeq < 0 {
		return fmt.Errorf("%w: since_seq must be >= 0", domain.ErrInvalid)
	}
	return nil
}

// Page returns a page of redacted frames after p.SinceSeq. A session with
// no journal file reads as empty and done.
func Page(a Access, p rpc.SessionJournalParams) (rpc.SessionJournalResult, error) {
	limit := p.Limit
	if limit <= 0 {
		limit = PageDefaultLimit
	}
	limit = min(limit, PageMaxLimit)
	frames, _, err := host.ReadJournal(filepath.Join(a.Dir, host.JournalFile), p.SinceSeq)
	if errors.Is(err, fs.ErrNotExist) {
		frames, err = nil, nil
	}
	if err != nil {
		return rpc.SessionJournalResult{}, err
	}
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
		res.Frames = append(res.Frames, rpc.JournalFrame{Seq: f.Seq, T: f.T, From: f.From, Msg: a.Redact.JSON(f.Msg)})
	}
	return res, nil
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

// Summary is the text indexed for recall (#265) from the session's
// journal, and the time of its newest frame. ok is false when there is
// nothing to index: no journal, an unreadable one, or no frames.
func Summary(dir string) (text string, at time.Time, ok bool) {
	frames, _, err := host.ReadJournal(filepath.Join(dir, host.JournalFile), 0)
	if err != nil || len(frames) == 0 {
		return "", time.Time{}, false
	}
	return host.Summarize(frames, RecallBytes), time.UnixMilli(frames[len(frames)-1].T), true
}

// Search is one journal.search request, already authorized.
type Search struct {
	Client *Client
	// Project restricts the search to one project's streams. Empty is
	// every project (admin).
	Project string
	// Now anchors Params.Since.
	Now time.Time
	rpc.JournalSearchParams
}

// errStop ends a query once the limit is reached.
var errStop = errors.New("limit reached")

// RunSearch runs a LogsQL query on VictoriaLogs (#530). Project scoping is
// VictoriaLogs' extra_stream_filters, outside the query text.
func RunSearch(ctx context.Context, s Search) (rpc.JournalSearchResult, error) {
	if strings.TrimSpace(s.Query) == "" {
		return rpc.JournalSearchResult{}, fmt.Errorf("%w: query is required (LogsQL, e.g. '*')", domain.ErrInvalid)
	}
	limit := s.Limit
	if limit <= 0 {
		limit = SearchDefaultLimit
	}
	limit = min(limit, SearchMaxLimit)
	// One more than asked, to tell a full result from a truncated one.
	q := Query{LogsQL: s.Query, Start: s.Start, End: s.Until, Limit: limit + 1, Project: s.Project}
	if q.Start.IsZero() && s.Since > 0 {
		q.Start = s.Now.Add(-s.Since)
	}
	res := rpc.JournalSearchResult{Rows: []json.RawMessage{}}
	err := s.Client.Run(ctx, q, func(row json.RawMessage) error {
		if len(res.Rows) == limit {
			res.Truncated = true
			return errStop
		}
		res.Rows = append(res.Rows, row)
		return nil
	})
	if errors.Is(err, errStop) {
		err = nil
	}
	var se *StatusError
	if errors.As(err, &se) && se.Code == 400 {
		return rpc.JournalSearchResult{}, fmt.Errorf("%w: %s", domain.ErrInvalid, se.Body)
	}
	if err != nil {
		return rpc.JournalSearchResult{}, err
	}
	return res, nil
}
