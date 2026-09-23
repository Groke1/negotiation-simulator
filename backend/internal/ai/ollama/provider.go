package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/ai"
)

type Config struct {
	BaseURL string
	Timeout time.Duration
	Think   bool
	Stream  bool
}

type Provider struct {
	client *http.Client
	cfg    Config
}

func NewProvider(cfg Config) *Provider {
	return &Provider{client: &http.Client{Timeout: cfg.Timeout}, cfg: cfg}
}

func (p *Provider) Generate(ctx context.Context, req ai.GenerateRequest) (string, error) {
	baseURL := strings.TrimRight(p.cfg.BaseURL, "/")
	if baseURL == "" {
		return "", ai.ErrProviderUnavailable
	}
	payload := chatRequest{
		Model:    req.Model,
		Messages: make([]chatMessage, 0, len(req.Messages)),
		Think:    p.cfg.Think,
		Stream:   p.cfg.Stream,
		Options:  map[string]any{"temperature": req.Temperature},
	}
	for _, msg := range req.Messages {
		payload.Messages = append(payload.Messages, chatMessage{Role: msg.Role, Content: msg.Content})
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusServiceUnavailable || resp.StatusCode == http.StatusBadGateway || resp.StatusCode == http.StatusGatewayTimeout {
		return "", ai.ErrProviderUnavailable
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("ollama status %d", resp.StatusCode)
	}
	var decoded chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", err
	}
	return decoded.Message.Content, nil
}

type chatRequest struct {
	Model    string         `json:"model"`
	Messages []chatMessage  `json:"messages"`
	Think    bool           `json:"think"`
	Stream   bool           `json:"stream"`
	Options  map[string]any `json:"options,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Message chatMessage `json:"message"`
}
