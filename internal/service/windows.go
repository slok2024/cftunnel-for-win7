//go:build windows

package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Windows struct{}

const svcName = "cftunnel-kernel" // 建议改个名字，避免跟主程序服务冲突

func (w *Windows) Install(binPath, token string) error {
	// 【安全性增强】强制校验 binPath 是否有效
	// 如果传入的是相对路径，通过 os.Executable 转换为绝对路径，确保服务能启动
	absPath, err := filepath.Abs(binPath)
	if err != nil {
		absPath = binPath // 降级处理
	}

	// 检查内核文件是否存在
	if _, err := os.Stat(absPath); err != nil {
		return fmt.Errorf("安装失败：在路径 %s 未找到内核文件", absPath)
	}

	// Windows sc 命令要求 binPath 参数如果包含空格，必须用引号包裹
	// 且 sc 的参数格式非常古怪，"binPath=" 后面必须有一个空格
	binArg := fmt.Sprintf(`"%s" tunnel --protocol http2 run --token %s`, absPath, token)
	
	// 创建服务：设置自动启动
	cmd := exec.Command("sc", "create", svcName, "binPath=", binArg, "start=", "auto", "DisplayName=", "Cloudflare Tunnel Kernel")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("创建系统服务失败(请尝试以管理员权限运行): %w", err)
	}
	
	// 启动服务
	return exec.Command("sc", "start", svcName).Run()
}

func (w *Windows) Uninstall() error {
	// 停止并删除服务
	exec.Command("sc", "stop", svcName).Run()
	return exec.Command("sc", "delete", svcName).Run()
}

func (w *Windows) Running() bool {
	out, err := exec.Command("sc", "query", svcName).Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "RUNNING")
}

func New() Service {
	return &Windows{}
}