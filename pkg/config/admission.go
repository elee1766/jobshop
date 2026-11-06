package config

import (
	"errors"
	"fmt"
	"path"
	"slices"
	"time"
)

// Admission hook types (#348). Mutating hooks run first, in config
// order, and may rewrite the event; validating hooks then answer allow,
// deny or ask_human on the final event.
const (
	AdmissionMutating   = "mutating"
	AdmissionValidating = "validating"
)

// Admission failure policies. fail (the default) denies on a validating
// error and marks a mutating error failed; ignore skips the hook.
const (
	FailurePolicyFail   = "fail"
	FailurePolicyIgnore = "ignore"
)

// AdmissionHook is one ingest.admission entry (#348).
type AdmissionHook struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
	// Exactly one of Command (sh -c, JSON on stdin) and URL (HTTP POST).
	Command string `yaml:"command"`
	URL     string `yaml:"url"`
	// SecretFile, relative to the root, signs url requests
	// (X-Jobshop-Signature-256). Required with url.
	SecretFile    Secret         `yaml:"secret_file"`
	Retry         HTTPRetry      `yaml:"retry"`   // url only
	Timeout       string         `yaml:"timeout"` // Go duration per try, default 30s
	FailurePolicy string         `yaml:"failure_policy"`
	Match         AdmissionMatch `yaml:"match"`
	// Intake marks an on: intake hook (#288) run as a validating hook;
	// its event keeps on: intake.
	Intake bool `yaml:"-"`
}

// AdmissionMatch limits a hook to events with these values. An empty
// list matches all. Repos and Teams are path.Match globs.
type AdmissionMatch struct {
	Sources  []string `yaml:"sources"`
	Kinds    []string `yaml:"kinds"`
	Repos    []string `yaml:"repos"`
	Teams    []string `yaml:"teams"`
	Actors   []string `yaml:"actors"`
	Projects []string `yaml:"projects"`
	// Bots, when set, matches only bot actors (true) or only humans (false).
	Bots *bool `yaml:"bots"`
}

// AdmissionSubject is what a match is tested against.
type AdmissionSubject struct {
	Source, Kind, RepoOrTeam, Actor, Project string
	Bot                                      bool
}

// Matches reports whether m selects s.
func (m AdmissionMatch) Matches(s AdmissionSubject) bool {
	in := func(list []string, v string) bool { return len(list) == 0 || slices.Contains(list, v) }
	glob := func(list []string, v string) bool {
		if len(list) == 0 {
			return true
		}
		return slices.ContainsFunc(list, func(p string) bool {
			ok, _ := path.Match(p, v)
			return ok
		})
	}
	if !in(m.Sources, s.Source) || !in(m.Kinds, s.Kind) || !in(m.Actors, s.Actor) || !in(m.Projects, s.Project) {
		return false
	}
	// repos only constrain GitHub events and teams only Linear ones, so
	// a hook listing both covers each source by its own field.
	if s.Source == IngestGitHub && !glob(m.Repos, s.RepoOrTeam) {
		return false
	}
	if s.Source == IngestLinear && !glob(m.Teams, s.RepoOrTeam) {
		return false
	}
	return m.Bots == nil || *m.Bots == s.Bot
}

// Policy is FailurePolicy or the default fail.
func (h AdmissionHook) Policy() string {
	if h.FailurePolicy == "" {
		return FailurePolicyFail
	}
	return h.FailurePolicy
}

func (c *Config) validateAdmission() error {
	names := map[string]bool{}
	for _, h := range c.Ingest.Admission {
		where := fmt.Sprintf("%s: ingest.admission %q", FileName, h.Name)
		if err := c.checkAdmission(h, names); err != nil {
			return fmt.Errorf("%s: %w", where, err)
		}
		names[h.Name] = true
	}
	return nil
}

func (c *Config) checkAdmission(h AdmissionHook, names map[string]bool) error {
	if !endpointName.MatchString(h.Name) {
		return errors.New("name must match [a-z0-9-]+")
	}
	if names[h.Name] {
		return errors.New("duplicate name")
	}
	if h.Type != AdmissionMutating && h.Type != AdmissionValidating {
		return fmt.Errorf("type %q must be mutating or validating", h.Type)
	}
	if err := checkHookURL(h.Command, h.URL, h.SecretFile, h.Retry); err != nil {
		return err
	}
	if h.Timeout != "" {
		if d, err := time.ParseDuration(h.Timeout); err != nil || d <= 0 {
			return fmt.Errorf("timeout %q must be a positive duration", h.Timeout)
		}
	}
	if p := h.FailurePolicy; p != "" && p != FailurePolicyFail && p != FailurePolicyIgnore {
		return fmt.Errorf("failure_policy %q must be fail or ignore", p)
	}
	for _, s := range h.Match.Sources {
		if !slices.Contains(IngestSources, s) {
			return fmt.Errorf("match.sources %q must be github, linear or generic", s)
		}
	}
	for _, g := range slices.Concat(h.Match.Repos, h.Match.Teams) {
		if _, err := path.Match(g, ""); err != nil {
			return fmt.Errorf("match glob %q: %w", g, err)
		}
	}
	for _, p := range h.Match.Projects {
		if _, ok := c.Projects[p]; !ok {
			return fmt.Errorf("match.projects: unknown project %q", p)
		}
	}
	return nil
}
