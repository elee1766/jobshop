package deploy_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDocsNoNumberedAppendLists guards #131: parallel PRs that append
// numbered sections take the same number and conflict, so plan.md section
// 7 and gasserver-verify.md hold no numbered subsections, and every doc
// path they or the deploy scripts name exists.
func TestDocsNoNumberedAppendLists(t *testing.T) {
	plan, err := os.ReadFile("../docs/plan.md")
	require.NoError(t, err)
	assert.NotRegexp(t, regexp.MustCompile(`(?m)^### 7\.[0-9]`), string(plan))
	verify, err := os.ReadFile("../docs/gasserver-verify.md")
	require.NoError(t, err)
	assert.NotRegexp(t, regexp.MustCompile(`(?m)^## [0-9]+\.`), string(verify))

	link := regexp.MustCompile(`\]\(((?:notes|verify)/[a-z0-9-]+\.md)\)`)
	for _, doc := range []string{string(plan), string(verify)} {
		for _, m := range link.FindAllStringSubmatch(doc, -1) {
			assert.FileExists(t, filepath.Join("../docs", m[1]))
		}
	}
	scripts, err := filepath.Glob("*.sh")
	require.NoError(t, err)
	ref := regexp.MustCompile(`docs/[a-z0-9/-]+\.md`)
	for _, s := range scripts {
		b, err := os.ReadFile(s)
		require.NoError(t, err)
		for _, p := range ref.FindAllString(string(b), -1) {
			assert.FileExists(t, filepath.Join("..", p), s)
		}
	}
}
