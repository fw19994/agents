package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Config 调用配置
type Config struct {
	APIKey      string
	BaseURL     string // 如 https://api.openai.com/v1
	Model       string
	Temperature float32
	MaxTokens   int
}

// StreamCallback 流式回调，每次收到一段 content 调用
type StreamCallback func(chunk string)

// Client LLM 客户端（兼容 OpenAI 格式的 chat completions stream）
type Client struct {
	cfg Config
	cli *http.Client
}

func NewClient(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}
	return &Client{cfg: cfg, cli: &http.Client{}}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatReq struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Stream      bool          `json:"stream"`
	Temperature float32       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

type streamDelta struct {
	Content string `json:"content,omitempty"`
}

type streamChunk struct {
	Choices []struct {
		Delta streamDelta `json:"delta"`
	} `json:"choices"`
}

// StreamChat 流式对话，对每个 chunk 调用 fn
func (c *Client) StreamChat(system, user string, fn StreamCallback) error {
	body := chatReq{
		Model: c.cfg.Model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Stream:      true,
		Temperature: c.cfg.Temperature,
		MaxTokens:   c.cfg.MaxTokens,
	}
	if body.MaxTokens <= 0 {
		body.MaxTokens = 2048
	}
	raw, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, c.cfg.BaseURL+"/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)

	resp, err := c.cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("llm api: %s %s", resp.Status, string(b))
	}

	sc := bufio.NewScanner(resp.Body)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}
		var chunk streamChunk
		if json.Unmarshal([]byte(data), &chunk) != nil {
			continue
		}
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			fn(chunk.Choices[0].Delta.Content)
		}
	}
	return sc.Err()
}
