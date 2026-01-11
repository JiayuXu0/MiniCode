package config

// Config 是主配置结构
type Config struct {
	Providers map[string]ProviderConfig `json:"providers"`
	Models    ModelConfig               `json:"models"`
	Options   Options                   `json:"options"`
}

// ProviderConfig 定义 LLM 提供者配置
type ProviderConfig struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url,omitempty"`
	Name    string `json:"name,omitempty"`
}

// ModelConfig 定义模型配置
type ModelConfig struct {
	Default string `json:"default"`
	Small   string `json:"small,omitempty"`
}

// Options 定义其他选项
type Options struct {
	MaxTokens int `json:"max_tokens,omitempty"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Providers: make(map[string]ProviderConfig),
		Models: ModelConfig{
			Default: "glm-4.7",
		},
		Options: Options{
			MaxTokens: 4096,
		},
	}
}

// GetProvider 获取指定 Provider 配置
func (c *Config) GetProvider(name string) (ProviderConfig, bool) {
	p, ok := c.Providers[name]
	return p, ok
}

// GetDefaultModel 获取默认模型名称
func (c *Config) GetDefaultModel() string {
	if c.Models.Default != "" {
		return c.Models.Default
	}
	return "glm-4.7"
}
