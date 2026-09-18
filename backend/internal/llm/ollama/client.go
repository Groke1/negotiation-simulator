package ollama

import (
	"context"
	"net/http"
	"strings"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
	pkghttpclient "sourcecraft.dev/lct-hackaton-2026.arena/pkg/httpclient"
)

type Client struct {
	httpClient *http.Client
	endpoint   string
	model      string
}

type Config struct {
	BaseURL string
	Model   string
}
type request struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
	Think    bool      `json:"think"`
	Stream   bool      `json:"stream"`
}
type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type response struct {
	Model   string  `json:"model"`
	Message message `json:"message"`
}

func NewClient(httpClient *http.Client, cfg Config) *Client {
	return &Client{httpClient: httpClient, endpoint: pkghttpclient.NormalizeAddr(cfg.BaseURL) + "/api/chat", model: cfg.Model}
}

func (c *Client) Chat(ctx context.Context, input entity.LLMChat) (*entity.LLMResponse, error) {
	model := strings.TrimSpace(input.Model)
	if model == "" {
		model = c.model
	}
	result, err := pkghttpclient.SendAndFetchJSON[request, response](ctx, c.httpClient, http.MethodPost, c.endpoint, request{Model: model, Messages: []message{{Role: "user", Content: input.Prompt}}, Think: false, Stream: false}, pkghttpclient.Options{})
	if err != nil {
		return nil, entity.ErrLLMUnavailable
	}
	if result == nil || strings.TrimSpace(result.Message.Content) == "" {
		return nil, entity.ErrLLMResponseInvalid
	}
	return &entity.LLMResponse{Content: result.Message.Content, Model: result.Model}, nil
}
