//go:build !unix

package automation

import "os/exec"

func configureProcess(_ *exec.Cmd) {}
