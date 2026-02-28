package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/qingchencloud/cftunnel/internal/config"
)

// pidFilePath 返回 cloudflared PID 文件路径
func pidFilePath() string {
	// 确保使用 config.Dir() 获取的绝对路径
	return filepath.Join(config.Dir(), "cloudflared.pid")
}

// Start 启动 cloudflared
func Start(token string) error {
	// 1. 强制使用绝对路径定位 cloudflared.exe
	dir := config.Dir()
	binPath := filepath.Join(dir, "cloudflared.exe")

	// 检查本地文件是否存在
	if _, err := os.Stat(binPath); err != nil {
		return fmt.Errorf("错误: 未在程序目录找到内核文件 cloudflared.exe (路径: %s)", dir)
	}

	if Running() {
		return fmt.Errorf("cloudflared 已在运行")
	}

	// 2. 构造启动命令
	cmd := exec.Command(binPath, "tunnel", "--protocol", "http2", "run", "--token", token)
	
	// 3. 关键修复：设置子进程的工作目录
	// 这保证了 cloudflared.exe 如果需要产生临时文件，也会留在程序目录下
	cmd.Dir = dir

	// 4. 隐藏窗口逻辑
	hideWindow(cmd)

	// 5. 关键修复：处理标准输出，防止 Win7 下父子进程管道死锁
	// 当作为内核被 GUI 调用时，不要直接赋值给 os.Stdout
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动 cloudflared 失败: %w", err)
	}

	// 6. 记录 PID
	// 修复：放宽权限到 0755 和 0644，避免 Win7 报 Access Denied
	_ = os.MkdirAll(dir, 0755)
	err := os.WriteFile(pidFilePath(), []byte(strconv.Itoa(cmd.Process.Pid)), 0644)
	if err != nil {
		// 如果写 PID 失败，虽然不影响进程运行，但会影响后续停止操作
		fmt.Printf("警告: 无法写入 PID 文件: %v\n", err)
	}

	fmt.Printf("cloudflared 已启动 (PID: %d)\n", cmd.Process.Pid)
	return nil
}

// Stop 停止 cloudflared
func Stop() error {
	pid, err := readPID()
	if err != nil {
		return fmt.Errorf("未找到运行中的 cloudflared (可能已停止)")
	}

	// 调用 process_windows.go 里的逻辑
	if err := processKill(pid); err != nil {
		return fmt.Errorf("停止 cloudflared 失败: %w", err)
	}

	// 删除 PID 文件
	_ = os.Remove(pidFilePath())
	fmt.Println("cloudflared 已停止")
	return nil
}

// Running 检查是否在运行
func Running() bool {
	pid, err := readPID()
	if err != nil {
		return false
	}
	return processRunning(pid)
}

// PID 返回当前 PID
func PID() int {
	pid, _ := readPID()
	return pid
}

// --- 内部辅助函数 ---

func readPID() (int, error) {
	data, err := os.ReadFile(pidFilePath())
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(string(data))
	if s == "" {
		return 0, fmt.Errorf("PID file is empty")
	}
	return strconv.Atoi(s)
}

func hideWindow(cmd *exec.Cmd) {
	if runtime.GOOS == "windows" {
		if cmd.SysProcAttr == nil {
			cmd.SysProcAttr = &syscall.SysProcAttr{}
		}
		// Win7 隐藏窗口最稳妥的组合
		cmd.SysProcAttr.HideWindow = true
		
		// CREATE_NO_WINDOW (0x08000000)
		// 如果在某些精简版 Win7 下依然报错，可以尝试去掉这一行
		cmd.SysProcAttr.CreationFlags = 0x08000000 
	}
}