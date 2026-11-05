package control

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"

	host "github.com/elee1766/acphost"
)

// tailBlock is the first read size from the end of a journal; it doubles
// until a complete frame fits.
const tailBlock = 64 << 10

// journalTail returns the last seq and its time in a session dir's
// journal. It reads from the end instead of decoding every line like
// host.JournalTail (#614): seq is monotonic, so the last frame is the max.
// Move to acphost once it carries the same change.
func journalTail(dir string) (int64, time.Time) {
	return journalTailFile(filepath.Join(dir, host.JournalFile), tailBlock)
}

func journalTailFile(path string, block int64) (int64, time.Time) {
	f, err := os.Open(path)
	if err != nil {
		return 0, time.Time{}
	}
	defer func() { _ = f.Close() }()
	st, err := f.Stat()
	if err != nil {
		return 0, time.Time{}
	}
	size := st.Size()
	for {
		n := min(block, size)
		buf := make([]byte, n)
		if _, err := f.ReadAt(buf, size-n); err != nil && err != io.EOF {
			return 0, time.Time{}
		}
		// Walk complete lines from the end. The first line in the block is
		// complete only when the block starts at the file's beginning.
		rest := bytes.TrimRight(buf, "\n")
		for len(rest) > 0 {
			i := bytes.LastIndexByte(rest, '\n')
			if i < 0 && n < size {
				break // partial line: read a bigger block
			}
			line := rest[i+1:]
			rest = rest[:max(i, 0)]
			var e struct {
				Seq int64 `json:"seq"`
				T   int64 `json:"t"`
			}
			if json.Unmarshal(line, &e) == nil && e.Seq > 0 {
				return e.Seq, time.UnixMilli(e.T)
			}
			if i < 0 {
				break
			}
		}
		if n >= size {
			return 0, time.Time{}
		}
		block *= 2
	}
}
