package config

import (
	"errors"
	"fmt"
	"time"
)

// DeskCommand is one desk_commands entry (#542): a command a Desk plugin
// runs by name with api.exec. Command runs with sh -c in the root dir,
// gets the plugin's input as JSON on stdin, and its stdout is the result.
// Handler is a Starlark file.star:function called as fn(input, args)
// whose return value is the result. A failure or a timeout is an error.
type DeskCommand struct {
	Command string         `yaml:"command,omitempty" json:"command,omitempty"`
	Handler string         `yaml:"handler,omitempty" json:"handler,omitempty"`
	Args    map[string]any `yaml:"args,omitempty" json:"args,omitempty"`
	Timeout string         `yaml:"timeout,omitempty" json:"timeout,omitempty"` // Go duration, default 30s
}

func checkDeskCommand(d DeskCommand) error {
	if (d.Command == "") == (d.Handler == "") {
		return errors.New("set exactly one of command and handler")
	}
	if len(d.Args) > 0 && d.Handler == "" {
		return errors.New("args needs handler")
	}
	if d.Timeout != "" {
		if t, err := time.ParseDuration(d.Timeout); err != nil || t <= 0 {
			return fmt.Errorf("timeout %q must be a positive duration", d.Timeout)
		}
	}
	return nil
}
