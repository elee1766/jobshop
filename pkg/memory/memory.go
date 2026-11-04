// Package memory reads a memory directory (#266): bounded markdown memory
// files scoped global, per repo and per pool, and agentskills.io skills,
// rendered into prompt text by a prompt rule.
//
// Layout under the directory:
//
//	memory/global.md
//	memory/repo/<owner>__<name>.md
//	memory/pool/<name>.md
//	skills/<name>/SKILL.md   (frontmatter: name, description)
package memory

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

// DefaultCap is the per-file character cap when none is configured.
const DefaultCap = 2200

// Scope says which memory files apply to a turn. Empty fields select
// nothing at that scope.
type Scope struct {
	Repo string // owner/name
	Pool string
}

// File is one memory file that applies to a scope.
type File struct {
	Scope string // global, repo or pool
	Path  string // relative to the memory directory
	Text  string // after the cap
	Chars int    // before the cap
	Over  bool   // Chars > cap: Text is truncated
}

// Skill is one skills/<dir>/SKILL.md.
type Skill struct {
	Name, Description, Path string // Path is dir joined with skills/<dir>/SKILL.md
}

// RepoFile is the memory file name for an owner/name repo.
func RepoFile(repo string) string {
	return strings.ReplaceAll(repo, "/", "__") + ".md"
}

// Files returns the memory files for s in order global, repo, pool,
// each cut to limit characters (0 means DefaultCap). Missing files and a
// missing directory are skipped.
func Files(dir string, s Scope, limit int) ([]File, error) {
	type cand struct{ scope, rel string }
	cands := []cand{{"global", filepath.Join("memory", "global.md")}}
	if s.Repo != "" && !strings.Contains(s.Repo, "..") {
		cands = append(cands, cand{"repo", filepath.Join("memory", "repo", RepoFile(s.Repo))})
	}
	if s.Pool != "" && !strings.ContainsAny(s.Pool, `/\`) {
		cands = append(cands, cand{"pool", filepath.Join("memory", "pool", s.Pool+".md")})
	}
	var out []File
	for _, c := range cands {
		f, ok, err := read(dir, c.scope, c.rel, limit)
		if err != nil {
			return out, err
		}
		if ok {
			out = append(out, f)
		}
	}
	return out, nil
}

func read(dir, scope, rel string, limit int) (File, bool, error) {
	if limit <= 0 {
		limit = DefaultCap
	}
	b, err := os.ReadFile(filepath.Join(dir, rel))
	if errors.Is(err, fs.ErrNotExist) {
		return File{}, false, nil
	}
	if err != nil {
		return File{}, false, err
	}
	text := strings.TrimSpace(string(b))
	if text == "" {
		return File{}, false, nil
	}
	f := File{Scope: scope, Path: rel, Text: text, Chars: utf8.RuneCountInString(text)}
	if f.Chars > limit {
		f.Over = true
		f.Text = string([]rune(text)[:limit])
	}
	return f, true, nil
}

// Skills lists skills/<dir>/SKILL.md sorted by name. A SKILL.md without
// a name in its frontmatter uses the directory name. A missing directory
// is no skills.
func Skills(dir string) ([]Skill, error) {
	entries, err := os.ReadDir(filepath.Join(dir, "skills"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []Skill
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		rel := filepath.Join("skills", e.Name(), "SKILL.md")
		f, err := os.Open(filepath.Join(dir, rel))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return out, err
		}
		meta := frontmatter(f)
		_ = f.Close()
		sk := Skill{Name: meta["name"], Description: meta["description"], Path: filepath.Join(dir, rel)}
		if sk.Name == "" {
			sk.Name = e.Name()
		}
		out = append(out, sk)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// frontmatter reads flat key: value pairs between leading --- lines.
func frontmatter(f *os.File) map[string]string {
	m := map[string]string{}
	sc := bufio.NewScanner(f)
	if !sc.Scan() || strings.TrimSpace(sc.Text()) != "---" {
		return m
	}
	for sc.Scan() {
		line := sc.Text()
		if strings.TrimSpace(line) == "---" {
			break
		}
		k, v, ok := strings.Cut(line, ":")
		if !ok || strings.HasPrefix(line, " ") {
			continue
		}
		m[strings.TrimSpace(k)] = strings.Trim(strings.TrimSpace(v), `"'`)
	}
	return m
}

// Render is the prompt text for s: the memory files, then a one-line-per-
// skill index the agent reads on demand. It returns the files cut to the
// cap so the caller can warn. A missing directory renders "".
func Render(dir string, s Scope, limit int) (string, []File, error) {
	files, err := Files(dir, s, limit)
	if err != nil {
		return "", nil, err
	}
	skills, err := Skills(dir)
	if err != nil {
		return "", nil, err
	}
	var b strings.Builder
	var over []File
	for _, f := range files {
		if b.Len() > 0 {
			b.WriteString("\n\n")
		}
		fmt.Fprintf(&b, "Memory (%s, %s):\n%s", f.Scope, f.Path, f.Text)
		if f.Over {
			over = append(over, f)
		}
	}
	if len(skills) > 0 {
		if b.Len() > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString("Skills (read the file when one applies):")
		for _, sk := range skills {
			if sk.Description == "" {
				fmt.Fprintf(&b, "\n- %s (%s)", sk.Name, sk.Path)
				continue
			}
			fmt.Fprintf(&b, "\n- %s: %s (%s)", sk.Name, sk.Description, sk.Path)
		}
	}
	return b.String(), over, nil
}

// Check returns every memory file under dir/memory over the cap.
func Check(dir string, limit int) ([]File, error) {
	root := filepath.Join(dir, "memory")
	var over []File
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
		f, ok, err := read(dir, "", rel, limit)
		if err == nil && ok && f.Over {
			over = append(over, f)
		}
		return err
	})
	return over, err
}
