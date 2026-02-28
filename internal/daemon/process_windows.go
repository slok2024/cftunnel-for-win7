//go:build windows

package daemon

import (
	"os"
	"syscall"
	"time"
)

// processRunning 增强版：处理 Win7 下权限受限导致的误判
func processRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	
	// 在 Windows 上，os.FindProcess 总是成功（它只是建立一个对象）
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// 使用 Signal(0) 探测
	err = p.Signal(syscall.Signal(0))
	if err == nil {
		return true
	}

	// 针对 Win7 的特殊处理：如果 err 是权限问题，说明进程通常还在跑
	// 如果是 syscall.Errno(0x5) 即 Access Denied
	if errno, ok := err.(syscall.Errno); ok && errno == 5 {
		return true
	}

	return false
}

// processKill 增强版：带确认逻辑的强杀
func processKill(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	// 强制终止
	err = p.Kill()
	if err != nil {
		return err
	}

	// 给 Win7 一点响应时间，循环确认进程是否真的消失
	// 防止因进程僵死导致 PID 文件被删但端口仍被占用的情况
	for i := 0; i < 10; i++ {
		if !processRunning(pid) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}