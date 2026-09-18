package training

import (
	"context"
	"encoding/json"
	"strings"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
)

type trainingRepository interface {
	CreateSession(context.Context, entity.TrainingSession) (*entity.TrainingSession, error)
	GetSession(context.Context, string) (*entity.TrainingSession, error)
	UpdateSessionState(context.Context, string, *string, entity.TrainingStatus) (*entity.TrainingSession, error)
	CompleteSession(context.Context, string, []byte, []byte) (*entity.TrainingSession, error)
	CreateEvent(context.Context, entity.TrainingEvent) (*entity.TrainingEvent, error)
	CountEvents(context.Context, string) (int32, error)
	ListEvents(context.Context, string) ([]entity.TrainingEvent, error)
	CreateMessage(context.Context, entity.TrainingMessage) (*entity.TrainingMessage, error)
	ListMessages(context.Context, string) ([]entity.TrainingMessage, error)
}
type scenarioRepository interface {
	GetScenarioByID(context.Context, string) (*entity.Scenario, error)
	GetScenarioVersion(context.Context, string, string) (*entity.ScenarioVersion, error)
}
type userRepository interface {
	GetUserByID(context.Context, string) (*entity.User, error)
}
type transactor interface {
	WithTx(context.Context, func(context.Context) error) error
}

type Service struct {
	training  trainingRepository
	scenarios scenarioRepository
	users     userRepository
	tx        transactor
}

func NewService(training trainingRepository, scenarios scenarioRepository, users userRepository, tx transactor) *Service {
	return &Service{training: training, scenarios: scenarios, users: users, tx: tx}
}

func (s *Service) Start(ctx context.Context, claims entity.AccessClaims, input entity.StartTrainingSession) (*entity.TrainingSession, *entity.TrainingNode, []entity.TrainingMessage, error) {
	if !input.Mode.Valid() || input.Channel != "chat" {
		return nil, nil, nil, entity.ErrInvalidRequest
	}
	scenario, err := s.scenarios.GetScenarioByID(ctx, input.ScenarioID)
	if err != nil {
		return nil, nil, nil, err
	}
	user, err := s.users.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, nil, nil, err
	}
	if user.OrganizationID == nil || scenario.OrganizationID == nil || *user.OrganizationID != *scenario.OrganizationID {
		return nil, nil, nil, entity.ErrScenarioNotFound
	}
	if !supports(scenario.ModeSupport, string(input.Mode)) {
		return nil, nil, nil, entity.ErrTrainingModeNotSupported
	}
	if input.Preview && claims.ActiveRole != entity.RoleMethodist {
		return nil, nil, nil, entity.ErrTrainingAccessDenied
	}
	var snapshot []byte
	var versionID *string
	if !input.Preview {
		if scenario.Status != entity.ScenarioStatusPublished {
			return nil, nil, nil, entity.ErrScenarioNotFound
		}
		version, err := s.scenarios.GetScenarioVersion(ctx, scenario.ID, scenario.Version)
		if err != nil {
			return nil, nil, nil, err
		}
		snapshot, versionID = version.Snapshot, &version.ID
	} else {
		snapshot, err = snapshotScenario(scenario)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	nodes, err := nodesFromSnapshot(snapshot)
	if err != nil {
		return nil, nil, nil, entity.ErrInvalidScenarioJSON
	}
	var current *string
	if input.Mode == entity.TrainingModeDeterministic {
		if len(nodes) == 0 {
			return nil, nil, nil, entity.ErrTrainingModeNotSupported
		}
		current = &nodes[0].ID
	}
	id, err := entity.NewUUID()
	if err != nil {
		return nil, nil, nil, err
	}
	session, err := s.training.CreateSession(ctx, entity.TrainingSession{ID: id, UserID: claims.UserID, ScenarioID: scenario.ID, ScenarioVersionID: versionID, ScenarioVersion: scenario.Version, ScenarioSnapshot: snapshot, Mode: input.Mode, Channel: input.Channel, IsPreview: input.Preview, Status: entity.TrainingStatusInProgress, CurrentNodeID: current})
	if err != nil {
		return nil, nil, nil, err
	}
	return session, nodeByID(nodes, current), []entity.TrainingMessage{}, nil
}

func (s *Service) Get(ctx context.Context, claims entity.AccessClaims, sessionID string) (*entity.TrainingSession, *entity.TrainingNode, []entity.TrainingMessage, error) {
	session, err := s.training.GetSession(ctx, sessionID)
	if err != nil {
		return nil, nil, nil, err
	}
	if err := trainingAccess(session, claims); err != nil {
		return nil, nil, nil, err
	}
	nodes, err := nodesFromSnapshot(session.ScenarioSnapshot)
	if err != nil {
		return nil, nil, nil, entity.ErrInvalidScenarioJSON
	}
	messages, err := s.training.ListMessages(ctx, session.ID)
	if err != nil {
		return nil, nil, nil, err
	}
	return session, nodeByID(nodes, session.CurrentNodeID), messages, nil
}

func (s *Service) SelectAction(ctx context.Context, claims entity.AccessClaims, input entity.SelectTrainingAction) (*entity.TrainingActionResult, error) {
	var result *entity.TrainingActionResult
	err := s.tx.WithTx(ctx, func(tx context.Context) error {
		session, err := s.training.GetSession(tx, input.SessionID)
		if err != nil {
			return err
		}
		if err := trainingAccess(session, claims); err != nil {
			return err
		}
		if session.Mode != entity.TrainingModeDeterministic || session.Status != entity.TrainingStatusInProgress {
			return entity.ErrInvalidTrainingState
		}
		nodes, err := nodesFromSnapshot(session.ScenarioSnapshot)
		if err != nil {
			return entity.ErrInvalidScenarioJSON
		}
		current := nodeByID(nodes, session.CurrentNodeID)
		if current == nil {
			return entity.ErrInvalidTrainingState
		}
		var action *entity.AvailableAction
		for i := range current.AvailableActions {
			if current.AvailableActions[i].ID == input.ActionID {
				action = &current.AvailableActions[i]
				break
			}
		}
		if action == nil {
			return entity.ErrTrainingActionNotAvailable
		}
		count, err := s.training.CountEvents(tx, session.ID)
		if err != nil {
			return err
		}
		eventID, err := entity.NewUUID()
		if err != nil {
			return err
		}
		nextID := action.NextNodeID
		event, err := s.training.CreateEvent(tx, entity.TrainingEvent{ID: eventID, SessionID: session.ID, ClientEventID: input.ClientEventID, SequenceNumber: count + 1, NodeID: current.ID, ActionID: action.ID, SemanticAction: action.SemanticAction, NextNodeID: optional(nextID)})
		if err != nil {
			return err
		}
		next := nodeByID(nodes, optional(nextID))
		status := entity.TrainingStatusInProgress
		if next == nil || next.Terminal {
			status = entity.TrainingStatusReadyToComplete
		}
		if _, err := s.training.UpdateSessionState(tx, session.ID, optional(nextID), status); err != nil {
			return err
		}
		result = &entity.TrainingActionResult{Event: *event, NextNode: next, Status: status}
		return nil
	})
	return result, err
}

func (s *Service) SendMessage(ctx context.Context, claims entity.AccessClaims, input entity.SendTrainingMessage) (*entity.TrainingMessageResult, error) {
	var result *entity.TrainingMessageResult
	err := s.tx.WithTx(ctx, func(tx context.Context) error {
		session, err := s.training.GetSession(tx, input.SessionID)
		if err != nil {
			return err
		}
		if err := trainingAccess(session, claims); err != nil {
			return err
		}
		if session.Mode != entity.TrainingModeVirtualPartner || session.Status != entity.TrainingStatusInProgress {
			return entity.ErrInvalidTrainingState
		}
		learner, err := s.training.CreateMessage(tx, entity.TrainingMessage{ID: input.MessageID, SessionID: session.ID, Author: "learner", Text: input.Text})
		if err != nil {
			return err
		}
		partnerID, err := entity.NewUUID()
		if err != nil {
			return err
		}
		partner, err := s.training.CreateMessage(tx, entity.TrainingMessage{ID: partnerID, SessionID: session.ID, Author: "virtual_partner", Text: partnerReply(input.Text)})
		if err != nil {
			return err
		}
		result = &entity.TrainingMessageResult{LearnerMessage: *learner, PartnerMessage: *partner, SemanticActions: classify(input.Text)}
		return nil
	})
	return result, err
}

func (s *Service) Complete(ctx context.Context, claims entity.AccessClaims, sessionID string) (*entity.TrainingCompletion, error) {
	session, err := s.training.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if err := trainingAccess(session, claims); err != nil {
		return nil, err
	}
	if session.Status != entity.TrainingStatusInProgress && session.Status != entity.TrainingStatusReadyToComplete {
		return nil, entity.ErrInvalidTrainingState
	}
	events, err := s.training.ListEvents(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	messages, err := s.training.ListMessages(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	outcome, _ := json.Marshal(map[string]any{"goals_met": []string{}, "goals_missed": []string{}})
	process := make([]map[string]any, 0, len(events)+len(messages))
	for _, event := range events {
		process = append(process, map[string]any{"step": event.SequenceNumber, "summary": event.ActionID, "semantic_actions": []string{event.SemanticAction}})
	}
	if len(messages) > 0 {
		process = append(process, map[string]any{"step": len(process) + 1, "summary": "chat messages exchanged", "semantic_actions": []string{}})
	}
	analysis, _ := json.Marshal(map[string]any{"result": json.RawMessage(outcome), "process": process, "competency": []any{}})
	if _, err := s.training.CompleteSession(ctx, sessionID, outcome, analysis); err != nil {
		return nil, err
	}
	return &entity.TrainingCompletion{Outcome: outcome, Analysis: analysis}, nil
}
func (s *Service) Analysis(ctx context.Context, claims entity.AccessClaims, sessionID string) (json.RawMessage, error) {
	session, err := s.training.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if err := trainingAccess(session, claims); err != nil {
		return nil, err
	}
	if session.Status != entity.TrainingStatusCompleted {
		return nil, entity.ErrInvalidTrainingState
	}
	return session.Analysis, nil
}

func trainingAccess(session *entity.TrainingSession, claims entity.AccessClaims) error {
	if session.UserID != claims.UserID && !(session.IsPreview && claims.ActiveRole == entity.RoleMethodist) {
		return entity.ErrTrainingAccessDenied
	}
	return nil
}
func supports(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}
func optional(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
func nodeByID(nodes []entity.TrainingNode, id *string) *entity.TrainingNode {
	if id == nil {
		return nil
	}
	for i := range nodes {
		if nodes[i].ID == *id {
			return &nodes[i]
		}
	}
	return nil
}
func nodesFromSnapshot(snapshot []byte) ([]entity.TrainingNode, error) {
	var value struct {
		Nodes json.RawMessage `json:"nodes"`
	}
	if err := json.Unmarshal(snapshot, &value); err != nil {
		return nil, err
	}
	var nodes []entity.TrainingNode
	if err := json.Unmarshal(value.Nodes, &nodes); err == nil && len(nodes) > 0 {
		return nodes, nil
	}
	var alternate []struct {
		ID       string                   `json:"id"`
		Speaker  string                   `json:"speaker"`
		Text     string                   `json:"text"`
		Actions  []entity.AvailableAction `json:"actions"`
		Terminal bool                     `json:"terminal"`
	}
	if err := json.Unmarshal(value.Nodes, &alternate); err != nil {
		return nil, err
	}
	for _, node := range alternate {
		nodes = append(nodes, entity.TrainingNode{ID: node.ID, Speaker: node.Speaker, Text: node.Text, AvailableActions: node.Actions, Terminal: node.Terminal})
	}
	return nodes, nil
}
func snapshotScenario(s *entity.Scenario) ([]byte, error) {
	return json.Marshal(map[string]any{"scenario_id": s.ID, "version": s.Version, "title": s.Title, "nodes": json.RawMessage(s.Nodes)})
}
func classify(text string) []string {
	lower := strings.ToLower(text)
	actions := make([]string, 0, 2)
	if strings.Contains(lower, "?") {
		actions = append(actions, "ASK_QUESTION")
	}
	if strings.Contains(lower, "предлага") {
		actions = append(actions, "MAKE_PROPOSAL")
	}
	if len(actions) == 0 {
		actions = append(actions, "SUMMARIZE")
	}
	return actions
}
func partnerReply(text string) string {
	return "Понял вашу позицию. Уточните, пожалуйста, какие условия для вас принципиальны?"
}
