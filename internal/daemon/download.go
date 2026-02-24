package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/qingchencloud/cftunnel/internal/config"
)

// CloudflaredPath 修改：取消 bin 目录，直接返回程序同级目录下的路径
func CloudflaredPath() string {
	name := "cloudflared"
	if runtime.GOOS == "windows" {
		name = "cloudflared.exe"
	}
	// config.Dir() 在我们之前的修改中已经指向了程序所在目录
	return filepath.Join(config.Dir(), name)
}

// EnsureCloudflared 修改：彻底禁用自动下载，只做本地存在性检查
func EnsureCloudflared() (string, error) {
	path := CloudflaredPath()
	
	// 1. 检查同级目录下是否存在
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	// 2. 报错并提示用户，不再调用 download(path)
	return "", fmt.Errorf("错误: 未在同级目录找到 %s\n请手动将该文件放入目录: %s", filepath.Base(path), config.Dir())
}

// 彻底删除或注释掉以下不再使用的函数以保持精简
// func download(dest string) error { ... }
// func extractTgz(r io.Reader, dest string) error { ... }
// func downloadURL() (string, error) { ... }
