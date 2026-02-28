package config

import (
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version     int               `yaml:"version"`
	Auth        AuthConfig        `yaml:"auth"`
	Tunnel      TunnelConfig      `yaml:"tunnel"`
	Routes      []RouteConfig     `yaml:"routes"`
	Relay       RelayConfig       `yaml:"relay,omitempty"`
	Cloudflared CloudflaredConfig `yaml:"cloudflared"`
}

type AuthConfig struct {
	APIToken  string `yaml:"api_token"`
	AccountID string `yaml:"account_id"`
}

type TunnelConfig struct {
	ID    string `yaml:"id"`
	Name  string `yaml:"name"`
	Token string `yaml:"token"`
}

type RouteConfig struct {
	Name        string     `yaml:"name"`
	Hostname    string     `yaml:"hostname"`
	Service     string     `yaml:"service"`
	ZoneID      string     `yaml:"zone_id"`
	DNSRecordID string     `yaml:"dns_record_id"`
	Auth        *AuthProxy `yaml:"auth,omitempty"`
}

type AuthProxy struct {
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	SigningKey string `yaml:"signing_key,omitempty"`
	CookieTTL  int    `yaml:"cookie_ttl,omitempty"`
}

func (a *AuthProxy) CookieTTLOrDefault() int {
	if a.CookieTTL > 0 {
		return a.CookieTTL
	}
	return 86400
}

type RelayConfig struct {
	Server string      `yaml:"server,omitempty"`
	Token  string      `yaml:"token,omitempty"`
	Rules  []RelayRule `yaml:"rules,omitempty"`
}

type RelayRule struct {
	Name       string `yaml:"name"`
	Proto      string `yaml:"proto"`
	LocalIP    string `yaml:"local_ip,omitempty"`
	LocalPort  int    `yaml:"local_port"`
	RemotePort int    `yaml:"remote_port,omitempty"`
	Domain     string `yaml:"domain,omitempty"`
}

type CloudflaredConfig struct {
	Path string `yaml:"path"`
}

var (
	dirOnce sync.Once
	dirPath string
)

// Dir 强制只返回程序当前所在的目录
func Dir() string {
	dirOnce.Do(func() {
		exe, err := os.Executable()
		if err == nil {
			// 关键：先转绝对路径，再取目录
			res, _ := filepath.Abs(exe)
			dirPath = filepath.Dir(res)
		}
		// 如果获取失败，再次尝试用 Args[0] 兜底
		if dirPath == "" || dirPath == "." {
			p, _ := filepath.Abs(os.Args[0])
			dirPath = filepath.Dir(p)
		}
	})
	return dirPath
}

// Portable 既然只看当前目录，那么永远是便携模式
func Portable() bool {
	return true
}

func Path() string {
	return filepath.Join(Dir(), "config.yml")
}

func Load() (*Config, error) {
	data, err := os.ReadFile(Path())
	if err != nil {
		if os.IsNotExist(err) {
			// 如果没找到配置文件，返回一个空的
			return &Config{Version: 1}, nil
		}
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Save() error {
	// 即使是在当前目录，也确保路径合法（虽然通常 exe 目录肯定存在）
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(Path(), data, 0600)
}

// --- 保持后续 Find/Remove 逻辑不变 ---
func (c *Config) FindRoute(name string) *RouteConfig {
	for i := range c.Routes {
		if c.Routes[i].Name == name { return &c.Routes[i] }
	}
	return nil
}

func (c *Config) RemoveRoute(name string) bool {
	for i, r := range c.Routes {
		if r.Name == name {
			c.Routes = append(c.Routes[:i], c.Routes[i+1:]...)
			return true
		}
	}
	return false
}

func (c *Config) FindRelayRule(name string) *RelayRule {
	for i := range c.Relay.Rules {
		if c.Relay.Rules[i].Name == name { return &c.Relay.Rules[i] }
	}
	return nil
}

func (c *Config) RemoveRelayRule(name string) bool {
	for i, r := range c.Relay.Rules {
		if r.Name == name {
			c.Relay.Rules = append(c.Relay.Rules[:i], c.Relay.Rules[i+1:]...)
			return true
		}
	}
	return false
}