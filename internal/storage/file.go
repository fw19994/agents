package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const defaultDataDir = "data"

// HistoryRecord 单条翻译历史
type HistoryRecord struct {
	SessionID string `json:"session_id"`
	Direction string `json:"direction"`
	Input     string `json:"input"`
	Output    string `json:"output"`
	Timestamp int64  `json:"timestamp"`
}

// Config 本地配置（可落盘）
type Config struct {
	Model       string  `json:"model"`
	Temperature float32 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
}

func dataPath(dir string, name string) string {
	if dir == "" {
		return filepath.Join(defaultDataDir, name)
	}
	return filepath.Join(dir, defaultDataDir, name)
}

func ensureDataDir(dir string) error {
	return os.MkdirAll(filepath.Dir(dataPath(dir, "dummy")), 0755)
}

// SaveHistory 追加一条历史到 data/history.json（dir 为项目根，空表示当前目录）
func SaveHistory(dir string, r HistoryRecord) error {
	if err := os.MkdirAll(dataPath(dir, ""), 0755); err != nil {
		return err
	}
	var list []HistoryRecord
	_ = LoadAllHistory(dir, &list)
	list = append(list, r)
	return saveAllHistory(dir, list)
}

// LoadAllHistory 读取全部历史（内部用）
func LoadAllHistory(dir string, out *[]HistoryRecord) error {
	path := dataPath(dir, "history.json")
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var list []HistoryRecord
	if json.Unmarshal(b, &list) != nil {
		return nil
	}
	*out = list
	return nil
}

func saveAllHistory(dir string, list []HistoryRecord) error {
	path := dataPath(dir, "history.json")
	b, _ := json.MarshalIndent(list, "", "  ")
	return os.WriteFile(path, b, 0644)
}

// LoadHistory 读取历史，limit 条（兼容旧逻辑，不按会话过滤）
func LoadHistory(dir string, limit int, out *[]HistoryRecord) error {
	var list []HistoryRecord
	if err := LoadAllHistory(dir, &list); err != nil {
		return err
	}
	if len(list) > limit {
		list = list[len(list)-limit:]
	}
	*out = list
	return nil
}

// LoadHistoryBySession 按会话 ID 读取历史，limit 条
func LoadHistoryBySession(dir string, sessionID string, limit int, out *[]HistoryRecord) error {
	var list []HistoryRecord
	if err := LoadAllHistory(dir, &list); err != nil {
		return err
	}
	var filtered []HistoryRecord
	for _, r := range list {
		if r.SessionID == sessionID {
			filtered = append(filtered, r)
		}
	}
	if limit > 0 && len(filtered) > limit {
		filtered = filtered[len(filtered)-limit:]
	}
	*out = filtered
	return nil
}

// SaveConfig 保存配置到 data/config.json
func SaveConfig(dir string, c Config) error {
	if err := os.MkdirAll(dataPath(dir, ""), 0755); err != nil {
		return err
	}
	path := dataPath(dir, "config.json")
	b, _ := json.MarshalIndent(c, "", "  ")
	return os.WriteFile(path, b, 0644)
}

// LoadConfig 读取配置
func LoadConfig(dir string, out *Config) error {
	path := dataPath(dir, "config.json")
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}
