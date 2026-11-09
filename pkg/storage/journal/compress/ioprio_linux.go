//go:build linux

package compress

import "golang.org/x/sys/unix"

// lowIOPriority puts the calling thread in the idle IO class, so a
// journal compression sweep does not compete with sessions for disk.
// Best effort: an error leaves the priority as it was.
func lowIOPriority() {
	const (
		whoProcess = 1 // IOPRIO_WHO_PROCESS; id 0 is the calling thread
		classIdle  = 3 // IOPRIO_CLASS_IDLE
		classShift = 13
	)
	_, _, _ = unix.Syscall(unix.SYS_IOPRIO_SET, whoProcess, 0, classIdle<<classShift)
}
