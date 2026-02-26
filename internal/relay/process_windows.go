//go:build windows

package relay

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

// processKill 优雅终止进程（Windows: taskkill）
func processKill(pid int) error {
	return exec.Command("taskkill", "/PID", strconv.Itoa(pid)).Run()
}
