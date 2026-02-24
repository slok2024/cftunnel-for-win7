package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version     int               `yaml:"version"`
	Auth        AuthConfig        `yaml:"auth"`
	Tunnel      TunnelConfig      `yaml:"tunnel"`
	Routes      []RouteConfig     `yaml:"routes"`
	Cloudflared CloudflaredConfig `yaml:"cloudflared"`
	SelfUpdate  SelfUpdateConfig  `yaml:"self_update"`
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
	Name        string `yaml:"name"`
	Hostname    string `yaml:"hostname"`
	Service     string `yaml:"service"`
	ZoneID      string `yaml:"zone_id"`
	DNSRecordID string `yaml:"dns_record_id"`
}

type CloudflaredConfig struct {
	Path       string `yaml:"path"`
	AutoUpdate bool   `yaml:"auto_update"`
}

type SelfUpdateConfig struct {
	AutoCheck bool `yaml:"auto_check"` 
}

// Dir 修改：返回程序所在目录
func Dir() string {
	exePath, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exePath)
}

// Path 修改：config.yml 就在程序旁边
func Path() string {
	return filepath.Join(Dir(), "config.yml")
}

// 新增：专门给 cloudflared 使用的路径，直接指向同级目录下的 exe
func CloudflaredBinPath() string {
	// 这里直接返回同级目录下的 cloudflared.exe 路径
	// 这样就跳过了创建 bin 文件夹的逻辑
	return filepath.Join(Dir(), "cloudflared.exe")
}

func Load() (*Config, error) {
	data, err := os.ReadFile(Path())
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				Version: 1,
				Cloudflared: CloudflaredConfig{
					Path:       CloudflaredBinPath(), // 默认路径设为当前目录
					AutoUpdate: false,
				},
				SelfUpdate: SelfUpdateConfig{
					AutoCheck: false,
				},
			}, nil
		}
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	
	// 强制禁用更新，并修正路径为当前目录
	cfg.Cloudflared.AutoUpdate = false
	cfg.SelfUpdate.AutoCheck = false
	if cfg.Cloudflared.Path == "" || filepath.Dir(cfg.Cloudflared.Path) != Dir() {
		cfg.Cloudflared.Path = CloudflaredBinPath()
	}
	
	return &cfg, nil
}

func (c *Config) Save() error {
	// 不再需要 MkdirAll(Dir(), 0755)，因为程序所在目录肯定存在
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(Path(), data, 0644)
}

// ... 保持 FindRoute 和 RemoveRoute 不变 ...
func (c *Config) FindRoute(name string) *RouteConfig {
	for i := range c.Routes {
		if c.Routes[i].Name == name {
			return &c.Routes[i]
		}
	}
	return nil
}

func (c *Config) RemoveRoute(name string) bool {
	for i, r := range c.Routes {
		if r.Name == name {
			// 核心逻辑：将 i 之后的所有元素前移，覆盖掉第 i 个元素
			c.Routes = append(c.Routes[:i], c.Routes[i+1:]...)
			return true
		}
	}
	return false
}
