package control

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJournalTailFile reads the last frame from the end, including one
// larger than the read block and a trailing partial line.
func TestJournalTailFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "j")
	big := strings.Repeat("x", 1000)
	body := ""
	for i := 1; i <= 5; i++ {
		body += fmt.Sprintf(`{"seq":%d,"t":%d,"from":"agent","msg":{"text":%q}}`+"\n", i, 1000+i, big)
	}
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
	for _, block := range []int64{16, 999, 1 << 20} {
		seq, at := journalTailFile(path, block)
		assert.Equal(t, int64(5), seq, "block %d", block)
		assert.Equal(t, int64(1005), at.UnixMilli())
	}
	// A torn last line is skipped.
	require.NoError(t, os.WriteFile(path, []byte(body+`{"seq":6,"t":`), 0o600))
	seq, _ := journalTailFile(path, 16)
	assert.Equal(t, int64(5), seq)
	// A single frame larger than the block.
	require.NoError(t, os.WriteFile(path, []byte(`{"seq":1,"t":7,"msg":"`+big+`"}`+"\n"), 0o600))
	seq, _ = journalTailFile(path, 16)
	assert.Equal(t, int64(1), seq)
	seq, _ = journalTailFile(filepath.Join(dir, "missing"), 16)
	assert.Zero(t, seq)
	require.NoError(t, os.WriteFile(path, nil, 0o600))
	seq, _ = journalTailFile(path, 16)
	assert.Zero(t, seq)
}
