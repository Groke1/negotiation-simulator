package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
)

type Role string

const (
	RoleScenarioArchitect     Role = "scenario_architect"
	RoleBlockDesigner         Role = "block_designer"
	RoleDialogueBuilder       Role = "dialogue_builder"
	RoleActionClassifier      Role = "action_classifier"
	RoleVirtualPartner        Role = "virtual_partner"
	RoleVirtualPartnerRuntime Role = "virtual_partner_runtime"
	RoleTrainingState         Role = "training_state"
	RoleTrainingAssessment    Role = "training_assessment"
)

type message struct {
	Role    string
	Content string
}

type GenerateRequest struct {
	Model       string
	Messages    []message
	Temperature float64
	Think       *bool
}

type Provider interface {
	Generate(ctx context.Context, req GenerateRequest) (string, error)
}

type Config struct {
	ScenarioArchitectModel     string
	BlockDesignerModel         string
	DialogueBuilderModel       string
	ActionClassifierModel      string
	VirtualPartnerModel        string
	VirtualPartnerRuntimeModel string
	TrainingStateModel         string
	TrainingAssessmentModel    string
	ScenarioArchitectTemp      float64
	BlockDesignerTemp          float64
	DialogueBuilderTemp        float64
	ActionClassifierTemp       float64
	VirtualPartnerTemp         float64
	VirtualPartnerRuntimeTemp  float64
	TrainingStateTemp          float64
	TrainingAssessmentTemp     float64
	BlockDesignerThink         *bool
	DialogueBuilderThink       *bool
	VirtualPartnerRuntimeThink *bool
	TrainingStateThink         *bool
	TrainingAssessmentThink    *bool
	MaxStructuredOutputRetry   int
}

type service struct {
	provider Provider
	cfg      Config
}

func NewService(provider Provider, cfg Config) (*service, error) {

	cfg.MaxStructuredOutputRetry = max(0, cfg.MaxStructuredOutputRetry)

	return &service{provider: provider, cfg: cfg}, nil
}

func (s *service) GenerateJSON(ctx context.Context, role Role, input any, output any, validate func() error) error {
	if s == nil || s.provider == nil {
		return entity.ErrAIProviderUnavailable
	}
	payload, err := json.Marshal(input)
	if err != nil {
		return err
	}
	roleCfg, err := s.configForRole(role)
	if err != nil {
		return err
	}

	var lastErr error
	messages := []message{
		{Role: "user", Content: string(payload)},
	}
	for attempt := 0; attempt <= s.cfg.MaxStructuredOutputRetry; attempt++ {
		if attempt > 0 && lastErr != nil {
			messages = append(messages, message{Role: "user", Content: "Previous response did not pass backend JSON/schema validation: " + lastErr.Error() + ". Return only corrected JSON."})
		}
		content, err := s.provider.Generate(ctx, GenerateRequest{Model: roleCfg.model, Temperature: roleCfg.temperature, Think: roleCfg.think, Messages: messages})
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

type roleConfig struct {
	model       string
	temperature float64
	think       *bool
}

func (s *service) configForRole(role Role) (roleConfig, error) {
	switch role {
	case RoleScenarioArchitect:
		return roleConfig{model: s.cfg.ScenarioArchitectModel, temperature: s.cfg.ScenarioArchitectTemp}, nil
	case RoleBlockDesigner:
		return roleConfig{model: s.cfg.BlockDesignerModel, temperature: s.cfg.BlockDesignerTemp, think: s.cfg.BlockDesignerThink}, nil
	case RoleDialogueBuilder:
		return roleConfig{model: s.cfg.DialogueBuilderModel, temperature: s.cfg.DialogueBuilderTemp, think: s.cfg.DialogueBuilderThink}, nil
	case RoleActionClassifier:
		return roleConfig{model: s.cfg.ActionClassifierModel, temperature: s.cfg.ActionClassifierTemp}, nil
	case RoleVirtualPartner:
		return roleConfig{model: s.cfg.VirtualPartnerModel, temperature: s.cfg.VirtualPartnerTemp}, nil
	case RoleVirtualPartnerRuntime:
		return roleConfig{model: s.cfg.VirtualPartnerRuntimeModel, temperature: s.cfg.VirtualPartnerRuntimeTemp, think: s.cfg.VirtualPartnerRuntimeThink}, nil
	case RoleTrainingState:
		return roleConfig{model: s.cfg.TrainingStateModel, temperature: s.cfg.TrainingStateTemp, think: s.cfg.TrainingStateThink}, nil
	case RoleTrainingAssessment:
		return roleConfig{model: s.cfg.TrainingAssessmentModel, temperature: s.cfg.TrainingAssessmentTemp, think: s.cfg.TrainingAssessmentThink}, nil
	default:
		return roleConfig{}, fmt.Errorf("unknown ai role %s", role)
	}
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
