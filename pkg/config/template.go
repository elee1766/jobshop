package config

import (
	"github.com/gfx-labs/jobshop/pkg/pack"
	"github.com/gfx-labs/jobshop/pkg/pack/tmpl"
)

// Config values such as step titles, check commands and escalation prompts
// are Go templates (#383) rendered with the language pack loading uses. The
// helpers below are the only template entry points config keeps.

// Fill renders the Go template s from vars. A missing var is an error.
func Fill(s string, vars map[string]any) (string, error) {
	return tmpl.Render(s, vars)
}

// refs returns the var names s references.
func refs(s string) ([]string, error) {
	return tmpl.Vars(s)
}

// placeholder is the reference to var name, as it reads in an error.
func placeholder(name string) string {
	return pack.Placeholder(name)
}
