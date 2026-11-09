package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/knowledge/memory"
	"github.com/elee1766/jobshop/pkg/rpc"
)

type memoryCmd struct {
	Write memoryWriteCmd `cmd:"" help:"Write a local memory file under <root>/memory and record the edit (#592)."`
	Show  memoryShowCmd  `cmd:"" help:"Print the local memory files."`
	Log   memoryLogCmd   `cmd:"" help:"List memory edits with their evaluation, newest first."`
	Check memoryCheckCmd `cmd:"" help:"Report memory files over the character cap (exits 1 if any)."`
}

// memoryWriteCmd is `jobshop memory write <scope> --learnings <ids>`:
// the curator's write path. The controller writes the file, so the edit
// and its previous text are recorded for the gate.
type memoryWriteCmd struct {
	Scope     string  `arg:"" help:"global, repo/<owner>__<name>, pool/<name> or skills/<name>."`
	Learnings []int64 `help:"Learning ids the edit addresses (comma-separated)."`
	FromFile  string  `name:"from-file" help:"Read the new file text from this path. Default: stdin."`
}

func (c memoryWriteCmd) Run(g *Globals) error {
	var text string
	if c.FromFile != "" {
		b, err := os.ReadFile(c.FromFile)
		if err != nil {
			return err
		}
		text = string(b)
	} else {
		t, err := readBody("-")
		if err != nil {
			return err
		}
		text = t + "\n"
	}
	return g.call(rpc.MethodMemoryWrite, rpc.MemoryWriteParams{Scope: c.Scope, Text: text, Learnings: c.Learnings}, nil)
}

type memoryShowCmd struct {
	Scope string `arg:"" optional:"" help:"Only this scope (global, repo/<owner>__<name>, pool/<name>, skills/<name>)."`
}

func (c memoryShowCmd) Run(g *Globals) error {
	return g.call(rpc.MethodMemoryShow, rpc.MemoryShowParams{Scope: c.Scope}, nil)
}

type memoryLogCmd struct {
	State string `enum:",pending,helped,not_helped" default:"" help:"pending, helped or not_helped."`
	Limit int    `default:"50" help:"Most edits to print."`
}

func (c memoryLogCmd) Run(g *Globals) error {
	return g.call(rpc.MethodLearnEdits, rpc.LearnEditsParams{State: c.State, Limit: c.Limit}, nil)
}

type memoryCheckCmd struct {
	Dir string `arg:"" optional:"" help:"Memory directory (holds memory/ and skills/). Default: every prompt rule's memory.dir in jobshop.yaml, e.g. <root>/memory."`
	Cap int    `help:"Per-file character cap. Default: the rule's cap, else 2200."`
}

// Run reads files only; it needs no controller.
func (c memoryCheckCmd) Run(g *Globals) error {
	type target struct {
		dir string
		cap int
	}
	var targets []target
	if c.Dir != "" {
		targets = append(targets, target{c.Dir, c.Cap})
	} else {
		r, err := g.root()
		if err != nil {
			return err
		}
		cfg, err := config.Load(r)
		if err != nil {
			return err
		}
		for _, rule := range cfg.PromptRules {
			if rule.Memory == nil || rule.Memory.Dir == "" {
				continue
			}
			d := rule.Memory.Dir
			if !filepath.IsAbs(d) {
				d = filepath.Join(r.Dir, d)
			}
			cp := rule.Memory.Cap
			if c.Cap > 0 {
				cp = c.Cap
			}
			targets = append(targets, target{d, cp})
		}
		if len(targets) == 0 {
			return errors.New("no prompt rule sets memory.dir; pass a directory")
		}
	}
	n := 0
	for _, t := range targets {
		over, err := memory.Check(t.dir, t.cap)
		if err != nil {
			return err
		}
		lim := t.cap
		if lim <= 0 {
			lim = memory.DefaultCap
		}
		for _, f := range over {
			fmt.Printf("%s: %d chars, cap %d\n", filepath.Join(t.dir, f.Path), f.Chars, lim)
		}
		n += len(over)
	}
	if n > 0 {
		return fmt.Errorf("%d memory file(s) over the cap", n)
	}
	return nil
}
