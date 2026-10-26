//go:build !windows

package main

import (
	"os"
	"os/exec"
	"syscall"
)

func setupProcessGroup(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setpgid = true
}

func killProcessGroup(cmd *exec.Cmd, sig os.Signal) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	if sysSig, ok := sig.(syscall.Signal); ok {
		_ = syscall.Kill(-cmd.Process.Pid, sysSig)
		return
	}
	_ = cmd.Process.Signal(sig)
}
