//go:build windows

package daemon

import (
	"os/exec"
	"strconv"
	"strings"
)

// processRunning 检查进程是否存活（Windows: tasklist）
func processRunning(pid int) bool {
	out, err := exec.Command("tasklist", "/FI", "PID eq "+strconv.Itoa(pid), "/NH").Output()
	if err != nil {
		return false
	}
	return !strings.Contains(string(out), "No tasks")
}

// processKill 终止进程（Windows: taskkill）
func processKill(pid int) error {
	return exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/F").Run()
}
