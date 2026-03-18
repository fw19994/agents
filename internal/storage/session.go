package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Session 会话元数据
type Session struct {
	ID        string `json:"id"`
	Title     string `json:"title"` // 可选，如首条输入摘要
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

const sessionsFile = "sessions.json"

func sessionsPath(dir string) string {
	return dataPath(dir, sessionsFile)
}

// SaveSession 新增或更新会话（存在则只更新 UpdatedAt/Title）
func SaveSession(dir string, s Session) error {
	if err := os.MkdirAll(filepath.Dir(dataPath(dir, "")), 0755); err != nil {
		return err
	}
	path := sessionsPath(dir)
	var list []Session
	_ = loadSessions(dir, &list)
	now := time.Now().Unix()
	if s.CreatedAt == 0 {
		s.CreatedAt = now
	}
	s.UpdatedAt = now
	found := false
	for i := range list {
		if list[i].ID == s.ID {
			list[i].UpdatedAt = s.UpdatedAt
			if s.Title != "" {
				list[i].Title = s.Title
			}
			found = true
			break
		}
	}
	if !found {
		list = append(list, s)
	}
	b, _ := json.MarshalIndent(list, "", "  ")
	return os.WriteFile(path, b, 0644)
}

func loadSessions(dir string, out *[]Session) error {
	path := sessionsPath(dir)
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(b, out)
}

// ListSessions 返回会话列表，按 UpdatedAt 倒序，最多 limit 条
func ListSessions(dir string, limit int) ([]Session, error) {
	var list []Session
	if err := loadSessions(dir, &list); err != nil {
		return nil, err
	}
	// 简单倒序：按 UpdatedAt 降序
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}
	if limit > 0 && len(list) > limit {
		list = list[:limit]
	}
	return list, nil
}

// GetSession 获取单条会话元数据
func GetSession(dir string, id string) (*Session, error) {
	var list []Session
	if err := loadSessions(dir, &list); err != nil {
		return nil, err
	}
	for i := range list {
		if list[i].ID == id {
			return &list[i], nil
		}
	}
	return nil, nil
}

// DeleteSession 删除会话及其历史记录
func DeleteSession(dir string, id string) error {
	var list []Session
	if err := loadSessions(dir, &list); err != nil {
		return err
	}
	var newList []Session
	for _, s := range list {
		if s.ID != id {
			newList = append(newList, s)
		}
	}
	path := sessionsPath(dir)
	b, _ := json.MarshalIndent(newList, "", "  ")
	if err := os.WriteFile(path, b, 0644); err != nil {
		return err
	}
	return deleteHistoryBySession(dir, id)
}

// deleteHistoryBySession 删除某会话下所有历史记录
func deleteHistoryBySession(dir string, sessionID string) error {
	var list []HistoryRecord
	if err := LoadAllHistory(dir, &list); err != nil {
		return err
	}
	var kept []HistoryRecord
	for _, r := range list {
		if r.SessionID != sessionID {
			kept = append(kept, r)
		}
	}
	return saveAllHistory(dir, kept)
}
