package control

import (
	"time"

	host "github.com/elee1766/acphost"
)

// journalTail returns the last seq and its time in a session dir's
// journal. acphost reads it from the end or the index, plain or
// compressed (acphost#74), so a closed session's .zst is read the same.
func journalTail(dir string) (int64, time.Time) {
	return host.JournalTail(dir)
}
