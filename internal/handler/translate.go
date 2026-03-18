package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"translate-agent/internal/agent"
	"translate-agent/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// StreamTranslate 流式翻译：POST /api/translate/stream
func (h *Handler) StreamTranslate(c *gin.Context) {
	var req agent.TranslateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 无 session_id 时创建新会话
	if req.SessionID == "" {
		req.SessionID = uuid.New().String()
		_ = storage.SaveSession(h.DataDir, storage.Session{
			ID:        req.SessionID,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		})
	}

	einoCfg := h.resolveEinoConfig(req.Model, req.Temperature, req.MaxTokens)
	if einoCfg.Model == "" {
		einoCfg.Model = "gpt-4o-mini"
	}
	if einoCfg.MaxTokens <= 0 {
		einoCfg.MaxTokens = 2048
	}

	// 本会话历史 → 注入 Agent 多轮上下文（仅服务端加载，防伪造）
	const maxPriorTurns = 30
	var hist []storage.HistoryRecord
	_ = storage.LoadHistoryBySession(h.DataDir, req.SessionID, maxPriorTurns, &hist)
	req.PriorTurns = make([]agent.PriorTurn, 0, len(hist))
	for _, r := range hist {
		req.PriorTurns = append(req.PriorTurns, agent.PriorTurn{
			Direction: r.Direction,
			Input:     r.Input,
			Output:    r.Output,
		})
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	var full strings.Builder
	req.Model = einoCfg.Model
	req.Temperature = einoCfg.Temperature
	req.MaxTokens = einoCfg.MaxTokens
	err := agent.StreamTranslate(c.Request.Context(), req, einoCfg, func(chunk string) {
		full.WriteString(chunk)
		payload, _ := json.Marshal(map[string]interface{}{
			"choices": []map[string]interface{}{
				{"delta": map[string]string{"content": chunk}},
			},
		})
		fmt.Fprintf(c.Writer, "data: %s\n\n", payload)
		flusher.Flush()
	})
	if err != nil {
		payload, _ := json.Marshal(map[string]string{
			"error":  agent.FormatErrorChain(err),
			"source": "agent",
		})
		fmt.Fprintf(c.Writer, "data: %s\n\n", payload)
		flusher.Flush()
		return
	}
	// 落盘历史并更新会话时间；新会话时用首条输入做标题
	title := ""
	if len(req.Content) > 50 {
		title = req.Content[:50] + "…"
	} else {
		title = req.Content
	}
	_ = storage.SaveHistory(h.DataDir, storage.HistoryRecord{
		SessionID: req.SessionID,
		Direction: req.Direction,
		Input:     req.Content,
		Output:    full.String(),
		Timestamp: time.Now().Unix(),
	})
	_ = storage.SaveSession(h.DataDir, storage.Session{
		ID: req.SessionID, Title: title, UpdatedAt: time.Now().Unix(),
	})
	// 流式结束后返回 session_id，便于前端绑定当前会话
	payload, _ := json.Marshal(map[string]string{"session_id": req.SessionID})
	fmt.Fprintf(c.Writer, "data: %s\n\n", payload)
	flusher.Flush()
}

func (h *Handler) resolveEinoConfig(model string, temperature float32, maxTokens int) agent.EinoConfig {
	apiKey, baseURL := h.Cfg.GetProviderForModel(model)
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			apiKey = os.Getenv("API_KEY")
		}
	}
	return agent.EinoConfig{
		APIKey:      apiKey,
		BaseURL:     baseURL,
		Model:       model,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}
}
