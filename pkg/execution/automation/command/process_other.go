//go:build !unix

package command

import "os/exec"

func configureProcess(_ *exec.Cmd) {}
