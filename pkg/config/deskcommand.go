package config

import (
	"errors"
	"fmt"
	"time"
)

// DeskCommand is one desk_commands entry (#542): a command a Desk plugin
// runs by name with api.exec. It runs with sh -c in the root dir, gets
// the plugin's input as JSON on stdin, and its stdout is the result. A
// non-zero exit or a timeout is an error.
type DeskCommand struct {
	Command string `yaml:"command" json:"command"`
	Timeout string `yaml:"timeout,omitempty" json:"timeout,omitempty"` // Go duration, default 30s
}

func checkDeskCommand(d DeskCommand) error {
	if d.Command == "" {
		return errors.New("command is required")
	}
	if d.Timeout != "" {
		if t, err := time.ParseDuration(d.Timeout); err != nil || t <= 0 {
			return fmt.Errorf("timeout %q must be a positive duration", d.Timeout)
		}
	}
	return nil
}
