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

// Register 将 API 路由注册到 gin 引擎（可选前缀 /{project_path}）
func (h *Handler) Register(r *gin.Engine) {
	var g gin.IRoutes = r
	g = r.Group("/translate-agent")
	g.POST("/api/translate/stream", h.StreamTranslate)
	g.GET("/api/models", h.GetModels)
	g.POST("/api/settings", h.SaveSettings)
	g.GET("/api/evaluate/cases", h.GetEvaluateCases)
	g.POST("/api/evaluate/run", h.RunEvaluate)
	g.POST("/api/sessions", h.CreateSession)
	g.GET("/api/sessions", h.ListSessions)
	g.GET("/api/sessions/:id", h.GetSessionDetail)
	g.DELETE("/api/sessions/:id", h.DeleteSession)
}
