//go:build windows || plan9 || js

package proc

import "os/exec"

// started keeps the command a process "group" refers to on platforms that have no process
// groups, so that the timeout can still terminate the process it started.
var started = map[int]*exec.Cmd{}

// setProcessGroup is a no-op on platforms without process groups.
func setProcessGroup(cmd *exec.Cmd) {}

// processGroup returns a handle for the started command, or 0 when it has no process.
func processGroup(cmd *exec.Cmd) int {
	if cmd == nil || cmd.Process == nil {
		return 0
	}

	started[cmd.Process.Pid] = cmd

	return cmd.Process.Pid
}

// terminateProcessGroup terminates the command, as its descendants cannot be signaled as a
// group on this platform.
func terminateProcessGroup(group int) {
	killProcessGroup(group)
}

// killProcessGroup kills the command the given handle refers to.
func killProcessGroup(group int) {
	cmd, ok := started[group]

	if !ok || cmd.Process == nil {
		return
	}

	delete(started, group)

	_ = cmd.Process.Kill()
}
