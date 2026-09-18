package llm

import (
	"context"
	"strings"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
)

type provider interface {
	Chat(context.Context, entity.LLMChat) (*entity.LLMResponse, error)
}
type Service struct{ provider provider }

func NewService(provider provider) *Service { return &Service{provider: provider} }
func (s *Service) Chat(ctx context.Context, input entity.LLMChat) (*entity.LLMResponse, error) {
	if len(strings.TrimSpace(input.Prompt)) == 0 || len(input.Prompt) > 20000 {
		return nil, entity.ErrInvalidRequest
	}
	return s.provider.Chat(ctx, input)
}
