package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var Version = "dev"

var rootCmd = &cobra.Command{
	Use:     "cftunnel",
	Short:   "Cloudflare Tunnel 一键管理工具 (本地内核版)",
	Version: Version,
	// 【关键修改】取消在这里的 PersistentPreRun
	// 否则 version 命令也会输出干扰信息
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// 初始化
func init() {
	// 我们可以把原来的逻辑移到那些真正需要路径的命令里
	// 或者通过这种方式判断：如果是 version 命令，就不打印路径
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		// 只有当执行的不是 version 命令时，才打印这些调试信息
		if cmd.Name() != "version" {
			checkWindowsVersion()
			// 如果有需要，可以打印调试信息
			// fmt.Printf("[绿色版] 运行目录: %s\n", config.Dir())
		}
	}
}