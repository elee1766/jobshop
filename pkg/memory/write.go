package memory

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

// Local writes (#592): the curator writes memory through the controller,
// which records each write so the gate can restore it.

var namePart = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// Target maps a scope name to its file under the memory directory:
// global, repo/<owner>__<name>, pool/<name> or skills/<name>. A trailing
// ".md" on the name is accepted.
func Target(scope string) (string, error) {
	s := strings.TrimSuffix(strings.TrimSpace(scope), ".md")
	kind, name, _ := strings.Cut(s, "/")
	switch {
	case kind == "global" && name == "":
		return filepath.Join("memory", "global.md"), nil
	case (kind == "repo" || kind == "pool") && namePart.MatchString(name) && !strings.Contains(name, ".."):
		return filepath.Join("memory", kind, name+".md"), nil
	case kind == "skills" && namePart.MatchString(name) && !strings.Contains(name, ".."):
		return filepath.Join("skills", name, "SKILL.md"), nil
	}
	return "", fmt.Errorf("memory scope %q: want global, repo/<owner>__<name>, pool/<name> or skills/<name>", scope)
}

// ReadRaw returns the file at rel under dir and whether it exists.
func ReadRaw(dir, rel string) (string, bool, error) {
	b, err := os.ReadFile(filepath.Join(dir, rel))
	if errors.Is(err, fs.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return string(b), true, nil
}

// Write writes text to rel under dir, creating parent directories.
// Memory files (not skills) over limit characters are refused, since
// injection would cut them.
func Write(dir, rel, text string, limit int) error {
	if limit <= 0 {
		limit = DefaultCap
	}
	if strings.HasPrefix(rel, "memory"+string(filepath.Separator)) {
		if n := utf8.RuneCountInString(strings.TrimSpace(text)); n > limit {
			return fmt.Errorf("%w: %s is %d chars, cap %d", ErrOverCap, rel, n, limit)
		}
	}
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, []byte(text), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// Restore puts back before (existed) or removes the file (not existed).
func Restore(dir, rel, before string, existed bool) error {
	if existed {
		return Write(dir, rel, before, int(^uint(0)>>1))
	}
	err := os.Remove(filepath.Join(dir, rel))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err == nil && strings.HasPrefix(rel, "skills"+string(filepath.Separator)) {
		// An empty skill directory would still list as a skill.
		_ = os.Remove(filepath.Dir(filepath.Join(dir, rel)))
	}
	return err
}

// ErrOverCap is returned by Write for a memory file over the cap.
var ErrOverCap = errors.New("over the memory cap")

// List returns every memory and skill file under dir, relative, sorted.
func List(dir string) ([]File, error) {
	var out []File
	for _, sub := range []string{"memory", "skills"} {
		root := filepath.Join(dir, sub)
		err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if errors.Is(err, fs.ErrNotExist) && p == root {
				return fs.SkipAll
			}
			if err != nil || d.IsDir() || filepath.Ext(p) != ".md" {
				return err
			}
			rel, err := filepath.Rel(dir, p)
			if err != nil {
				return err
			}
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			text := string(b)
			out = append(out, File{Scope: sub, Path: rel, Text: text, Chars: utf8.RuneCountInString(strings.TrimSpace(text))})
			return nil
		})
		if err != nil {
			return out, err
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}
