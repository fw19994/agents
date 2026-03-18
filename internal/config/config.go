package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Env 当前环境名
var Env string

// Provider 单个模型供应商配置
type Provider struct {
	APIKey  string   `json:"api_key"`
	BaseURL string   `json:"base_url"`
	Models  []string `json:"models"` // 该供应商下的模型 ID 列表
}

// C 当前环境配置
type C struct {
	OpenAIAPIKey string `json:"openai_api_key"` // 兼容旧配置，无 providers 时使用
	LLMBaseURL   string `json:"llm_base_url"`
	Addr         string `json:"addr"`
	DataDir      string `json:"data_dir"`
	// ProjectPath 路由前缀（项目名），如 translate-agent → /translate-agent/api/... ；空则仍挂在根路径
	ProjectPath string              `json:"project_path"`
	Providers   map[string]Provider `json:"providers"` // 多供应商：名称 -> 配置
}

// configFile 结构：按环境名取配置
type configFile struct {
	Dev  C `json:"dev"`
	Prod C `json:"prod"`
}

// Load 根据环境变量 APP_ENV（dev/prod，默认 dev）加载配置。
// 配置文件路径：工作目录下的 config/config.json
func Load() (C, error) {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}
	Env = env

	path := filepath.Join("config", "config.json")
	b, err := os.ReadFile(path)
	if err != nil {
		// 无配置文件时返回默认值，API Key 为空
		return C{Addr: ":9001", DataDir: ".", LLMBaseURL: "https://api.openai.com/v1"}, nil
	}

	var f configFile
	if err := json.Unmarshal(b, &f); err != nil {
		return C{}, err
	}

	var c C
	switch env {
	case "prod", "production":
		c = f.Prod
	default:
		c = f.Dev
	}

	if c.Addr == "" {
		c.Addr = ":9001"
	}
	if c.DataDir == "" {
		c.DataDir = "."
	}
	if c.LLMBaseURL == "" {
		c.LLMBaseURL = "https://api.openai.com/v1"
	}
	if v := strings.TrimSpace(os.Getenv("PROJECT_PATH")); v != "" {
		c.ProjectPath = v
	}
	return c, nil
}

// HTTPRoutePrefix 返回 Gin 路由前缀，如 "/translate-agent"；空表示根路径
func (c *C) HTTPRoutePrefix() string {
	s := strings.TrimSpace(c.ProjectPath)
	s = strings.Trim(s, "/")
	if s == "" {
		return ""
	}
	return "/" + s
}

// GetProviderForModel 根据模型 ID 返回该模型所属供应商的 api_key 和 base_url。
// 若配置了 providers，则按各 provider 的 models 列表匹配；否则使用全局 openai_api_key / llm_base_url。
func (c *C) GetProviderForModel(model string) (apiKey, baseURL string) {
	if model == "" {
		model = "gpt-4o-mini"
	}
	if len(c.Providers) > 0 {
		for _, p := range c.Providers {
			for _, m := range p.Models {
				if m == model {
					apiKey = p.APIKey
					baseURL = p.BaseURL
					if baseURL == "" {
						baseURL = "https://api.openai.com/v1"
					}
					return
				}
			}
		}
	}
	apiKey = c.OpenAIAPIKey
	baseURL = c.LLMBaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return
}

// AllModels 返回当前配置下所有可用模型 ID 列表（来自各 provider 的 models 并集）。
// 若未配置 providers，返回默认列表。
func (c *C) AllModels() []string {
	if len(c.Providers) == 0 {
		return []string{"gpt-4o-mini", "gpt-4o", "claude-3-5-sonnet", "qwen-plus"}
	}
	seen := make(map[string]bool)
	var list []string
	for _, p := range c.Providers {
		for _, m := range p.Models {
			if !seen[m] {
				seen[m] = true
				list = append(list, m)
			}
		}
	}
	return list
}
