package handler

import (
	"translate-agent/internal/config"

	"github.com/gin-gonic/gin"
)

// Handler 持有配置与数据目录，提供 HTTP 处理能力
type Handler struct {
	Cfg     *config.C
	DataDir string
}

// New 根据配置构造 Handler
func New(cfg *config.C, dataDir string) *Handler {
	return &Handler{Cfg: cfg, DataDir: dataDir}
}

// Register 将 API 路由注册到 gin 引擎
func (h *Handler) Register(r *gin.Engine) {
	r.POST("/api/translate/stream", h.StreamTranslate)
	r.GET("/api/models", h.GetModels)
	r.POST("/api/settings", h.SaveSettings)
	r.GET("/api/evaluate/cases", h.GetEvaluateCases)
	r.POST("/api/evaluate/run", h.RunEvaluate)
	// 会话管理
	r.POST("/api/sessions", h.CreateSession)
	r.GET("/api/sessions", h.ListSessions)
	r.GET("/api/sessions/:id", h.GetSessionDetail)
	r.DELETE("/api/sessions/:id", h.DeleteSession)
}
