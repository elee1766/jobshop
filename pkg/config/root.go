package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// FileName is the config file that marks a Jobshop root directory.
const FileName = "jobshop.yaml"

// Root is a Jobshop root directory. Everything the server owns lives under
// it, so one directory is one deployment.
type Root struct{ Dir string }

// Config returns the config file path.
func (r Root) Config() string { return filepath.Join(r.Dir, FileName) }

// DB returns the SQLite database path.
func (r Root) DB() string { return filepath.Join(r.Dir, "jobshop.db") }

// AdminSocket returns the admin socket path.
func (r Root) AdminSocket() string { return filepath.Join(r.Dir, "admin.sock") }

// Lock returns the controller lock file path.
func (r Root) Lock() string { return filepath.Join(r.Dir, "controller.lock") }

// SessionDir returns a session's directory (host record, journal, sockets).
func (r Root) SessionDir(id string) string { return filepath.Join(r.Dir, "sessions", id) }

// Path resolves a config path: absolute as is, else relative to the root.
func (r Root) Path(p string) string {
	if p == "" || filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(r.Dir, p)
}

// ErrNoRoot means no root was given and none was found above the cwd.
var ErrNoRoot = errors.New("no jobshop root: pass --root, set JOBSHOP_ROOT, or run inside a directory containing " + FileName)

// FindRoot resolves the root: explicit dir if non-empty, else the nearest
// ancestor of start (inclusive) that contains jobshop.yaml.
func FindRoot(explicit, start string) (Root, error) {
	if explicit != "" {
		abs, err := filepath.Abs(explicit)
		if err != nil {
			return Root{}, err
		}
		if _, err := os.Stat(filepath.Join(abs, FileName)); err != nil {
			return Root{}, fmt.Errorf("root %s: %w", abs, err)
		}
		return Root{Dir: abs}, nil
	}
	dir, err := filepath.Abs(start)
	if err != nil {
		return Root{}, err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, FileName)); err == nil {
			return Root{Dir: dir}, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return Root{}, ErrNoRoot
		}
		dir = parent
	}
}
