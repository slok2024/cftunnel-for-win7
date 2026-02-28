package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "更新 cftunnel (当前版本已禁用自动更新)",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 直接拦截，不再调用 internal/selfupdate
		fmt.Printf("当前版本: %s (本地绿色定制版)\n", Version)
		fmt.Println("提示: 为了系统安全与稳定性，已禁用在线自动更新功能。")
		fmt.Println("如需更新，请联系管理员获取最新的安装包并手动替换 cftunnel.exe。")
		return nil
	},
}