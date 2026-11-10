package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gfx-labs/jobshop/pkg/tooling/runbook"
)

// runbookCmd is `jobshop runbook` (#635). It runs locally: it reads a
// docs/verify file and never talks to the controller.
type runbookCmd struct {
	Run runbookRunCmd `cmd:"" help:"Run a runbook's read-only steps (fenced blocks tagged readonly) and print the results."`
}

type runbookRunCmd struct {
	File    string        `arg:"" help:"Runbook markdown file, e.g. docs/verify/desk.md."`
	Dir     string        `help:"Working directory for the steps (default: the cwd)."`
	Timeout time.Duration `default:"2m" help:"Timeout per step."`
	JSON    bool          `name:"json" help:"Print the report as JSON instead of markdown."`
}

func (c runbookRunCmd) Run(_ *Globals) error {
	src, err := os.ReadFile(c.File)
	if err != nil {
		return err
	}
	rep := runbook.Run(context.Background(), filepath.ToSlash(c.File), runbook.Parse(src),
		runbook.Options{Dir: c.Dir, Timeout: c.Timeout})
	if c.JSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(rep); err != nil {
			return err
		}
	} else {
		fmt.Print(rep.Markdown())
	}
	if !rep.OK() {
		return fmt.Errorf("runbook %s: a read-only step failed", c.File)
	}
	return nil
}
