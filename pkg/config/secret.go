package config

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// SecretCommandTimeout bounds a secret command, so a stuck credential
// helper cannot hang a webhook or a token mint.
const SecretCommandTimeout = 30 * time.Second

// IsSecretCommand reports whether a *_file value is a command: a leading
// "!" (as in git's credential helpers) means run the rest with sh -c and
// use its stdout, so secrets can come from pass, op, vault, systemd-creds
// and the like instead of a plain file.
func IsSecretCommand(v string) bool { return strings.HasPrefix(v, "!") }

// SecretPath resolves a root-relative *_file value; commands pass through.
func (r Root) SecretPath(v string) string {
	if IsSecretCommand(v) {
		return v
	}
	return r.Path(v)
}

// ReadSecret returns the bytes behind a *_file value: the file's contents,
// or a secret command's stdout. Callers trim as they need.
func ReadSecret(ctx context.Context, v string) ([]byte, error) {
	if !IsSecretCommand(v) {
		return os.ReadFile(v)
	}
	ctx, cancel := context.WithTimeout(ctx, SecretCommandTimeout)
	defer cancel()
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "sh", "-c", v[1:])
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		// Only stderr's last line: the command, and the secret, stay out of logs.
		msg := strings.TrimSpace(stderr.String())
		if i := strings.LastIndexByte(msg, '\n'); i >= 0 {
			msg = msg[i+1:]
		}
		return nil, fmt.Errorf("secret command: %w: %s", err, msg)
	}
	return out, nil
}
