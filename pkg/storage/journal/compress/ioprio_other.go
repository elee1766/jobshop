//go:build !linux

package compress

// lowIOPriority is a no-op where there is no ioprio_set.
func lowIOPriority() {}
