package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
)

type Role string

const (
	RoleScenarioArchitect Role = "scenario_architect"
	RoleBlockDesigner     Role = "block_designer"
	RoleDialogueBuilder   Role = "dialogue_builder"
	RoleActionClassifier  Role = "action_classifier"
	RoleVirtualPartner    Role = "virtual_partner"
)

type Message struct {
	Role    string
	Content string
}

type GenerateRequest struct {
	Model       string
	Messages    []Message
	Temperature float64
}

type Provider interface {
	Generate(ctx context.Context, req GenerateRequest) (string, error)
}

type Config struct {
	PromptDir                string
	ScenarioArchitectModel   string
	BlockDesignerModel       string
	DialogueBuilderModel     string
	ActionClassifierModel    string
	VirtualPartnerModel      string
	ScenarioArchitectTemp    float64
	BlockDesignerTemp        float64
	DialogueBuilderTemp      float64
	ActionClassifierTemp     float64
	VirtualPartnerTemp       float64
	MaxStructuredOutputRetry int
}

type Service struct {
	provider Provider
	cfg      Config
	prompts  map[Role]string
}

func NewService(provider Provider, cfg Config) (*Service, error) {
	prompts, err := loadPrompts(cfg.PromptDir)
	if err != nil {
		return nil, err
	}
	if cfg.MaxStructuredOutputRetry < 0 {
		cfg.MaxStructuredOutputRetry = 0
	}
	return &Service{provider: provider, cfg: cfg, prompts: prompts}, nil
}

func (s *Service) GenerateJSON(ctx context.Context, role Role, input any, output any, validate func() error) error {
	if s == nil || s.provider == nil {
		return entity.ErrAIProviderUnavailable
	}
	payload, err := json.Marshal(input)
	if err != nil {
		return err
	}
	model, temperature, err := s.modelForRole(role)
	if err != nil {
		return err
	}
	systemPrompt, ok := s.prompts[role]
	if !ok || strings.TrimSpace(systemPrompt) == "" {
		return fmt.Errorf("ai prompt for role %s is missing", role)
	}
	var lastErr error
	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: string(payload)},
	}
	for attempt := 0; attempt <= s.cfg.MaxStructuredOutputRetry; attempt++ {
		if attempt > 0 && lastErr != nil {
			messages = append(messages, Message{Role: "user", Content: "Previous response did not pass backend JSON/schema validation: " + lastErr.Error() + ". Return only corrected JSON."})
		}
		content, err := s.provider.Generate(ctx, GenerateRequest{Model: model, Temperature: temperature, Messages: messages})
		if err != nil {
			return mapProviderError(err)
		}
		if err := decodeModelJSON(content, output); err != nil {
			lastErr = err
			continue
		}
		if validate != nil {
			if err := validate(); err != nil {
				lastErr = err
				continue
			}
		}
		return nil
	}
	if lastErr != nil {
		return fmt.Errorf("%w: %v", entity.ErrAIInvalidStructuredOutput, lastErr)
	}
	return entity.ErrAIInvalidStructuredOutput
}

func (s *Service) modelForRole(role Role) (string, float64, error) {
	switch role {
	case RoleScenarioArchitect:
		return s.cfg.ScenarioArchitectModel, s.cfg.ScenarioArchitectTemp, nil
	case RoleBlockDesigner:
		return s.cfg.BlockDesignerModel, s.cfg.BlockDesignerTemp, nil
	case RoleDialogueBuilder:
		return s.cfg.DialogueBuilderModel, s.cfg.DialogueBuilderTemp, nil
	case RoleActionClassifier:
		return s.cfg.ActionClassifierModel, s.cfg.ActionClassifierTemp, nil
	case RoleVirtualPartner:
		return s.cfg.VirtualPartnerModel, s.cfg.VirtualPartnerTemp, nil
	default:
		return "", 0, fmt.Errorf("unknown ai role %s", role)
	}
}

func loadPrompts(dir string) (map[Role]string, error) {
	if strings.TrimSpace(dir) == "" {
		dir = "prompts"
	}
	files := map[Role]string{
		RoleScenarioArchitect: "01_scenario_architect.system.md",
		RoleBlockDesigner:     "02_block_designer.system.md",
		RoleDialogueBuilder:   "03_deterministic_dialogue_builder.system.md",
		RoleActionClassifier:  "04_semantic_action_classifier.system.md",
		RoleVirtualPartner:    "05_virtual_partner_turn.system.md",
	}
	prompts := make(map[Role]string, len(files))
	for role, name := range files {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		prompts[role] = string(data)
	}
	return prompts, nil
}

func decodeModelJSON(content string, output any) error {
	content = strings.TrimSpace(content)
	if content == "" {
		return fmt.Errorf("empty ai response")
	}
	if !strings.HasPrefix(content, "{") || !strings.HasSuffix(content, "}") {
		return fmt.Errorf("ai response must be a single JSON object without markdown or surrounding text")
	}
	decoder := json.NewDecoder(bytes.NewBufferString(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(output); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("ai response contains extra JSON tokens")
	}
	return nil
}

func mapProviderError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, os.ErrDeadlineExceeded) {
		return entity.ErrAIGenerationTimeout
	}
	if errors.Is(err, ErrProviderUnavailable) {
		return entity.ErrAIProviderUnavailable
	}
	var timeout interface{ Timeout() bool }
	if errors.As(err, &timeout) && timeout.Timeout() {
		return entity.ErrAIGenerationTimeout
	}
	return fmt.Errorf("%w: %v", entity.ErrAIGenerationFailed, err)
}

var ErrProviderUnavailable = errors.New("provider unavailable")

func TimeoutContext(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, timeout)
}
