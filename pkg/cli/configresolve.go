package cli

import (
	"encoding/json"
	"fmt"

	"github.com/goccy/go-yaml"

	"github.com/elee1766/jobshop/pkg/config"
)

type configResolveCmd struct {
	File    string   `help:"Config file. Default: the root's jobshop.yaml." type:"path"`
	Format  string   `help:"Output format." enum:"json,yaml" default:"json"`
	Replace []string `help:"OLD=NEW: read packs from a different checkout." placeholder:"OLD=NEW"`
}

func (c configResolveCmd) Run(g *Globals) error {
	file := c.File
	if file == "" {
		r, err := g.root()
		if err != nil {
			return err
		}
		file = r.Config()
	}
	rewrite, err := pathRewrite(c.Replace)
	if err != nil {
		return err
	}
	cfg, err := config.LoadFileWith(file, rewrite)
	if err != nil {
		return err
	}
	// Composition metadata differs between formats, not runtime behavior.
	cfg.Pack = ""
	cfg.Packs = nil
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	var resolved map[string]any
	if err := yaml.Unmarshal(b, &resolved); err != nil {
		return err
	}
	delete(resolved, "pack")
	delete(resolved, "packs")
	if c.Format == "yaml" {
		b, err := yaml.Marshal(resolved)
		if err != nil {
			return err
		}
		_, err = stdout.Write(b)
		return err
	}
	if c.Format != "json" {
		return fmt.Errorf("unsupported config output format %q", c.Format)
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(resolved)
}
