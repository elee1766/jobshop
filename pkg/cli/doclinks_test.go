package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// mdLink matches inline markdown links: [text](target).
var mdLink = regexp.MustCompile(`\]\(([^)\s]+)\)`)

// TestDocLinks checks that relative links in README.md, llm/*.md and
// docs/*.md resolve to files in the repo (#551). Anchors are not checked.
func TestDocLinks(t *testing.T) {
	repo, err := filepath.Abs("../..")
	require.NoError(t, err)
	files := []string{filepath.Join(repo, "README.md")}
	for _, g := range []string{"llm/*.md", "docs/*.md"} {
		m, err := filepath.Glob(filepath.Join(repo, g))
		require.NoError(t, err)
		files = append(files, m...)
	}
	for _, f := range files {
		b, err := os.ReadFile(f)
		require.NoError(t, err)
		for _, m := range mdLink.FindAllStringSubmatch(stripCode(string(b)), -1) {
			target := m[1]
			if strings.Contains(target, "://") || strings.HasPrefix(target, "#") || strings.HasPrefix(target, "mailto:") {
				continue
			}
			target, _, _ = strings.Cut(target, "#")
			p := filepath.Join(filepath.Dir(f), target)
			_, err := os.Stat(p)
			rel, _ := filepath.Rel(repo, f)
			require.NoError(t, err, "%s: broken link %s", rel, m[1])
		}
	}
}

// stripCode drops fenced code blocks, where ](...) is not a link.
func stripCode(s string) string {
	var b strings.Builder
	in := false
	for _, line := range strings.Split(s, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			in = !in
			continue
		}
		if !in {
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}
	return b.String()
}
