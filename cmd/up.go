package cmd

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/qingchencloud/cftunnel/internal/authproxy"
	"github.com/qingchencloud/cftunnel/internal/cfapi"
	"github.com/qingchencloud/cftunnel/internal/config"
	"github.com/qingchencloud/cftunnel/internal/daemon"
	// "github.com/qingchencloud/cftunnel/internal/selfupdate" // 【删除】不再需要检查更新模块
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

		// 为有鉴权配置的路由启动代理
		var proxies []*authproxy.Proxy
		for i, r := range cfg.Routes {
			if r.Auth == nil {
				continue
			}
			sigKey, err := hex.DecodeString(r.Auth.SigningKey)
			if err != nil {
				return fmt.Errorf("路由 %s 的 signing_key 无效: %w", r.Name, err)
			}
			port := extractPort(r.Service)
			if port == "" {
				return fmt.Errorf("路由 %s 的 service 格式无效: %s", r.Name, r.Service)
			}
			proxy, err := authproxy.New(authproxy.Config{
				Username:   r.Auth.Username,
				Password:   r.Auth.Password,
				TargetPort: port,
				SigningKey: sigKey,
				CookieTTL:  time.Duration(r.Auth.CookieTTLOrDefault()) * time.Second,
			})
			if err != nil {
				return fmt.Errorf("路由 %s 启动鉴权代理失败: %w", r.Name, err)
			}
			if err := proxy.Start(); err != nil {
				return fmt.Errorf("路由 %s 启动鉴权代理失败: %w", r.Name, err)
			}
			proxies = append(proxies, proxy)
			proxyPort := strconv.Itoa(proxy.ListenPort())
			fmt.Printf("鉴权代理已启动: %s → 127.0.0.1:%s → 127.0.0.1:%s\n", r.Hostname, proxyPort, port)
			cfg.Routes[i].Service = "http://localhost:" + proxyPort
		}
		defer func() {
			for _, p := range proxies {
				p.Stop()
			}
		}()

		if len(cfg.Routes) > 0 {
			client := cfapi.New(cfg.Auth.APIToken, cfg.Auth.AccountID)
			if err := pushIngress(client, context.Background(), cfg); err != nil {
				fmt.Printf("警告: 同步 ingress 失败: %v（将使用远端现有配置）\n", err)
			} else {
				fmt.Println("ingress 配置已同步")
			}
		}

		// 【删除】自动检查更新逻辑
		// 删除了关于 cfg.SelfUpdate.AutoCheck 的整个代码块

		return daemon.Start(cfg.Tunnel.Token)
	},
}

func extractPort(service string) string {
	idx := strings.LastIndex(service, ":")
	if idx < 0 {
		return ""
	}
	return service[idx+1:]
}