//go:build !linux

package journalcompress

// lowIOPriority is a no-op where there is no ioprio_set.
func lowIOPriority() {}
