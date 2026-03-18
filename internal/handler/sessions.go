package handler

import (
	"net/http"
	"strconv"
	"time"

	"translate-agent/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateSession 新建会话：POST /api/sessions
func (h *Handler) CreateSession(c *gin.Context) {
	id := uuid.New().String()
	now := time.Now().Unix()
	s := storage.Session{ID: id, CreatedAt: now, UpdatedAt: now}
	if err := storage.SaveSession(h.DataDir, s); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "created_at": now, "title": ""})
}

// ListSessions 会话列表：GET /api/sessions?limit=50
func (h *Handler) ListSessions(c *gin.Context) {
	limit := 50
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	list, err := storage.ListSessions(h.DataDir, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"sessions": list})
}

// GetSessionDetail 会话详情（含历史消息）：GET /api/sessions/:id
func (h *Handler) GetSessionDetail(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing session id"})
		return
	}
	s, err := storage.GetSession(h.DataDir, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if s == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	var messages []storage.HistoryRecord
	_ = storage.LoadHistoryBySession(h.DataDir, id, 0, &messages)
	c.JSON(http.StatusOK, gin.H{"session": s, "messages": messages})
}

// DeleteSession 删除会话：DELETE /api/sessions/:id
func (h *Handler) DeleteSession(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing session id"})
		return
	}
	if err := storage.DeleteSession(h.DataDir, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
