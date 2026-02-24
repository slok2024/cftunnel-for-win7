package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath" // 新增引用
	"strconv"
	"strings"

	"github.com/qingchencloud/cftunnel/internal/config"
)

// 修改：pidFile 现在也放在程序同级目录
func getPidFile() string {
	return filepath.Join(config.Dir(), "cloudflared.pid")
}

// Start 启动 cloudflared（token 模式）
func Start(token string) error {
	// 这里会调用 EnsureCloudflared，请记得在 download.go 里把该函数改掉
	binPath, err := EnsureCloudflared()
	if err != nil {
		return err
	}
	if Running() {
		return fmt.Errorf("cloudflared 已在运行")
	}

	cmd := exec.Command(binPath, "tunnel", "run", "--token", token)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动 cloudflared 失败: %w", err)
	}

	// 移除 os.MkdirAll(config.Dir(), 0700)，因为当前目录肯定存在
	os.WriteFile(getPidFile(), []byte(strconv.Itoa(cmd.Process.Pid)), 0644)
	fmt.Printf("cloudflared 已启动 (PID: %d)\n", cmd.Process.Pid)
	return nil
}

// Stop 停止 cloudflared
func Stop() error {
	pid, err := readPID()
	if err != nil {
		return fmt.Errorf("未找到运行中的 cloudflared")
	}
	if err := processKill(pid); err != nil {
		return fmt.Errorf("停止 cloudflared 失败: %w", err)
	}
	os.Remove(getPidFile())
	fmt.Println("cloudflared 已停止")
	return nil
}

// Running 检查 cloudflared 是否在运行
func Running() bool {
	pid, err := readPID()
	if err != nil {
		return false
	}
	return processRunning(pid)
}

// PID 返回当前运行的 PID
func PID() int {
	pid, _ := readPID()
	return pid
}

func readPID() (int, error) {
	data, err := os.ReadFile(getPidFile())
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(data)))
}
