//go:build windows

package daemon

import "os/exec"

// stopChildProcess 终止子进程（Windows: Kill）
func stopChildProcess(cmd *exec.Cmd) {
	cmd.Process.Kill()
}
