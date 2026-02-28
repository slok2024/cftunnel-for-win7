package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// 获取程序所在绝对目录的通用工具函数
func getAbsDir() string {
	// 优先获取执行文件路径
	exe, err := os.Executable()
	if err != nil {
		// 如果 Executable 失败，尝试用 Args[0] 兜底（Win7 下较稳）
		exe = os.Args[0]
	}
	// 关键：强制转换为绝对路径，解决工作目录偏移问题
	absPath, _ := filepath.Abs(exe)
	return filepath.Dir(absPath)
}

// CloudflaredPath 返回程序同级目录下的二进制路径
func CloudflaredPath() string {
	name := "cloudflared"
	if runtime.GOOS == "windows" {
		name = "cloudflared.exe"
	}
	// 使用强化的路径获取逻辑
	return filepath.Join(getAbsDir(), name)
}

// FrpcPath 返回程序同级目录下的 frpc 路径
func FrpcPath() string {
	name := "frpc"
	if runtime.GOOS == "windows" {
		name = "frpc.exe"
	}
	// 使用强化的路径获取逻辑
	return filepath.Join(getAbsDir(), name)
}

// EnsureCloudflared 仅检查本地，不下载
func EnsureCloudflared() (string, error) {
	path := CloudflaredPath()
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		return path, nil
	}
	// 报错时输出具体寻找的路径，方便在 Win7 上调试
	return "", fmt.Errorf("缺失内核: 请将 cloudflared.exe 放入目录 %s", getAbsDir())
}

// EnsureFrpc 仅检查本地，不下载
func EnsureFrpc() (string, error) {
	path := FrpcPath()
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		return path, nil
	}
	return "", fmt.Errorf("缺失内核: 请将 frpc.exe 放入目录 %s", getAbsDir())
}

// 屏蔽下载逻辑
func download(dest string) error { return fmt.Errorf("下载功能已禁用") }