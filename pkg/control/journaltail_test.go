package control

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	host "github.com/elee1766/acphost"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJournalTail reads the last frame of a plain journal, skips a torn
// last line, and reads the same tail once the journal is compressed.
func TestJournalTail(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, host.JournalFile)
	big := strings.Repeat("x", 1000)
	body := ""
	for i := 1; i <= 5; i++ {
		body += fmt.Sprintf(`{"seq":%d,"t":%d,"from":"agent","msg":{"text":%q}}`+"\n", i, 1000+i, big)
	}
	require.NoError(t, os.WriteFile(path, []byte(body+`{"seq":6,"t":`), 0o600))
	seq, _ := journalTail(dir)
	assert.Equal(t, int64(5), seq)

	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
	seq, at := journalTail(dir)
	assert.Equal(t, int64(5), seq)
	assert.Equal(t, int64(1005), at.UnixMilli())

	require.NoError(t, host.CompressJournal(dir, 3))
	require.NoFileExists(t, path)
	seq, at = journalTail(dir)
	assert.Equal(t, int64(5), seq)
	assert.Equal(t, int64(1005), at.UnixMilli())

	seq, _ = journalTail(t.TempDir())
	assert.Zero(t, seq)
}
