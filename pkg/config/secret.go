package config

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
)

// SecretCommandTimeout bounds a secret command, so a stuck credential
// helper cannot hang a webhook or a token mint. A var so tests can
// shorten it.
var SecretCommandTimeout = 30 * time.Second

// Secret types (#435).
const (
	SecretFile    = "file"
	SecretCommand = "command"
)

// Secret is where a *_file setting's secret comes from: a file, or a
// command whose stdout is the secret (pass, op, vault, systemd-creds and
// the like). In yaml it is a path string, as before, or a mapping
//
//	{type: file, path: /run/secrets/x}
//	{type: command, command: "pass show gh/jobshop"}
//
// The type is explicit: no prefix makes a string a command (#435).
type Secret struct {
	Type    string `yaml:"type"`
	Path    string `yaml:"path,omitempty"`
	Command string `yaml:"command,omitempty"`
}

// SecretPath is a file secret at path.
func SecretPath(path string) Secret {
	if path == "" {
		return Secret{}
	}
	return Secret{Type: SecretFile, Path: path}
}

// IsZero reports an unset secret.
func (s Secret) IsZero() bool { return s.Type == "" }

// String names the secret for errors: the path, or "secret command" so a
// command line, which may hold a token, stays out of logs.
func (s Secret) String() string {
	if s.Type == SecretCommand {
		return "secret command"
	}
	return s.Path
}

// UnmarshalYAML takes a path string or a {type, path|command} mapping.
func (s *Secret) UnmarshalYAML(b []byte) error {
	var str string
	if err := yaml.Unmarshal(b, &str); err == nil {
		*s = SecretPath(str)
		return nil
	}
	var m struct {
		Type    string `yaml:"type"`
		Path    string `yaml:"path"`
		Command string `yaml:"command"`
	}
	if err := yaml.UnmarshalWithOptions(b, &m, yaml.Strict()); err != nil {
		return fmt.Errorf("secret: a path, {type: file, path: ...} or {type: command, command: ...}: %w", err)
	}
	switch m.Type {
	case SecretFile:
		if m.Path == "" || m.Command != "" {
			return errors.New("secret: type file needs path and no command")
		}
	case SecretCommand:
		if m.Command == "" || m.Path != "" {
			return errors.New("secret: type command needs command and no path")
		}
	default:
		return fmt.Errorf("secret: type %q must be file or command", m.Type)
	}
	*s = Secret{Type: m.Type, Path: m.Path, Command: m.Command}
	return nil
}

// MarshalYAML writes a file secret as its path string, so configs that
// round-trip (last-good, reload) keep their form.
func (s Secret) MarshalYAML() (any, error) {
	if s.Type == SecretFile || s.Type == "" {
		return s.Path, nil
	}
	return struct {
		Type    string `yaml:"type"`
		Command string `yaml:"command"`
	}{s.Type, s.Command}, nil
}

// In resolves a relative file path against dir; commands and absolute
// paths are unchanged.
func (s Secret) In(dir string) Secret {
	if s.Type == SecretFile && s.Path != "" && !filepath.IsAbs(s.Path) {
		s.Path = filepath.Join(dir, s.Path)
	}
	return s
}

// ReadSecret returns the bytes behind s: the file's contents, or the
// command's stdout. The command runs with sh -c as the jobshop user on
// every read, so rotation needs no restart. Callers trim as they need.
func ReadSecret(ctx context.Context, s Secret) ([]byte, error) {
	switch s.Type {
	case SecretFile:
		return os.ReadFile(s.Path)
	case SecretCommand:
	default:
		return nil, errors.New("secret is not set")
	}
	ctx, cancel := context.WithTimeout(ctx, SecretCommandTimeout)
	defer cancel()
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "sh", "-c", s.Command)
	cmd.Stderr = &stderr
	// sh's children keep stdout open after a kill; do not wait for them.
	cmd.WaitDelay = time.Second
	out, err := cmd.Output()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("secret command: timed out after %s", SecretCommandTimeout)
		}
		// Only stderr's last line: the command, and the secret, stay out of logs.
		msg := strings.TrimSpace(stderr.String())
		if i := strings.LastIndexByte(msg, '\n'); i >= 0 {
			msg = msg[i+1:]
		}
		return nil, fmt.Errorf("secret command: %w: %s", err, msg)
	}
	return out, nil
}
