//go:build windows

package main

import (
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func getTerminalSize() (int, int) {
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", "$Host.UI.RawUI.WindowSize.Height; $Host.UI.RawUI.WindowSize.Width")
	if out, err := cmd.Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(lines) >= 2 {
			r, err1 := strconv.Atoi(strings.TrimSpace(lines[0]))
			c, err2 := strconv.Atoi(strings.TrimSpace(lines[1]))
			if err1 == nil && err2 == nil && r > 0 && c > 0 {
				return r, c
			}
		}
	}
	return 24, 80
}

func isProcessAlive(pid int) bool {
	const PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
	h, err := syscall.OpenProcess(PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(h)
	var exitCode uint32
	if err := syscall.GetExitCodeProcess(h, &exitCode); err != nil {
		return false
	}
	const STILL_ACTIVE = 259
	return exitCode == STILL_ACTIVE
}
