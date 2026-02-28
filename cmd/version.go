package cmd

import (
	"fmt"
	"github.com/qingchencloud/cftunnel/internal/selfupdate"
	"github.com/spf13/cobra"
)

var checkUpdate bool

func init() {
	versionCmd.Flags().BoolVar(&checkUpdate, "check", false, "检查是否有新版本")
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	// 将 RunE 改为 Run，避免非预期的错误导致退出码异常
	Run: func(cmd *cobra.Command, args []string) {
		if !checkUpdate {
			// 【核心修改】只打印版本号，去掉 "cftunnel " 前缀
			// 确保 Wails 拿到的字符串就是纯粹的 "0.7.2"
			fmt.Print(Version) 
			return
		}

		// 只有带了 --check 参数时，才显示复杂信息
		fmt.Printf("cftunnel %s\n", Version)
		fmt.Println("正在检查更新...")
		latest, err := selfupdate.LatestVersion()
		if err != nil {
			fmt.Printf("检查更新失败: %v\n", err)
			return
		}
		
		if latest == "v"+Version || latest == Version {
			fmt.Println("已是最新版本")
		} else {
			fmt.Printf("发现新版本: %s → %s\n", Version, latest)
			fmt.Println("运行 cftunnel update 进行更新")
		}
	},
}