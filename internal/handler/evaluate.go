package handler

import (
	"net/http"
	"strings"
	"time"

	"translate-agent/internal/agent"
	"translate-agent/internal/eval"

	"github.com/gin-gonic/gin"
)

// GetEvaluateCases 返回评测用例列表：GET /api/evaluate/cases
func (h *Handler) GetEvaluateCases(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"cases": eval.DefaultCases})
}

// RunEvaluate 执行评测：POST /api/evaluate/run
func (h *Handler) RunEvaluate(c *gin.Context) {
	var body struct {
		CaseIDs []string `json:"case_ids"`
	}
	_ = c.ShouldBindJSON(&body)
	cases := eval.DefaultCases
	if len(body.CaseIDs) > 0 {
		set := make(map[string]bool)
		for _, id := range body.CaseIDs {
			set[id] = true
		}
		var filtered []eval.Case
		for _, tc := range cases {
			if set[tc.ID] {
				filtered = append(filtered, tc)
			}
		}
		cases = filtered
	}
	if len(cases) == 0 {
		cases = eval.DefaultCases
	}

	einoCfg := h.resolveEinoConfig("gpt-4o-mini", 0.7, 2048)
	start := time.Now()
	var results []eval.Result
	for _, tc := range cases {
		t0 := time.Now()
		var out strings.Builder
		err := agent.StreamTranslate(c.Request.Context(), agent.TranslateRequest{
			Direction:   tc.Direction,
			Content:     tc.Input,
			Model:       einoCfg.Model,
			Temperature: einoCfg.Temperature,
			MaxTokens:   einoCfg.MaxTokens,
		}, einoCfg, func(chunk string) {
			out.WriteString(chunk)
		})
		pass := err == nil && out.Len() > 0
		errStr := ""
		if err != nil {
			errStr = agent.FormatErrorChain(err)
		}
		results = append(results, eval.Result{
			CaseID:     tc.ID,
			Pass:       pass,
			Output:     out.String(),
			Error:      errStr,
			DurationMs: time.Since(t0).Milliseconds(),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"results":     results,
		"duration_ms": time.Since(start).Milliseconds(),
	})
}
