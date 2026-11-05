//go:build !linux

package control

// lowIOPriority is a no-op where there is no ioprio_set.
func lowIOPriority() {}
