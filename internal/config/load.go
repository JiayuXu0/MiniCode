package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Load 加载配置
// 按以下顺序加载并合并:
// 1. 默认配置
// 2. 全局配置 (~/.config/minicode/config.json)
// 3. 项目配置 (向上查找 minicode.json)
func Load() (*Config, error) {
	// 1. 从默认配置开始
	cfg := DefaultConfig()

	// 2. 加载全局配置
	globalPath := globalConfigPath()
	if globalPath != "" {
		if err := loadAndMerge(cfg, globalPath); err != nil {
			// 全局配置可选，忽略错误
		}
	}

	// 3. 加载项目配置
	projectPaths := findProjectConfigs()
	for _, path := range projectPaths {
		if err := loadAndMerge(cfg, path); err != nil {
			return nil, err
		}
	}

	// 4. 解析变量
	resolveConfig(cfg)

	return cfg, nil
}

// globalConfigPath 返回全局配置路径
func globalConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "minicode", "config.json")
}

// findProjectConfigs 从当前目录向上查找配置文件
func findProjectConfigs() []string {
	var configs []string

	cwd, err := os.Getwd()
	if err != nil {
		return configs
	}

	dir := cwd
	for {
		// 检查 minicode.json
		path := filepath.Join(dir, "minicode.json")
		if _, err := os.Stat(path); err == nil {
			// 插入到开头（父目录的配置先加载）
			configs = append([]string{path}, configs...)
		}

		// 检查 .minicode.json (隐藏文件)
		path = filepath.Join(dir, ".minicode.json")
		if _, err := os.Stat(path); err == nil {
			configs = append([]string{path}, configs...)
		}

		// 到达根目录，停止
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return configs
}

// loadAndMerge 加载配置文件并合并到现有配置
func loadAndMerge(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var fileCfg Config
	if err := json.Unmarshal(data, &fileCfg); err != nil {
		return err
	}

	// 合并 Providers
	for name, provider := range fileCfg.Providers {
		cfg.Providers[name] = provider
	}

	// 合并 Models
	if fileCfg.Models.Default != "" {
		cfg.Models.Default = fileCfg.Models.Default
	}
	if fileCfg.Models.Small != "" {
		cfg.Models.Small = fileCfg.Models.Small
	}

	// 合并 Options
	if fileCfg.Options.MaxTokens > 0 {
		cfg.Options.MaxTokens = fileCfg.Options.MaxTokens
	}

	return nil
}

// resolveConfig 解析配置中的所有变量
func resolveConfig(cfg *Config) {
	for name, provider := range cfg.Providers {
		provider.APIKey = ResolveVariables(provider.APIKey)
		provider.BaseURL = ResolveVariables(provider.BaseURL)
		provider.Name = ResolveVariables(provider.Name)
		cfg.Providers[name] = provider
	}
}
