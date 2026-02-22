package cmd

import (
	"github.com/qingchencloud/cftunnel/internal/daemon"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(quickCmd)
}

var quickCmd = &cobra.Command{
	Use:   "quick <端口>",
	Short: "快速启动免域名隧道（生成 *.trycloudflare.com 随机域名）",
	Long:  "无需 Cloudflare 账户、API Token 或域名，一条命令生成临时公网地址。\n适合临时分享、快速调试，Ctrl+C 退出后域名自动失效。",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return daemon.StartQuick(args[0])
	},
}
