package control_test

import "os"

// selfPid is skipped by cleanup that signals host pids: an in-process
// host (controltest.InProcHosts) writes the test process's pid.
var selfPid = os.Getpid()
