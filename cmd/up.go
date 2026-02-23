package cmd

import (
	"context"
	"fmt"

	"github.com/qingchencloud/cftunnel/internal/cfapi"
	"github.com/qingchencloud/cftunnel/internal/config"
	"github.com/qingchencloud/cftunnel/internal/daemon"
	"github.com/qingchencloud/cftunnel/internal/selfupdate"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(upCmd)
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "启动隧道",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if cfg.Tunnel.Token == "" {
			return fmt.Errorf("请先运行 cftunnel init && cftunnel create <名称>")
		}
		// 启动前同步 ingress 配置到远端，确保本地与远端一致
		if len(cfg.Routes) > 0 {
			client := cfapi.New(cfg.Auth.APIToken, cfg.Auth.AccountID)
			if err := pushIngress(client, context.Background(), cfg); err != nil {
				fmt.Printf("警告: 同步 ingress 失败: %v（将使用远端现有配置）\n", err)
			} else {
				fmt.Println("ingress 配置已同步")
			}
		}

		// 自动检查更新（非阻塞，仅提示）
		if cfg.SelfUpdate.AutoCheck {
			if latest, err := selfupdate.LatestVersion(); err == nil {
				if latest != "v"+Version && latest != Version {
					fmt.Printf("发现新版本: %s → %s (运行 cftunnel update 更新)\n", Version, latest)
				}
			}
		}
		return daemon.Start(cfg.Tunnel.Token)
	},
}
