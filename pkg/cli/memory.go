package cli

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/memory"
)

type memoryCmd struct {
	Check memoryCheckCmd `cmd:"" help:"Report memory files over the character cap (exits 1 if any)."`
}

type memoryCheckCmd struct {
	Dir string `arg:"" optional:"" help:"Memory directory (holds memory/ and skills/). Default: every prompt rule's memory.dir in jobshop.yaml."`
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
