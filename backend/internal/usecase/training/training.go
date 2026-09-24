package training

import (
	"context"
	"encoding/json"
	"strings"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/ai"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
)

type repository interface {
	CreateSession(ctx context.Context, params entity.StartTrainingSession, snapshot, state []byte) (*entity.TrainingSession, error)
	GetSession(ctx context.Context, id string) (*entity.TrainingSession, error)
	ListSessions(ctx context.Context, filter entity.TrainingSessionFilter) ([]entity.TrainingSession, error)
	UpdateState(ctx context.Context, sessionID string, state []byte, status entity.TrainingStatus) (*entity.TrainingSession, error)
	Complete(ctx context.Context, sessionID string, score int, analysis []byte) (*entity.TrainingSession, error)
	CompleteLinkedAssignment(ctx context.Context, sessionID string) error
	CreateMessage(ctx context.Context, params entity.CreateTrainingMessage) (*entity.TrainingMessage, error)
	ListMessages(ctx context.Context, sessionID string) ([]entity.TrainingMessage, error)
	GetMessage(ctx context.Context, sessionID, messageID string) (*entity.TrainingMessage, error)
	UpdateMessageClassification(ctx context.Context, sessionID, messageID string, classification []byte) (*entity.TrainingMessage, error)
}

type scenarioRepository interface {
	GetScenarioByID(ctx context.Context, id string) (*entity.Scenario, error)
}

type profileRecalculator interface {
	RecalculateAfterTraining(ctx context.Context, userID, sourceID string, score int) ([]entity.ProfileClusterChange, error)
}

type transactor interface {
	WithTx(ctx context.Context, f func(ctx context.Context) error) error
}

type aiService interface {
	GenerateJSON(ctx context.Context, role ai.Role, input any, output any, validate func() error) error
}

type Service struct {
	repository         repository
	scenarioRepository scenarioRepository
	profile            profileRecalculator
	transactor         transactor
	ai                 aiService
}

func NewService(repository repository, scenarioRepository scenarioRepository, profile profileRecalculator, transactor transactor, ai aiService) *Service {
	return &Service{repository: repository, scenarioRepository: scenarioRepository, profile: profile, transactor: transactor, ai: ai}
}

func (s *Service) ListSessions(ctx context.Context, claims entity.AccessClaims, filter entity.TrainingSessionFilter) ([]entity.TrainingSession, error) {
	filter.UserID = claims.UserID
	return s.repository.ListSessions(ctx, filter)
}

func (s *Service) StartSession(ctx context.Context, claims entity.AccessClaims, params entity.StartTrainingSession) (*entity.TrainingSessionDetails, error) {
	params.UserID = claims.UserID
	scenario, err := s.scenarioRepository.GetScenarioByID(ctx, params.ScenarioID)
	if err != nil {
		return nil, err
	}
	doc := entity.ScenarioDocumentFromScenario(*scenario)
	if scenario.Status != entity.ScenarioStatusPublished {
		return nil, entity.ErrScenarioNotAvailable
	}
	if !contains(doc.SupportedRuntimes, string(params.Mode)) || !contains(doc.SupportedChannels, string(params.Channel)) {
		return nil, entity.ErrTrainingConfigurationNotSupported
	}
	if _, ok := roleByID(doc.Roles, params.ScenarioRoleID); !ok {
		return nil, entity.ErrTrainingConfigurationNotSupported
	}
	state := initialState(doc, params.Mode, params.ScenarioRoleID)
	if params.Mode == entity.TrainingModeDeterministic && (state.BlockID == "" || state.CurrentNodeID == "") {
		return nil, entity.ErrTrainingConfigurationNotSupported
	}
	stateRaw, err := json.Marshal(state)
	if err != nil {
		return nil, err
	}
	snapshot, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	session, err := s.repository.CreateSession(ctx, params, snapshot, stateRaw)
	if err != nil {
		return nil, err
	}
	session.ScenarioTitle = scenario.Title
	if params.Mode == entity.TrainingModeDeterministic {
		if node, ok := currentNode(doc, state); ok && node.IncomingMessage != nil {
			msg, err := s.createPartnerMessage(ctx, session.ID, node.IncomingMessage.Text, node.IncomingMessage.SemanticActionCode)
			if err != nil {
				return nil, err
			}
			return &entity.TrainingSessionDetails{Session: *session, Messages: []entity.TrainingMessage{*msg}}, nil
		}
	}
	return &entity.TrainingSessionDetails{Session: *session, Messages: []entity.TrainingMessage{}}, nil
}

func (s *Service) GetSession(ctx context.Context, claims entity.AccessClaims, sessionID string) (*entity.TrainingSessionDetails, error) {
	session, err := s.requireSessionOwner(ctx, claims, sessionID)
	if err != nil {
		return nil, err
	}
	messages, err := s.repository.ListMessages(ctx, session.ID)
	if err != nil {
		return nil, err
	}
	return &entity.TrainingSessionDetails{Session: *session, Messages: messages}, nil
}

func (s *Service) SubmitAction(ctx context.Context, claims entity.AccessClaims, sessionID, actionID, selectedCode, clientEventID string) (*entity.TrainingActionResult, error) {
	session, err := s.requireSessionOwner(ctx, claims, sessionID)
	if err != nil {
		return nil, err
	}
	if session.Mode != entity.TrainingModeDeterministic {
		return nil, entity.ErrTrainingConfigurationNotSupported
	}
	if session.Status != entity.TrainingStatusInProgress {
		return nil, entity.ErrInvalidTrainingState
	}
	doc, state, err := sessionRuntime(session)
	if err != nil {
		return nil, err
	}
	node, ok := currentNode(doc, state)
	if !ok {
		return nil, entity.ErrInvalidTrainingState
	}
	option, ok := actionByID(node.ReplyOptions, actionID)
	if !ok {
		return nil, entity.ErrActionNotAvailable
	}
	classification := classificationJSON(selectedCode, option.SemanticActionCode)
	msg, err := s.repository.CreateMessage(ctx, entity.CreateTrainingMessage{
		SessionID:          session.ID,
		Author:             "learner",
		Content:            chatContent(option.Text),
		SemanticActionCode: &selectedCode,
		ExpectedActionCode: &option.SemanticActionCode,
		Classification:     classification,
	})
	if err != nil {
		return nil, err
	}
	status := entity.TrainingStatusInProgress
	state.CurrentNodeID = option.NextNodeID
	if state.CurrentNodeID == "" {
		status = entity.TrainingStatusReadyToComplete
	} else if next, ok := currentNode(doc, state); ok && next.IncomingMessage != nil {
		if _, err := s.createPartnerMessage(ctx, session.ID, next.IncomingMessage.Text, next.IncomingMessage.SemanticActionCode); err != nil {
			return nil, err
		}
	} else if !ok {
		status = entity.TrainingStatusReadyToComplete
	}
	stateRaw, err := json.Marshal(state)
	if err != nil {
		return nil, err
	}
	if _, err := s.repository.UpdateState(ctx, session.ID, stateRaw, status); err != nil {
		return nil, err
	}
	return &entity.TrainingActionResult{EventID: msg.ID, Classification: classification, NextState: stateRaw, Status: status}, nil
}

func (s *Service) SendMessage(ctx context.Context, claims entity.AccessClaims, sessionID string, messageID string, content json.RawMessage, semanticActionCode *string, annotations json.RawMessage) (*entity.TrainingMessageExchange, error) {
	session, err := s.requireSessionOwner(ctx, claims, sessionID)
	if err != nil {
		return nil, err
	}
	if session.Mode != entity.TrainingModeVirtualPartner {
		return nil, entity.ErrTrainingConfigurationNotSupported
	}
	if session.Status != entity.TrainingStatusInProgress {
		return nil, entity.ErrInvalidTrainingState
	}
	if !contentMatchesChannel(content, session.Channel) {
		return nil, entity.ErrTrainingMessageChannelMismatch
	}
	doc, state, err := sessionRuntime(session)
	if err != nil {
		return nil, err
	}
	history, err := s.repository.ListMessages(ctx, session.ID)
	if err != nil {
		return nil, err
	}
	learnerText, err := contentText(content, session.Channel)
	if err != nil {
		return nil, err
	}
	classification, err := s.classifyLearnerMessage(ctx, session, state, learnerText)
	if err != nil {
		return nil, err
	}
	classificationRaw := classificationJSONFromAI(semanticActionCode, classification)
	turn, err := s.generateVirtualPartnerTurn(ctx, session, doc, state, history, content, learnerText, classification)
	if err != nil {
		return nil, err
	}
	partnerContent, err := normalizePartnerContent(turn.Content, session.Channel)
	if err != nil {
		return nil, err
	}
	nextState, status, err := applyVirtualPartnerTurn(doc, state, turn)
	if err != nil {
		return nil, err
	}
	var learner *entity.TrainingMessage
	var partner *entity.TrainingMessage
	err = s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		learner, err = s.repository.CreateMessage(txCtx, entity.CreateTrainingMessage{ID: messageID, SessionID: session.ID, Author: "learner", Content: content, SemanticActionCode: semanticActionCode, SemanticAnnotations: annotations, Classification: classificationRaw})
		if err != nil {
			return err
		}
		expected := turn.PartnerSemanticActionCode
		partner, err = s.repository.CreateMessage(txCtx, entity.CreateTrainingMessage{SessionID: session.ID, Author: "virtual_partner", Content: partnerContent, ExpectedActionCode: &expected})
		if err != nil {
			return err
		}
		stateRaw, err := json.Marshal(nextState)
		if err != nil {
			return err
		}
		_, err = s.repository.UpdateState(txCtx, session.ID, stateRaw, status)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &entity.TrainingMessageExchange{LearnerMessage: *learner, PartnerMessage: *partner, Status: status}, nil
}

func (s *Service) ClassifyMessage(ctx context.Context, claims entity.AccessClaims, sessionID, messageID, selectedCode string) (*entity.TrainingMessage, error) {
	session, err := s.requireSessionOwner(ctx, claims, sessionID)
	if err != nil {
		return nil, err
	}
	msg, err := s.repository.GetMessage(ctx, session.ID, messageID)
	if err != nil {
		return nil, err
	}
	if msg.Author != "virtual_partner" || msg.ExpectedActionCode == nil {
		return nil, entity.ErrMessageNotClassifiable
	}
	if len(msg.Classification) > 0 {
		return nil, entity.ErrMessageAlreadyClassified
	}
	return s.repository.UpdateMessageClassification(ctx, session.ID, messageID, classificationJSON(selectedCode, *msg.ExpectedActionCode))
}

func (s *Service) Complete(ctx context.Context, claims entity.AccessClaims, sessionID string) (*entity.TrainingCompletion, error) {
	session, err := s.requireSessionOwner(ctx, claims, sessionID)
	if err != nil {
		return nil, err
	}
	if session.Status == entity.TrainingStatusCompleted {
		return nil, entity.ErrInvalidTrainingState
	}
	var result *entity.TrainingCompletion
	err = s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		messages, err := s.repository.ListMessages(txCtx, session.ID)
		if err != nil {
			return err
		}
		score := scoreMessages(messages)
		outcome := outcomeJSON(score)
		analysis := analysisJSON(session.ID, score, outcome, messages, nil)
		completed, err := s.repository.Complete(txCtx, session.ID, score, analysis)
		if err != nil {
			return err
		}
		changes, err := s.profile.RecalculateAfterTraining(txCtx, session.UserID, session.ID, score)
		if err != nil {
			return err
		}
		analysis = analysisJSON(session.ID, score, outcome, messages, changes)
		completed, err = s.repository.Complete(txCtx, session.ID, score, analysis)
		if err != nil {
			return err
		}
		if err := s.repository.CompleteLinkedAssignment(txCtx, session.ID); err != nil {
			return err
		}
		result = &entity.TrainingCompletion{Session: *completed, Score: score, Outcome: outcome, ProfileChange: changes}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) GetAnalysis(ctx context.Context, claims entity.AccessClaims, sessionID string) (json.RawMessage, error) {
	session, err := s.requireSessionOwner(ctx, claims, sessionID)
	if err != nil {
		return nil, err
	}
	if session.Status != entity.TrainingStatusCompleted || len(session.Analysis) == 0 {
		return nil, entity.ErrInvalidTrainingState
	}
	return session.Analysis, nil
}

func (s *Service) requireSessionOwner(ctx context.Context, claims entity.AccessClaims, sessionID string) (*entity.TrainingSession, error) {
	session, err := s.repository.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session.UserID != claims.UserID {
		return nil, entity.ErrTrainingAccessDenied
	}
	return session, nil
}

func (s *Service) createPartnerMessage(ctx context.Context, sessionID, text, expectedCode string) (*entity.TrainingMessage, error) {
	return s.repository.CreateMessage(ctx, entity.CreateTrainingMessage{SessionID: sessionID, Author: "virtual_partner", Content: chatContent(text), ExpectedActionCode: &expectedCode})
}

type trainingState struct {
	BlockID       string   `json:"block_id"`
	CurrentNodeID string   `json:"current_node_id"`
	TurnNumber    int      `json:"turn_number"`
	ShortSummary  string   `json:"short_summary"`
	RevealedFacts []string `json:"revealed_facts"`
}

func initialState(doc entity.ScenarioDocument, mode entity.TrainingMode, learnerRoleID string) trainingState {
	if len(doc.Blocks) == 0 {
		return trainingState{}
	}
	if mode == entity.TrainingModeVirtualPartner {
		for _, block := range doc.Blocks {
			if block.Type == "regular" {
				return trainingState{BlockID: block.ID}
			}
		}
	}
	if mode != entity.TrainingModeDeterministic {
		return trainingState{}
	}
	for _, block := range doc.Blocks {
		if block.DeterministicDialogue != nil && block.DeterministicDialogue.LearnerRoleID == learnerRoleID && block.DeterministicDialogue.StartNodeID != "" && len(block.DeterministicDialogue.Nodes) > 0 {
			return trainingState{BlockID: block.ID, CurrentNodeID: block.DeterministicDialogue.StartNodeID}
		}
	}
	return trainingState{}
}

func sessionRuntime(session *entity.TrainingSession) (entity.ScenarioDocument, trainingState, error) {
	var doc entity.ScenarioDocument
	if err := json.Unmarshal(session.Snapshot, &doc); err != nil {
		return doc, trainingState{}, entity.ErrInvalidScenarioJSON
	}
	var state trainingState
	if len(session.State) > 0 {
		if err := json.Unmarshal(session.State, &state); err != nil {
			return doc, state, entity.ErrInvalidRequest
		}
	}
	return doc, state, nil
}

func currentNode(doc entity.ScenarioDocument, state trainingState) (entity.DeterministicDialogueNode, bool) {
	for _, block := range doc.Blocks {
		if block.ID != state.BlockID {
			continue
		}
		if block.DeterministicDialogue == nil {
			return entity.DeterministicDialogueNode{}, false
		}
		for _, node := range block.DeterministicDialogue.Nodes {
			if node.ID == state.CurrentNodeID {
				return node, true
			}
		}
	}
	return entity.DeterministicDialogueNode{}, false
}

func actionByID(actions []entity.DeterministicReplyOption, id string) (entity.DeterministicReplyOption, bool) {
	for _, action := range actions {
		if action.ID == id {
			return action, true
		}
	}
	return entity.DeterministicReplyOption{}, false
}

func roleByID(roles []entity.ScenarioRole, id string) (entity.ScenarioRole, bool) {
	for _, role := range roles {
		if role.ID == id {
			return role, true
		}
	}
	return entity.ScenarioRole{}, false
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func chatContent(text string) json.RawMessage {
	raw, _ := json.Marshal(map[string]string{"type": "chat", "text": text})
	return raw
}

func contentMatchesChannel(raw json.RawMessage, channel entity.TrainingChannel) bool {
	var payload struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return false
	}
	return payload.Type == string(channel)
}

func (s *Service) classifyLearnerMessage(ctx context.Context, session *entity.TrainingSession, state trainingState, text string) (entity.SemanticClassification, error) {
	var output entity.SemanticClassification
	input := map[string]any{
		"task":             "classify_message",
		"semantic_actions": entity.DefaultSemanticActions(),
		"text":             text,
		"context": map[string]any{
			"channel":               session.Channel,
			"current_block_summary": state.ShortSummary,
		},
	}
	err := s.ai.GenerateJSON(ctx, ai.RoleActionClassifier, input, &output, func() error {
		if !semanticActionAllowed(output.SemanticActionCode) {
			return entity.ErrAIInvalidStructuredOutput
		}
		if output.Confidence < 0 || output.Confidence > 1 {
			return entity.ErrAIInvalidStructuredOutput
		}
		return nil
	})
	return output, err
}

func (s *Service) generateVirtualPartnerTurn(ctx context.Context, session *entity.TrainingSession, doc entity.ScenarioDocument, state trainingState, history []entity.TrainingMessage, learnerContent json.RawMessage, learnerText string, classification entity.SemanticClassification) (entity.VirtualPartnerTurn, error) {
	var output entity.VirtualPartnerTurn
	current, _ := blockByID(doc.Blocks, state.BlockID)
	learnerRole, _ := roleByID(doc.Roles, session.ScenarioRoleID)
	partnerRole := entity.ScenarioRole{}
	for _, role := range doc.Roles {
		if role.ID != session.ScenarioRoleID {
			partnerRole = role
			break
		}
	}
	input := map[string]any{
		"channel":                  session.Channel,
		"difficulty":               doc.Difficulty,
		"scenario":                 map[string]any{"expanded_context": doc.ExpandedContext},
		"current_block":            current,
		"partner_role":             partnerRole,
		"learner_role_public":      publicRole(learnerRole),
		"detected_learner_actions": []string{classification.SemanticActionCode},
		"runtime_state":            state,
		"history":                  history,
		"learner_message":          map[string]any{"content": learnerContent, "text": learnerText},
		"semantic_actions":         entity.DefaultSemanticActions(),
		"allowed_transitions":      allowedTransitions(current),
	}
	err := s.ai.GenerateJSON(ctx, ai.RoleVirtualPartner, input, &output, func() error {
		if !semanticActionAllowed(output.PartnerSemanticActionCode) {
			return entity.ErrAIInvalidStructuredOutput
		}
		if _, err := normalizePartnerContent(output.Content, session.Channel); err != nil {
			return err
		}
		if output.Transition.TransitionID != nil && !transitionAllowed(current, *output.Transition.TransitionID) {
			return entity.ErrAIInvalidStructuredOutput
		}
		return nil
	})
	return output, err
}

func contentText(raw json.RawMessage, channel entity.TrainingChannel) (string, error) {
	switch channel {
	case entity.TrainingChannelChat:
		var payload struct {
			Text string `json:"text"`
		}
		if json.Unmarshal(raw, &payload) != nil || strings.TrimSpace(payload.Text) == "" {
			return "", entity.ErrInvalidRequest
		}
		return strings.TrimSpace(payload.Text), nil
	case entity.TrainingChannelEmail:
		var payload struct {
			Subject string `json:"subject"`
			Body    string `json:"body"`
		}
		if json.Unmarshal(raw, &payload) != nil || strings.TrimSpace(payload.Body) == "" {
			return "", entity.ErrInvalidRequest
		}
		return strings.TrimSpace(payload.Subject + "\n" + payload.Body), nil
	case entity.TrainingChannelVoiceMessage:
		var payload struct {
			Transcript string `json:"transcript"`
			SpokenText string `json:"spoken_text"`
		}
		if json.Unmarshal(raw, &payload) != nil {
			return "", entity.ErrInvalidRequest
		}
		text := strings.TrimSpace(payload.Transcript)
		if text == "" {
			text = strings.TrimSpace(payload.SpokenText)
		}
		if text == "" {
			return "", entity.ErrInvalidRequest
		}
		return text, nil
	default:
		return "", entity.ErrInvalidRequest
	}
}

func normalizePartnerContent(content map[string]any, channel entity.TrainingChannel) (json.RawMessage, error) {
	if content == nil {
		return nil, entity.ErrAIInvalidStructuredOutput
	}
	content["type"] = string(channel)
	switch channel {
	case entity.TrainingChannelChat:
		if strings.TrimSpace(stringFromAny(content["text"])) == "" {
			return nil, entity.ErrAIInvalidStructuredOutput
		}
	case entity.TrainingChannelEmail:
		if strings.TrimSpace(stringFromAny(content["subject"])) == "" || strings.TrimSpace(stringFromAny(content["body"])) == "" {
			return nil, entity.ErrAIInvalidStructuredOutput
		}
	case entity.TrainingChannelVoiceMessage:
		if strings.TrimSpace(stringFromAny(content["spoken_text"])) == "" {
			return nil, entity.ErrAIInvalidStructuredOutput
		}
	default:
		return nil, entity.ErrAIInvalidStructuredOutput
	}
	raw, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func applyVirtualPartnerTurn(doc entity.ScenarioDocument, state trainingState, turn entity.VirtualPartnerTurn) (trainingState, entity.TrainingStatus, error) {
	next := state
	next.TurnNumber++
	if turn.StateUpdates.ShortSummary != "" {
		next.ShortSummary = turn.StateUpdates.ShortSummary
	}
	next.RevealedFacts = append(next.RevealedFacts, turn.StateUpdates.RevealedFacts...)
	status := entity.TrainingStatusInProgress
	if turn.Transition.TransitionID != nil {
		current, ok := blockByID(doc.Blocks, state.BlockID)
		if !ok {
			return next, status, entity.ErrInvalidTrainingState
		}
		var targetID string
		for _, transition := range current.Transitions {
			if transition.ID == *turn.Transition.TransitionID && !transition.Forbidden {
				targetID = transition.TargetBlockID
				break
			}
		}
		if targetID == "" {
			return next, status, entity.ErrAIInvalidStructuredOutput
		}
		next.BlockID = targetID
		if target, ok := blockByID(doc.Blocks, targetID); ok && target.Type == "end" {
			status = entity.TrainingStatusReadyToComplete
		}
	}
	if turn.Completion.Completed {
		status = entity.TrainingStatusReadyToComplete
	}
	return next, status, nil
}

func classificationJSONFromAI(selected *string, detected entity.SemanticClassification) json.RawMessage {
	expected := detected.SemanticActionCode
	selectedValue := ""
	if selected != nil {
		selectedValue = *selected
	}
	raw, _ := json.Marshal(map[string]any{"selected_semantic_action_code": selectedValue, "expected_semantic_action_code": expected, "correct": selectedValue == expected, "confidence": detected.Confidence, "reason": detected.Reason})
	return raw
}

func blockByID(blocks []entity.ScenarioBlock, id string) (entity.ScenarioBlock, bool) {
	for _, block := range blocks {
		if block.ID == id {
			return block, true
		}
	}
	return entity.ScenarioBlock{}, false
}

func allowedTransitions(block entity.ScenarioBlock) []entity.ScenarioTransition {
	items := []entity.ScenarioTransition{}
	for _, transition := range block.Transitions {
		if !transition.Forbidden {
			items = append(items, transition)
		}
	}
	return items
}

func transitionAllowed(block entity.ScenarioBlock, id string) bool {
	for _, transition := range block.Transitions {
		if transition.ID == id && !transition.Forbidden {
			return true
		}
	}
	return false
}

func semanticActionAllowed(code string) bool {
	for _, action := range entity.DefaultSemanticActions() {
		if action.Code == code {
			return true
		}
	}
	return false
}

func publicRole(role entity.ScenarioRole) map[string]any {
	return map[string]any{"id": role.ID, "name": role.Name, "description": role.Description, "goal": role.Goal, "interests": role.Interests, "constraints": role.Constraints}
}

func stringFromAny(value any) string {
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}

func classificationJSON(selected, expected string) json.RawMessage {
	raw, _ := json.Marshal(map[string]any{"selected_semantic_action_code": selected, "expected_semantic_action_code": expected, "correct": selected == expected})
	return raw
}

func outcomeJSON(score int) json.RawMessage {
	raw, _ := json.Marshal(map[string]any{"summary": "Training completed", "goals_met": []string{}, "goals_missed": []string{}, "score": score})
	return raw
}

func analysisJSON(sessionID string, score int, outcome json.RawMessage, messages []entity.TrainingMessage, changes []entity.ProfileClusterChange) json.RawMessage {
	process := make([]map[string]any, 0, len(messages))
	for i, message := range messages {
		if message.SemanticActionCode == nil {
			continue
		}
		process = append(process, map[string]any{"order": i + 1, "summary": "Learner action: " + *message.SemanticActionCode, "semantic_action_codes": []string{*message.SemanticActionCode}})
	}
	raw, _ := json.Marshal(map[string]any{
		"session_id":     sessionID,
		"score":          score,
		"result":         json.RawMessage(outcome),
		"process":        process,
		"strengths":      []map[string]string{{"title": "Consistency", "description": "Learner kept the dialogue moving."}},
		"improvements":   []map[string]string{{"title": "Precision", "description": "Make trade-offs and criteria more explicit."}},
		"competencies":   []map[string]any{},
		"profile_change": map[string]any{"changes": changes},
	})
	return raw
}

func scoreMessages(messages []entity.TrainingMessage) int {
	total := 0
	correct := 0
	for _, message := range messages {
		if len(message.Classification) == 0 {
			continue
		}
		var classification struct {
			Correct bool `json:"correct"`
		}
		if json.Unmarshal(message.Classification, &classification) == nil {
			total++
			if classification.Correct {
				correct++
			}
		}
	}
	if total == 0 {
		return 70
	}
	score := 50 + correct*50/total
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

func TrimSemanticActionCode(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
