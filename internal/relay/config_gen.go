package relay

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/qingchencloud/cftunnel/internal/config"
)

// FrpcConfigPath 确保返回程序目录下的 frpc.toml
func FrpcConfigPath() string {
	// 既然 config.Dir() 已经改成了程序目录，这里保持一致即可
	return filepath.Join(config.Dir(), "frpc.toml")
}

// FrpsConfigPath 返回 frps.toml 路径
func FrpsConfigPath() string {
	return filepath.Join(config.Dir(), "frps.toml")
}

// GenerateFrpcConfig 生成配置
func GenerateFrpcConfig(relay *config.RelayConfig) error {
	// ... 原有的逻辑非常标准，支持 frp 0.52.0+ 的 TOML 格式 ...
	// 保持不变即可
	if relay.Server == "" {
		return fmt.Errorf("未配置中继服务器")
	}

	host, port, err := net.SplitHostPort(relay.Server)
	if err != nil {
		return fmt.Errorf("服务器地址格式错误: %w", err)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "serverAddr = %q\n", host)
	fmt.Fprintf(&b, "serverPort = %s\n", port)
	if relay.Token != "" {
		// 注意：frp 新版 toml 格式是 auth.token
		fmt.Fprintf(&b, "auth.token = %q\n", relay.Token)
	}
	b.WriteString("\n")

	for _, rule := range relay.Rules {
		b.WriteString("[[proxies]]\n")
		fmt.Fprintf(&b, "name = %q\n", rule.Name)
		fmt.Fprintf(&b, "type = %q\n", rule.Proto)
		localIP := rule.LocalIP
		if localIP == "" {
			localIP = "127.0.0.1"
		}
		fmt.Fprintf(&b, "localIP = %q\n", localIP)
		fmt.Fprintf(&b, "localPort = %d\n", rule.LocalPort)
		if rule.RemotePort > 0 {
			fmt.Fprintf(&b, "remotePort = %d\n", rule.RemotePort)
		}
		if rule.Domain != "" {
			fmt.Fprintf(&b, "customDomains = [%q]\n", rule.Domain)
		}
		b.WriteString("\n")
	}

	// 确保目录存在（虽然在程序目录运行，通常已存在）
	os.MkdirAll(config.Dir(), 0755)
	return os.WriteFile(FrpcConfigPath(), []byte(b.String()), 0600)
}

// GenerateFrpsConfig 保持不变
func GenerateFrpsConfig(bindPort int, token string) error {
	var b strings.Builder
	fmt.Fprintf(&b, "bindPort = %d\n", bindPort)
	if token != "" {
		fmt.Fprintf(&b, "auth.token = %q\n", token)
	}
	os.MkdirAll(config.Dir(), 0755)
	return os.WriteFile(FrpsConfigPath(), []byte(b.String()), 0600)
}