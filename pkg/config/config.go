// Package config loads jobshop.yaml and locates the Jobshop root directory.
package config

import (
	"bytes"
	"fmt"
	"os"
	"regexp"

	"github.com/goccy/go-yaml"
)

// Config is the parsed jobshop.yaml. Phase 1 needs only projects; agents,
// pools and telemetry are added as their features land.
type Config struct {
	// DefaultProject is used when a request names no project.
	DefaultProject string             `yaml:"default_project"`
	Projects       map[string]Project `yaml:"projects"`
	// Lease is the claim lease (Go duration). Default 30m.
	Lease string `yaml:"lease"`
}

// Project is one repo / work area.
type Project struct {
	// Prefix starts every ticket id, e.g. "js" gives js-0031w02j0r012.
	// Default: the project name.
	Prefix string `yaml:"prefix"`
}

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`)

// Load reads and validates the root's config.
func Load(r Root) (*Config, error) {
	b, err := os.ReadFile(r.Config())
	if err != nil {
		return nil, err
	}
	return Parse(b)
}

// Parse decodes and validates config bytes. Unknown keys are rejected.
func Parse(b []byte) (*Config, error) {
	var c Config
	if len(bytes.TrimSpace(b)) > 0 {
		if err := yaml.UnmarshalWithOptions(b, &c, yaml.Strict()); err != nil {
			return nil, fmt.Errorf("%s: %w", FileName, err)
		}
	}
	if c.Projects == nil {
		c.Projects = map[string]Project{}
	}
	seen := map[string]string{}
	for name, p := range c.Projects {
		if !namePattern.MatchString(name) {
			return nil, fmt.Errorf("%s: project name %q must match %s", FileName, name, namePattern)
		}
		if p.Prefix == "" {
			p.Prefix = name
			c.Projects[name] = p
		}
		if !namePattern.MatchString(p.Prefix) {
			return nil, fmt.Errorf("%s: project %s: prefix %q must match %s", FileName, name, p.Prefix, namePattern)
		}
		if other, dup := seen[p.Prefix]; dup {
			return nil, fmt.Errorf("%s: projects %s and %s share prefix %q", FileName, other, name, p.Prefix)
		}
		seen[p.Prefix] = name
	}
	if c.DefaultProject == "" && len(c.Projects) == 1 {
		for name := range c.Projects {
			c.DefaultProject = name
		}
	}
	if c.DefaultProject != "" {
		if _, ok := c.Projects[c.DefaultProject]; !ok {
			return nil, fmt.Errorf("%s: default_project %q is not defined", FileName, c.DefaultProject)
		}
	}
	return &c, nil
}

// Prefix returns the id prefix for a project.
func (c *Config) Prefix(project string) string {
	if p, ok := c.Projects[project]; ok {
		return p.Prefix
	}
	return project
}
