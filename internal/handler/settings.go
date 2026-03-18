package handler

import (
	"net/http"

	"translate-agent/internal/storage"

	"github.com/gin-gonic/gin"
)

// SaveSettings 保存用户设置：POST /api/settings
func (h *Handler) SaveSettings(c *gin.Context) {
	var body storage.Config
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := storage.SaveConfig(h.DataDir, body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
