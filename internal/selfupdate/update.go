package selfupdate

import (
	"fmt"
)

// LatestVersion 禁用联网查询，永远返回当前版本或报错
func LatestVersion() (string, error) {
	// 禁用联网查询，直接返回错误，避免启动时产生网络请求
	return "", fmt.Errorf("绿色版已禁用自动更新检查")
}

// Update 彻底禁用更新替换功能
func Update(version string) error {
	// 防止程序尝试下载并覆盖自身
	return fmt.Errorf("当前版本为本地绿色定制版，不支持在线更新，请手动替换程序文件")
}

// 后面那些 extractZip 和 extractTarGz 函数可以直接删掉，
// 因为上面已经不再调用它们了，删掉可以减小最终编译出的 exe 体积。