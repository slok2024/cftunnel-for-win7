package relay

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// FrpcPath 返回 frpc 二进制在程序同级目录的路径
func FrpcPath() string {
	name := "frpc"
	if runtime.GOOS == "windows" {
		name = "frpc.exe"
	}
	// 锁死在 cftunnel.exe 所在的同级目录
	exe, _ := os.Executable()
	return filepath.Join(filepath.Dir(exe), name)
}

// FrpsPath 返回 frps 二进制在程序同级目录的路径（如果用到服务端功能）
func FrpsPath() string {
	name := "frps"
	if runtime.GOOS == "windows" {
		name = "frps.exe"
	}
	exe, _ := os.Executable()
	return filepath.Join(filepath.Dir(exe), name)
}

// EnsureFrpc 确保 frpc 已存在，不存在则报错，不再自动下载
func EnsureFrpc() (string, error) {
	path := FrpcPath()
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("缺失中继内核: 请将 frpc.exe 放入程序同级目录")
}

// EnsureFrps 确保 frps 已存在
func EnsureFrps() (string, error) {
	path := FrpsPath()
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("缺失服务端内核: 请将 frps.exe 放入程序同级目录")
}

// 屏蔽掉所有的下载函数逻辑，保持函数定义以兼容其他文件的调用
func downloadFrp(dest, binary string) error {
	return fmt.Errorf("下载功能已禁用，请手动提供内核文件")
}

func frpFilename() (string, error) {
	return "", nil
}

func extractFrpBinary(r any, dest, filename, binary string) error {
	return nil
}