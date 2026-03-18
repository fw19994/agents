package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetModels 返回可用模型列表：GET /api/models
func (h *Handler) GetModels(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"models": h.Cfg.AllModels()})
}
