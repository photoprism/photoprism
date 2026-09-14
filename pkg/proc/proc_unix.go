//go:build !windows && !plan9 && !js

package proc

import (
	"os/exec"
	"syscall"
)

// setProcessGroup makes the command the leader of a new process group, so that its descendants
// can be signaled as a unit.
func setProcessGroup(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}

	cmd.SysProcAttr.Setpgid = true
}

// processGroup returns the group a started command leads, or 0 when it does not lead one.
//
// Only a group the command itself created is reported. A command that inherited the caller's
// group must never be signaled by group, as that group is the caller's own.
func processGroup(cmd *exec.Cmd) int {
	if cmd == nil || cmd.Process == nil {
		return 0
	}

	pid := cmd.Process.Pid

	if pgid, err := syscall.Getpgid(pid); err != nil || pgid != pid {
		return 0
	}

	return pid
}

// terminateProcessGroup asks the given process group to terminate.
func terminateProcessGroup(group int) {
	signalProcessGroup(group, syscall.SIGTERM)
}

// killProcessGroup kills the given process group.
func killProcessGroup(group int) {
	signalProcessGroup(group, syscall.SIGKILL)
}

// signalProcessGroup sends the given signal to the given process group.
func signalProcessGroup(group int, sig syscall.Signal) {
	if group <= 0 {
		return
	}

	_ = syscall.Kill(-group, sig)
}
