package agent

import (
	"context"
)

// PriorTurn 本会话内历史一轮（仅服务端根据 storage 注入，json 绑定会忽略）
type PriorTurn struct {
	Direction string
	Input     string
	Output    string
}

// TranslateRequest 翻译请求
type TranslateRequest struct {
	SessionID   string  `json:"session_id"` // 可选，空则服务端会创建新会话并返回
	Direction   string  `json:"direction"`
	Content     string  `json:"content"`
	Model       string  `json:"model"`
	Temperature float32 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
	// PriorTurns 多轮上下文；由 Handler 从本地 history 加载，勿信任客户端提交
	PriorTurns []PriorTurn `json:"-"`
}

// StreamTranslate 流式翻译：按方向选 Skill，使用 Eino ChatModel 流式输出，对每个 chunk 调用 onChunk
func StreamTranslate(ctx context.Context, req TranslateRequest, cfg EinoConfig, onChunk func(string)) error {
	return streamTranslateEino(ctx, req, cfg, onChunk)
}
