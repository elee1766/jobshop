package desk

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
)

func TestArtifactFiles(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "artifacts", "js-1")
	require.NoError(t, os.MkdirAll(dir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "out.csv"), []byte("a,b\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pic.svg"), []byte("<svg/>"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "secret"), []byte("s"), 0o600))
	require.NoError(t, os.Symlink(filepath.Join(root, "secret"), filepath.Join(dir, "link.txt")))
	show := rpc.ArtifactShowResult{
		Artifact: rpc.Artifact{Ticket: domain.Ticket{ID: "js-1"}, Meta: rpc.ArtifactMeta{Files: []rpc.ArtifactFile{
			{Name: "out.csv"}, {Name: "pic.svg"}, {Name: "link.txt"}, {Name: "old.txt", Pruned: true},
		}}},
		Dir: dir,
	}
	ctl := &fakeCtl{results: map[string]any{rpc.MethodArtifactShow: show}}
	h := New(Config{}, ctl, nil).Handler()

	w := get(t, h, "/api/artifacts/js-1/files/out.csv", nil)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "a,b\n", w.Body.String())
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), `filename="out.csv"`)

	w = get(t, h, "/api/artifacts/js-1/files/pic.svg", nil)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment")
	assert.Equal(t, "application/octet-stream", w.Header().Get("Content-Type"))

	for name, path := range map[string]string{
		"traversal":   "/api/artifacts/js-1/files/..%2Fsecret",
		"dotdot":      "/api/artifacts/js-1/files/..",
		"unlisted":    "/api/artifacts/js-1/files/secret",
		"symlink out": "/api/artifacts/js-1/files/link.txt",
	} {
		w := get(t, h, path, nil)
		assert.NotEqual(t, http.StatusOK, w.Code, name)
		assert.NotContains(t, w.Body.String(), "s\n", name)
		assert.NotEqual(t, "s", w.Body.String(), name)
	}
	assert.Equal(t, http.StatusGone, get(t, h, "/api/artifacts/js-1/files/old.txt", nil).Code)

	// The JSON view hides the host dir.
	var got rpc.ArtifactShowResult
	get(t, h, "/api/artifacts/js-1", &got)
	assert.Empty(t, got.Dir)

	// Ack needs the CSRF header like the other writes.
	w = postJSON(h, "/api/artifacts/js-1/ack", `{"note":"ok"}`, nil)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.Equal(t, rpc.MethodArtifactAck+` {"id":"js-1","note":"ok"}`, ctl.calls[len(ctl.calls)-1])
	w = postJSON(h, "/api/artifacts/js-1/ack", `{}`, map[string]string{CSRFHeader: ""})
	assert.Equal(t, http.StatusForbidden, w.Code)
}
