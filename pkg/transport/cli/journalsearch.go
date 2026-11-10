package cli

import (
	"os"
	"time"

	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// journalCmd is journal search on VictoriaLogs (#530).
type journalCmd struct {
	Search   journalSearchCmd   `cmd:"" help:"Search shipped session journals with LogsQL, e.g. 'status:failed | stats by (session) count()' (admin, or journal-search grant for your project)."`
	Backfill journalBackfillCmd `cmd:"" help:"Ship existing journals to VictoriaLogs from their cursors now (admin)."`
}

type journalSearchCmd struct {
	Query string        `arg:"" help:"LogsQL query. Rows are deduped by (session, seq) with '| uniq by (session, seq)' if a resend made doubles."`
	Since time.Duration `help:"Look back this far, e.g. 2h. Default: all."`
	Until time.Time     `help:"End of the window (RFC3339)."`
	Limit int           `help:"Most rows (default 1000, max 10000)."`
}

func (c journalSearchCmd) Run(g *Globals) error {
	var res rpc.JournalSearchResult
	if err := g.call(rpc.MethodJournalSearch, rpc.JournalSearchParams{Query: c.Query, Since: c.Since, Until: c.Until, Limit: c.Limit}, &res); err != nil {
		return err
	}
	// One JSON row per line, as VictoriaLogs sends them.
	for _, r := range res.Rows {
		if _, err := stdout.Write(append(r, '\n')); err != nil {
			return err
		}
	}
	if res.Truncated {
		_, err := os.Stderr.WriteString("jobshop: more rows; raise --limit or narrow the query\n")
		return err
	}
	return nil
}

type journalBackfillCmd struct {
	Since time.Duration `help:"Only sessions closed this recently, e.g. 168h (open sessions always). Default: all."`
}

func (c journalBackfillCmd) Run(g *Globals) error {
	return g.callTimeout(time.Hour, rpc.MethodJournalBackfill, rpc.JournalBackfillParams{Since: c.Since}, nil)
}
