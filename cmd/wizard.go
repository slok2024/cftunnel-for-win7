package cmd

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/qingchencloud/cftunnel/internal/authproxy"
	"github.com/qingchencloud/cftunnel/internal/cfapi"
	"github.com/qingchencloud/cftunnel/internal/config"
	"github.com/qingchencloud/cftunnel/internal/daemon"
	"github.com/spf13/cobra"
)

var (
	wizardDomain string
	wizardPort   string
	wizardAuth   string
	wizardName   string
)

func init() {
	wizardCmd.Flags().StringVar(&wizardDomain, "domain", "", "完整域名 (如 chat.example.com)")
	wizardCmd.Flags().StringVar(&wizardPort, "port", "", "本地服务端口")
	wizardCmd.Flags().StringVar(&wizardName, "name", "", "路由名称 (默认使用域名前缀)")
	wizardCmd.Flags().StringVar(&wizardAuth, "auth", "", "密码保护 (格式: 用户名:密码)")
	rootCmd.AddCommand(wizardCmd)
}

var wizardCmd = &cobra.Command{
	Use:   "wizard",
	Short: "交互式引导，一条命令完成全部配置",
	Long: `交互式引导，一条命令完成 Tunnel 创建和路由添加。

如果已有配置，会自动跳过已完成的步骤，用户只需提供新增的路由信息。

示例:
  cftunnel wizard                          # 交互模式
  cftunnel wizard --domain chat.example.com --port 8080  # 命令行模式`,
	RunE: runWizard,
}

func runWizard(cmd *cobra.Command, args []string) error {
	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║     Cloudflare Tunnel 向导            ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println()

	// ============ 第1步: 加载或创建配置 ============
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// ============ 第2步: 配置认证信息 ============
	if cfg.Auth.APIToken == "" || cfg.Auth.AccountID == "" {
		fmt.Println("📋 第1步: 配置 Cloudflare 认证信息")
		fmt.Println("   (如果已有配置，可以直接按回车跳过)")
		fmt.Println()

		var apiToken, accountID string
		err := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().Title("API Token").Value(&apiToken).
					Placeholder("从 https://dash.cloudflare.com/profile/api-tokens 创建"),
				huh.NewInput().Title("Account ID").Value(&accountID).
					Placeholder("32位字符，从 Cloudflare  dashboard 获取"),
			),
		).Run()
		if err != nil {
			return err
		}

		if strings.TrimSpace(apiToken) != "" {
			cfg.Auth.APIToken = strings.TrimSpace(apiToken)
		}
		if strings.TrimSpace(accountID) != "" {
			cfg.Auth.AccountID = strings.TrimSpace(accountID)
		}

		if cfg.Auth.APIToken == "" || cfg.Auth.AccountID == "" {
			return fmt.Errorf("API Token 和 Account ID 不能同时为空")
		}

		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Println("✓ 认证信息已保存")
		fmt.Println()
	}

	// ============ 第3步: 创建 Tunnel (如果不存在) ============
	if cfg.Tunnel.ID == "" {
		fmt.Println("📋 第2步: 创建 Tunnel")
		fmt.Println()

		var tunnelName string
		err := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().Title("Tunnel 名称").Value(&tunnelName).
					Placeholder("如: my-tunnel"),
			),
		).Run()
		if err != nil {
			return err
		}

		tunnelName = strings.TrimSpace(tunnelName)
		if tunnelName == "" {
			return fmt.Errorf("Tunnel 名称不能为空")
		}

		fmt.Printf("正在创建 Tunnel: %s\n", tunnelName)
		client := cfapi.New(cfg.Auth.APIToken, cfg.Auth.AccountID)
		ctx := context.Background()

		tunnel, err := client.CreateTunnel(ctx, tunnelName)
		if err != nil {
			return fmt.Errorf("创建 Tunnel 失败: %w", err)
		}

		// 获取 tunnel token
		tunnelToken, err := client.GetTunnelToken(ctx, tunnel.ID)
		if err != nil {
			return fmt.Errorf("获取 Tunnel Token 失败: %w", err)
		}

		cfg.Tunnel = config.TunnelConfig{
			ID:    tunnel.ID,
			Name:  tunnelName,
			Token: tunnelToken,
		}
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf("✓ Tunnel 创建成功: %s\n", tunnelName)
		fmt.Println()
	} else {
		fmt.Printf("✓ 已有 Tunnel: %s (%s)\n", cfg.Tunnel.Name, cfg.Tunnel.ID)
		fmt.Println()
	}

	// ============ 第4步: 添加路由 ============
	client := cfapi.New(cfg.Auth.APIToken, cfg.Auth.AccountID)
	ctx := context.Background()

	// 启动 tunnel（如果未运行）
	if !daemon.Running() {
		fmt.Println("📋 第3步: 启动 Tunnel")
		go daemon.Start(cfg.Tunnel.Token)
		fmt.Println("✓ Tunnel 已启动")
		fmt.Println()
	}

	// 获取域名和端口
	domain := strings.TrimSpace(wizardDomain)
	port := strings.TrimSpace(wizardPort)
	routeName := strings.TrimSpace(wizardName)

	if domain == "" || port == "" {
		fmt.Println("📋 第4步: 添加路由")
		fmt.Println()

		err := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().Title("域名").Value(&domain).
					Placeholder("如: chat.example.com"),
				huh.NewInput().Title("本地端口").Value(&port).
					Placeholder("如: 8080"),
				huh.NewInput().Title("路由名称(可选)").Value(&routeName).
					Placeholder("默认使用域名前缀"),
			),
		).Run()
		if err != nil {
			return err
		}
	}

	domain = strings.TrimSpace(domain)
	port = strings.TrimSpace(port)
	routeName = strings.TrimSpace(routeName)

	if domain == "" || port == "" {
		return fmt.Errorf("域名和端口不能为空")
	}

	// 生成路由名称
	if routeName == "" {
		// 从域名提取前缀，如 chat.example.com -> chat
		parts := strings.Split(domain, ".")
		if len(parts) >= 2 {
			routeName = parts[0]
		} else {
			routeName = domain
		}
	}

	// 检查路由是否已存在
	if cfg.FindRoute(routeName) != nil {
		return fmt.Errorf("路由 %s 已存在", routeName)
	}

	service := "http://localhost:" + port

	fmt.Printf("正在添加路由: %s -> %s\n", domain, service)

	// 查找 Zone
	zone, err := findZoneForDomain(client, ctx, domain)
	if err != nil {
		return err
	}

	// 创建 DNS CNAME 记录
	target := cfg.Tunnel.ID + ".cfargotunnel.com"
	fmt.Printf("正在创建 DNS 记录: %s -> %s\n", domain, target)
	recordID, err := client.CreateCNAME(ctx, zone.ID, domain, target)
	if err != nil {
		return err
	}

	// 构建路由配置
	route := config.RouteConfig{
		Name:        routeName,
		Hostname:    domain,
		Service:     service,
		ZoneID:      zone.ID,
		DNSRecordID: recordID,
	}

	// 密码保护
	if wizardAuth != "" {
		user, pass, err := parseAuth(wizardAuth)
		if err != nil {
			return err
		}
		route.Auth = &config.AuthProxy{
			Username:   user,
			Password:   pass,
			SigningKey: hex.EncodeToString(authproxy.RandomKey()),
		}
		fmt.Printf("✓ 已启用密码保护: %s\n", wizardAuth)
	}

	// 保存路由
	cfg.Routes = append(cfg.Routes, route)
	if err := cfg.Save(); err != nil {
		return err
	}

	// 推送 ingress
	fmt.Println("正在同步 ingress 配置...")
	if err := pushIngress(client, ctx, cfg); err != nil {
		return fmt.Errorf("推送 ingress 失败: %w", err)
	}

	fmt.Println()
	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║            ✅ 全部完成!                 ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Printf("路由已添加: %s -> %s\n", domain, service)
	fmt.Printf("外网访问: https://%s\n", domain)
	fmt.Println()
	fmt.Println("提示: 使用 cftunnel status 查看状态")

	return nil
}
