//go:build windows

package main

import (
	"os"
	"os/exec"
)

func setupProcessGroup(cmd *exec.Cmd) {
	// On Windows, process lifecycle is managed via native console/job objects
}

func killProcessGroup(cmd *exec.Cmd, sig os.Signal) {
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Signal(sig)
	}
}
