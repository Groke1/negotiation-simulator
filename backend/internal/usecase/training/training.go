package training

import (
	"context"
	"encoding/json"
	"math"
	"strings"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/ai"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
)

type repository interface {
	CreateSession(ctx context.Context, params entity.StartTrainingSession, snapshot, state []byte) (*entity.TrainingSession, error)
	GetSession(ctx context.Context, id string) (*entity.TrainingSession, error)
	ListSessions(ctx context.Context, filter entity.TrainingSessionFilter) ([]entity.TrainingSession, error)
	UpdateState(ctx context.Context, sessionID string, state []byte, status entity.TrainingStatus) (*entity.TrainingSession, error)
	Complete(ctx context.Context, sessionID string, score float64, analysis []byte) (*entity.TrainingSession, error)
	SetAnalysis(ctx context.Context, sessionID string, analysis []byte) (*entity.TrainingSession, error)
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
	RecalculateAfterTraining(ctx context.Context, userID, sourceID string, results []entity.TrainingClusterResult) ([]entity.ProfileClusterChange, error)
	ListSemanticActions(ctx context.Context) ([]entity.SemanticActionDefinition, error)
}

type transactor interface {
	WithTx(ctx context.Context, f func(ctx context.Context) error) error
}

type aiService interface {
	GenerateJSON(ctx context.Context, role ai.Role, input any, output any, validate func() error) error
}

const (
	virtualPartnerPatienceMin       = 0
	virtualPartnerPatienceThreshold = 100
	virtualPartnerPatienceDeltaMin  = -5
	virtualPartnerPatienceDeltaMax  = 40
	trainingAnalysisFailedStatus    = "failed"
	trainingAnalysisFailedErrorCode = "AI_ANALYSIS_FAILED"
)

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
	block, node, ok := currentBlockAndNode(doc, state)
	if !ok {
		return nil, entity.ErrInvalidTrainingState
	}
	option, ok := actionByID(node.ReplyOptions, actionID)
	if !ok {
		return nil, entity.ErrActionNotAvailable
	}
	classification := deterministicEventJSON(block, node, option, selectedCode, clientEventID)
	status := entity.TrainingStatusInProgress
	terminal := false
	state.CurrentNodeID = option.NextNodeID
	if state.CurrentNodeID == "" && option.TransitionID != "" {
		transition, ok := transitionByID(block.Transitions, option.TransitionID)
		if !ok {
			return nil, entity.ErrInvalidTrainingState
		}
		state.BlockID = transition.TargetBlockID
		target, ok := blockByID(doc.Blocks, transition.TargetBlockID)
		if !ok {
			return nil, entity.ErrInvalidTrainingState
		}
		if target.DeterministicDialogue != nil && target.DeterministicDialogue.LearnerRoleID == session.ScenarioRoleID && target.DeterministicDialogue.StartNodeID != "" {
			state.CurrentNodeID = target.DeterministicDialogue.StartNodeID
		} else if target.Type == "end" {
			terminal = true
		} else {
			return nil, entity.ErrInvalidTrainingState
		}
	} else if state.CurrentNodeID == "" {
		terminal = true
	} else if next, ok := currentNode(doc, state); ok && next.IncomingMessage != nil {
	} else if !ok {
		terminal = true
	}
	stateRaw, err := json.Marshal(state)
	if err != nil {
		return nil, err
	}
	var msg *entity.TrainingMessage
	err = s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		msg, err = s.repository.CreateMessage(txCtx, entity.CreateTrainingMessage{
			ID:                 clientEventID,
			SessionID:          session.ID,
			Author:             "learner",
			Content:            chatContent(option.Text),
			SemanticActionCode: &selectedCode,
			ExpectedActionCode: &option.SemanticActionCode,
			Classification:     classification,
		})
		if err != nil {
			return err
		}
		if !terminal {
			if next, ok := currentNode(doc, state); ok && next.IncomingMessage != nil {
				if _, err := s.createPartnerMessage(txCtx, session.ID, next.IncomingMessage.Text, next.IncomingMessage.SemanticActionCode); err != nil {
					return err
				}
			}
		}
		updated, err := s.repository.UpdateState(txCtx, session.ID, stateRaw, status)
		if err != nil {
			return err
		}
		if !terminal {
			return nil
		}
		messages, err := s.repository.ListMessages(txCtx, session.ID)
		if err != nil {
			return err
		}
		if _, err := s.completeSessionInTx(txCtx, updated, doc, state, messages, nil); err != nil {
			return err
		}
		status = entity.TrainingStatusCompleted
		return nil
	})
	if err != nil {
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
	runtime, err := s.generateVirtualPartnerRuntime(ctx, session, doc, state, history, content, learnerText)
	if err != nil {
		return nil, err
	}
	classification := entity.SemanticClassification{SemanticActionCode: runtime.LearnerSemanticActionCode, Confidence: runtime.LearnerConfidence}
	classificationRaw := classificationJSONFromAI(semanticActionCode, classification)
	partnerContent, err := normalizePartnerContent(runtime.Content, session.Channel)
	if err != nil {
		return nil, err
	}
	evaluation := entity.TrainingStateEvaluation{Decision: runtime.Decision, TransitionID: runtime.TransitionID}
	nextState, status, err := applyTrainingStateEvaluation(doc, state, evaluation)
	if err != nil {
		return nil, err
	}
	nextState.Patience = clampVirtualPartnerPatience(nextState.Patience + runtime.PatienceDelta)
	if nextState.Patience >= virtualPartnerPatienceThreshold {
		status = entity.TrainingStatusCompleted
	}
	var learner *entity.TrainingMessage
	var partner *entity.TrainingMessage
	err = s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		learner, err = s.repository.CreateMessage(txCtx, entity.CreateTrainingMessage{ID: messageID, SessionID: session.ID, Author: "learner", Content: content, SemanticActionCode: semanticActionCode, SemanticAnnotations: annotations, Classification: classificationRaw})
		if err != nil {
			return err
		}
		expected := runtime.PartnerSemanticActionCode
		partner, err = s.repository.CreateMessage(txCtx, entity.CreateTrainingMessage{SessionID: session.ID, Author: "virtual_partner", Content: partnerContent, ExpectedActionCode: &expected})
		if err != nil {
			return err
		}
		stateRaw, err := json.Marshal(nextState)
		if err != nil {
			return err
		}
		persistedStatus := entity.TrainingStatusInProgress
		if status == entity.TrainingStatusCompleted {
			persistedStatus = entity.TrainingStatusCompleted
		}
		_, err = s.repository.UpdateState(txCtx, session.ID, stateRaw, persistedStatus)
		if err != nil {
			return err
		}
		if status != entity.TrainingStatusCompleted {
			return nil
		}
		return s.repository.CompleteLinkedAssignment(txCtx, session.ID)
	})
	if err != nil {
		return nil, err
	}
	if status == entity.TrainingStatusCompleted {
		s.completeTrainingAnalysisAsync(session.ID)
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
	if msg.Author == "learner" || msg.ExpectedActionCode == nil {
		return nil, entity.ErrMessageNotClassifiable
	}
	if len(msg.Classification) > 0 {
		return nil, entity.ErrMessageAlreadyClassified
	}
	return s.repository.UpdateMessageClassification(ctx, session.ID, messageID, classificationJSON(selectedCode, *msg.ExpectedActionCode))
}

func (s *Service) GetAnalysis(ctx context.Context, claims entity.AccessClaims, sessionID string) (json.RawMessage, error) {
	session, err := s.requireSessionOwner(ctx, claims, sessionID)
	if err != nil {
		return nil, err
	}
	if session.Status != entity.TrainingStatusCompleted {
		return nil, entity.ErrInvalidTrainingState
	}
	if isFailedTrainingAnalysis(session.Analysis) {
		analysis, err := s.calculateTrainingAnalysis(ctx, session)
		if err != nil {
			_ = s.markTrainingAnalysisFailed(ctx, session.ID)
			return nil, entity.ErrTrainingAnalysisNotReady
		}
		return analysis, nil
	}
	if len(session.Analysis) == 0 {
		return nil, entity.ErrTrainingAnalysisNotReady
	}
	return session.Analysis, nil
}

func (s *Service) completeSessionInTx(ctx context.Context, session *entity.TrainingSession, doc entity.ScenarioDocument, state trainingState, messages []entity.TrainingMessage, assessment *entity.TrainingAssessment) (*entity.TrainingCompletion, error) {
	clusterResults := deterministicClusterResults(messages)
	extras := (*trainingAnalysisExtras)(nil)
	if session.Mode == entity.TrainingModeVirtualPartner {
		if assessment != nil {
			clusterResults = virtualPartnerClusterResults(*assessment, doc.TargetClusterIDs)
			extras = analysisExtrasFromAssessment(*assessment)
		}
		clusterResults = ensureTerminatedVirtualPartnerResults(doc, state, clusterResults)
	}
	score := overallTrainingScore(clusterResults)
	outcome := outcomeJSON(doc, state)
	analysis := analysisJSON(session.ID, score, outcome, doc, messages, clusterResults, nil, extras)
	completed, err := s.repository.Complete(ctx, session.ID, score, analysis)
	if err != nil {
		return nil, err
	}
	changes, err := s.profile.RecalculateAfterTraining(ctx, session.UserID, session.ID, clusterResults)
	if err != nil {
		return nil, err
	}
	analysis = analysisJSON(session.ID, score, outcome, doc, messages, clusterResults, changes, extras)
	completed, err = s.repository.Complete(ctx, session.ID, score, analysis)
	if err != nil {
		return nil, err
	}
	if err := s.repository.CompleteLinkedAssignment(ctx, session.ID); err != nil {
		return nil, err
	}
	return &entity.TrainingCompletion{Session: *completed, Score: floorScore(score), Outcome: outcome, ProfileChange: changes}, nil
}

func (s *Service) completeTrainingAnalysisAsync(sessionID string) {
	go func() {
		ctx := context.Background()
		session, err := s.repository.GetSession(ctx, sessionID)
		if err != nil {
			return
		}
		if len(session.Analysis) > 0 && !isFailedTrainingAnalysis(session.Analysis) {
			return
		}
		if _, err := s.calculateTrainingAnalysis(ctx, session); err != nil {
			_ = s.markTrainingAnalysisFailed(ctx, session.ID)
		}
	}()
}

func (s *Service) calculateTrainingAnalysis(ctx context.Context, session *entity.TrainingSession) (json.RawMessage, error) {
	doc, state, err := sessionRuntime(session)
	if err != nil {
		return nil, err
	}
	messages, err := s.repository.ListMessages(ctx, session.ID)
	if err != nil {
		return nil, err
	}
	var assessment *entity.TrainingAssessment
	if session.Mode == entity.TrainingModeVirtualPartner {
		assessment, err = s.assessVirtualPartnerTraining(ctx, session, doc, state, messages)
		if err != nil {
			return nil, err
		}
	}
	var analysis json.RawMessage
	err = s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		fresh, err := s.repository.GetSession(txCtx, session.ID)
		if err != nil {
			return err
		}
		if len(fresh.Analysis) > 0 && !isFailedTrainingAnalysis(fresh.Analysis) {
			analysis = fresh.Analysis
			return nil
		}
		returned, err := s.completeSessionInTx(txCtx, fresh, doc, state, messages, assessment)
		if err != nil {
			return err
		}
		if returned == nil {
			return entity.ErrTrainingAnalysisNotReady
		}
		analysis = returned.Session.Analysis
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(analysis) == 0 || isFailedTrainingAnalysis(analysis) {
		return nil, entity.ErrTrainingAnalysisNotReady
	}
	return analysis, nil
}

func (s *Service) markTrainingAnalysisFailed(ctx context.Context, sessionID string) error {
	_, err := s.repository.SetAnalysis(ctx, sessionID, failedTrainingAnalysisJSON())
	return err
}

func failedTrainingAnalysisJSON() json.RawMessage {
	raw, _ := json.Marshal(map[string]any{
		"status":     trainingAnalysisFailedStatus,
		"retryable":  true,
		"error_code": trainingAnalysisFailedErrorCode,
	})
	return raw
}

func isFailedTrainingAnalysis(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	var payload struct {
		Status    string `json:"status"`
		ErrorCode string `json:"error_code"`
	}
	if json.Unmarshal(raw, &payload) != nil {
		return false
	}
	return payload.Status == trainingAnalysisFailedStatus && payload.ErrorCode == trainingAnalysisFailedErrorCode
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
	Patience      int      `json:"patience"`
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
	_, node, ok := currentBlockAndNode(doc, state)
	return node, ok
}

func currentBlockAndNode(doc entity.ScenarioDocument, state trainingState) (entity.ScenarioBlock, entity.DeterministicDialogueNode, bool) {
	for _, block := range doc.Blocks {
		if block.ID != state.BlockID {
			continue
		}
		if block.DeterministicDialogue == nil {
			return entity.ScenarioBlock{}, entity.DeterministicDialogueNode{}, false
		}
		for _, node := range block.DeterministicDialogue.Nodes {
			if node.ID == state.CurrentNodeID {
				return block, node, true
			}
		}
	}
	return entity.ScenarioBlock{}, entity.DeterministicDialogueNode{}, false
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
	actions, err := s.profile.ListSemanticActions(ctx)
	if err != nil {
		return output, err
	}
	input := map[string]any{
		"task":             "classify_message",
		"semantic_actions": actions,
		"text":             text,
		"context": map[string]any{
			"channel":               session.Channel,
			"current_block_summary": state.ShortSummary,
		},
	}
	err = s.ai.GenerateJSON(ctx, ai.RoleActionClassifier, input, &output, func() error {
		if !semanticActionAllowed(actions, output.SemanticActionCode) {
			return entity.ErrAIInvalidStructuredOutput
		}
		if output.Confidence < 0 || output.Confidence > 1 {
			return entity.ErrAIInvalidStructuredOutput
		}
		return nil
	})
	return output, err
}

func (s *Service) generateVirtualPartnerRuntime(ctx context.Context, session *entity.TrainingSession, doc entity.ScenarioDocument, state trainingState, history []entity.TrainingMessage, learnerContent json.RawMessage, learnerText string) (entity.VirtualPartnerRuntimeTurn, error) {
	var output entity.VirtualPartnerRuntimeTurn
	actions, err := s.profile.ListSemanticActions(ctx)
	if err != nil {
		return output, err
	}
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
		"task":                "virtual_partner_runtime_turn",
		"channel":             session.Channel,
		"difficulty":          doc.Difficulty,
		"scenario":            map[string]any{"expanded_context": doc.ExpandedContext},
		"current_block":       blockForVirtualPartnerRuntimeAI(doc, current, partnerRole.ID),
		"partner_role":        privateRoleForRuntimeAI(partnerRole),
		"learner_role_public": publicRole(learnerRole),
		"runtime_state":       map[string]any{"turn_number": state.TurnNumber, "patience": state.Patience},
		"history":             historyForRuntimeAI(history, session.Channel, 8),
		"learner_message":     map[string]any{"text": learnerText},
		"semantic_actions":    actions,
		"allowed_transitions": transitionsForRuntimeAI(doc, current),
	}
	err = s.ai.GenerateJSON(ctx, ai.RoleVirtualPartnerRuntime, input, &output, func() error {
		if !semanticActionAllowed(actions, output.LearnerSemanticActionCode) {
			return entity.ErrAIInvalidStructuredOutput
		}
		if output.LearnerConfidence < 0 || output.LearnerConfidence > 1 {
			return entity.ErrAIInvalidStructuredOutput
		}
		if !semanticActionAllowed(actions, output.PartnerSemanticActionCode) {
			return entity.ErrAIInvalidStructuredOutput
		}
		if _, err := normalizePartnerContent(output.Content, session.Channel); err != nil {
			return err
		}
		if output.PatienceDelta < virtualPartnerPatienceDeltaMin || output.PatienceDelta > virtualPartnerPatienceDeltaMax {
			return entity.ErrAIInvalidStructuredOutput
		}
		return validateVirtualPartnerRuntimeTurn(output, current)
	})
	return output, err
}

func (s *Service) generateVirtualPartnerTurn(ctx context.Context, session *entity.TrainingSession, doc entity.ScenarioDocument, state trainingState, history []entity.TrainingMessage, learnerContent json.RawMessage, learnerText string, classification entity.SemanticClassification) (entity.VirtualPartnerTurn, error) {
	var output entity.VirtualPartnerTurn
	actions, err := s.profile.ListSemanticActions(ctx)
	if err != nil {
		return output, err
	}
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
		"current_block":            blockForRuntimeAI(current),
		"partner_role":             partnerRole,
		"learner_role_public":      publicRole(learnerRole),
		"detected_learner_actions": []string{classification.SemanticActionCode},
		"runtime_state":            state,
		"history":                  history,
		"learner_message":          map[string]any{"content": learnerContent, "text": learnerText},
		"semantic_actions":         actions,
		"allowed_transitions":      allowedTransitions(current),
	}
	err = s.ai.GenerateJSON(ctx, ai.RoleVirtualPartner, input, &output, func() error {
		if !semanticActionAllowed(actions, output.PartnerSemanticActionCode) {
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

func (s *Service) evaluateVirtualPartnerState(ctx context.Context, session *entity.TrainingSession, doc entity.ScenarioDocument, state trainingState, history []entity.TrainingMessage, learnerContent json.RawMessage, learnerText string, classification entity.SemanticClassification, turn entity.VirtualPartnerTurn) (entity.TrainingStateEvaluation, error) {
	var output entity.TrainingStateEvaluation
	current, _ := blockByID(doc.Blocks, state.BlockID)
	input := map[string]any{
		"task":                 "evaluate_training_state",
		"channel":              session.Channel,
		"difficulty":           doc.Difficulty,
		"scenario":             map[string]any{"expanded_context": doc.ExpandedContext, "target_cluster_ids": doc.TargetClusterIDs},
		"current_block":        blockForRuntimeAI(current),
		"runtime_state":        state,
		"history":              history,
		"learner_message":      map[string]any{"content": learnerContent, "text": learnerText, "classification": classification},
		"virtual_partner_turn": turn,
		"allowed_transitions":  allowedTransitions(current),
		"terminal_rule":        "Если переход ведёт в end-блок, выбери TRANSITION. Если переговоры явно завершились без подходящего перехода, выбери TERMINATE. Иначе выбери STAY.",
	}
	err := s.ai.GenerateJSON(ctx, ai.RoleTrainingState, input, &output, func() error {
		return validateTrainingStateEvaluation(output, current)
	})
	return output, err
}

func (s *Service) assessVirtualPartnerTraining(ctx context.Context, session *entity.TrainingSession, doc entity.ScenarioDocument, state trainingState, messages []entity.TrainingMessage) (*entity.TrainingAssessment, error) {
	var output entity.TrainingAssessment
	input := map[string]any{
		"task":               "assess_completed_training",
		"channel":            session.Channel,
		"difficulty":         doc.Difficulty,
		"scenario":           scenarioAssessmentInput(doc),
		"runtime_state":      state,
		"messages":           messagesForAssessmentAI(messages, session.Channel),
		"target_cluster_ids": doc.TargetClusterIDs,
		"outcome":            json.RawMessage(outcomeJSON(doc, state)),
		"language":           "ru",
	}
	err := s.ai.GenerateJSON(ctx, ai.RoleTrainingAssessment, input, &output, func() error {
		return validateTrainingAssessment(output, doc.TargetClusterIDs)
	})
	if err != nil {
		return nil, err
	}
	return &output, nil
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

func applyTrainingStateEvaluation(doc entity.ScenarioDocument, state trainingState, evaluation entity.TrainingStateEvaluation) (trainingState, entity.TrainingStatus, error) {
	next := state
	next.TurnNumber++
	if evaluation.StateUpdates.ShortSummary != "" {
		next.ShortSummary = evaluation.StateUpdates.ShortSummary
	}
	next.RevealedFacts = append(next.RevealedFacts, evaluation.StateUpdates.RevealedFacts...)
	status := entity.TrainingStatusInProgress
	switch evaluation.Decision {
	case "STAY":
		return next, status, nil
	case "TRANSITION":
		current, ok := blockByID(doc.Blocks, state.BlockID)
		if !ok || evaluation.TransitionID == nil {
			return next, status, entity.ErrInvalidTrainingState
		}
		transition, ok := transitionByID(current.Transitions, *evaluation.TransitionID)
		if !ok {
			return next, status, entity.ErrAIInvalidStructuredOutput
		}
		next.BlockID = transition.TargetBlockID
		if target, ok := blockByID(doc.Blocks, transition.TargetBlockID); ok && target.Type == "end" {
			status = entity.TrainingStatusCompleted
		}
		return next, status, nil
	case "TERMINATE":
		status = entity.TrainingStatusCompleted
		return next, status, nil
	default:
		return next, status, entity.ErrAIInvalidStructuredOutput
	}
}

func validateTrainingStateEvaluation(evaluation entity.TrainingStateEvaluation, current entity.ScenarioBlock) error {
	switch evaluation.Decision {
	case "STAY":
		if evaluation.TransitionID != nil || evaluation.TerminationReason != nil {
			return entity.ErrAIInvalidStructuredOutput
		}
	case "TRANSITION":
		if evaluation.TransitionID == nil || !transitionAllowed(current, *evaluation.TransitionID) {
			return entity.ErrAIInvalidStructuredOutput
		}
	case "TERMINATE":
		if evaluation.TerminationReason == nil || strings.TrimSpace(*evaluation.TerminationReason) == "" {
			return entity.ErrAIInvalidStructuredOutput
		}
	default:
		return entity.ErrAIInvalidStructuredOutput
	}
	return nil
}

func validateVirtualPartnerRuntimeTurn(turn entity.VirtualPartnerRuntimeTurn, current entity.ScenarioBlock) error {
	switch turn.Decision {
	case "STAY":
		if turn.TransitionID != nil {
			return entity.ErrAIInvalidStructuredOutput
		}
	case "TRANSITION":
		if turn.TransitionID == nil || !transitionAllowed(current, *turn.TransitionID) {
			return entity.ErrAIInvalidStructuredOutput
		}
	case "TERMINATE":
		if turn.TransitionID != nil {
			return entity.ErrAIInvalidStructuredOutput
		}
	default:
		return entity.ErrAIInvalidStructuredOutput
	}
	return nil
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

func transitionByID(transitions []entity.ScenarioTransition, id string) (entity.ScenarioTransition, bool) {
	for _, transition := range transitions {
		if transition.ID == id && !transition.Forbidden {
			return transition, true
		}
	}
	return entity.ScenarioTransition{}, false
}

func semanticActionAllowed(actions []entity.SemanticActionDefinition, code string) bool {
	for _, action := range actions {
		if action.Code == code {
			return true
		}
	}
	return false
}

func clampVirtualPartnerPatience(value int) int {
	if value < virtualPartnerPatienceMin {
		return virtualPartnerPatienceMin
	}
	return value
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

func deterministicEventJSON(block entity.ScenarioBlock, node entity.DeterministicDialogueNode, option entity.DeterministicReplyOption, selected, clientEventID string) json.RawMessage {
	events := make([]assessmentEvent, 0, len(node.AssessmentTargets))
	assessmentByCluster := make(map[string]entity.DeterministicReplyAssessment, len(option.Assessments))
	for _, assessment := range option.Assessments {
		assessmentByCluster[assessment.ClusterID] = assessment
	}
	for _, target := range node.AssessmentTargets {
		assessment, ok := assessmentByCluster[target.ClusterID]
		if !ok {
			continue
		}
		events = append(events, assessmentEvent{
			ClusterID: target.ClusterID,
			Weight:    target.Weight,
			Level:     assessment.Level,
			Score:     levelToScore(assessment.Level),
			Comment:   assessment.Comment,
		})
	}
	transitionID := option.TransitionID
	feedbackComment := ""
	if transitionID != "" {
		if transition, ok := transitionByID(block.Transitions, transitionID); ok {
			feedbackComment = transition.FeedbackComment
		}
	}
	raw, _ := json.Marshal(map[string]any{
		"client_event_id":               clientEventID,
		"block_id":                      block.ID,
		"node_id":                       node.ID,
		"reply_option_id":               option.ID,
		"transition_id":                 transitionID,
		"transition_feedback_comment":   feedbackComment,
		"selected_semantic_action_code": selected,
		"expected_semantic_action_code": option.SemanticActionCode,
		"correct":                       selected == option.SemanticActionCode,
		"assessment_events":             events,
	})
	return raw
}

type assessmentEvent struct {
	ClusterID string  `json:"cluster_id"`
	Weight    float64 `json:"weight"`
	Level     int     `json:"level"`
	Score     float64 `json:"score"`
	Comment   string  `json:"comment,omitempty"`
}

func outcomeJSON(doc entity.ScenarioDocument, state trainingState) json.RawMessage {
	if block, ok := blockByID(doc.Blocks, state.BlockID); ok && block.Type == "end" && block.Result != nil {
		raw, _ := json.Marshal(block.Result)
		return raw
	}
	raw, _ := json.Marshal(map[string]any{"summary": "Тренировка завершена.", "goals_met": []string{}, "goals_missed": []string{}})
	return raw
}

type trainingAnalysisExtras struct {
	Process            []entity.TrainingProcessAssessment
	Timeline           []entity.TrainingTimelineItem
	MessageFeedback    []entity.TrainingMessageFeedback
	Highlights         entity.TrainingHighlights
	Strengths          []entity.TrainingAnalysisNote
	Improvements       []entity.TrainingAnalysisNote
	ClusterAssessments []entity.TrainingClusterAssessment
}

func analysisJSON(sessionID string, score float64, outcome json.RawMessage, doc entity.ScenarioDocument, messages []entity.TrainingMessage, results []entity.TrainingClusterResult, changes []entity.ProfileClusterChange, extras *trainingAnalysisExtras) json.RawMessage {
	timeline := analysisTimeline(doc, messages, extras)
	messageFeedback := analysisMessageFeedback(messages, extras)
	opponentClassifications := analysisOpponentClassifications(messages)
	highlights := analysisHighlights(messageFeedback, results, extras)
	process := legacyProcessFromTimeline(timeline, extras)
	competencies := analysisCompetencies(messages, results, extras)
	strengths := legacyNotesFromHighlights(highlights.Strengths, "Сильная сторона")
	improvements := legacyNotesFromHighlights(highlights.Improvements, "Зона роста")
	raw, _ := json.Marshal(map[string]any{
		"session_id":               sessionID,
		"score":                    floorScore(score),
		"result":                   json.RawMessage(outcome),
		"timeline":                 timeline,
		"message_feedback":         messageFeedback,
		"opponent_classifications": opponentClassifications,
		"highlights":               highlights,
		"process":                  process,
		"strengths":                strengths,
		"improvements":             improvements,
		"competencies":             competencies,
		"profile_change":           map[string]any{"changes": changes},
	})
	return raw
}

type deterministicEvent struct {
	ClientEventID              string            `json:"client_event_id"`
	BlockID                    string            `json:"block_id"`
	NodeID                     string            `json:"node_id"`
	ReplyOptionID              string            `json:"reply_option_id"`
	TransitionID               string            `json:"transition_id"`
	TransitionFeedbackComment  string            `json:"transition_feedback_comment"`
	SelectedSemanticActionCode string            `json:"selected_semantic_action_code"`
	ExpectedSemanticActionCode string            `json:"expected_semantic_action_code"`
	Correct                    bool              `json:"correct"`
	AssessmentEvents           []assessmentEvent `json:"assessment_events"`
}

func deterministicEventFromMessage(message entity.TrainingMessage) deterministicEvent {
	var event deterministicEvent
	if len(message.Classification) > 0 {
		_ = json.Unmarshal(message.Classification, &event)
	}
	return event
}

func analysisTimeline(doc entity.ScenarioDocument, messages []entity.TrainingMessage, extras *trainingAnalysisExtras) []map[string]any {
	if extras != nil {
		items := make([]map[string]any, 0, len(extras.Timeline))
		for _, item := range extras.Timeline {
			items = append(items, map[string]any{
				"order":                 item.Order,
				"block_id":              item.BlockID,
				"block_name":            item.BlockName,
				"transition_id":         item.TransitionID,
				"from_block_id":         item.FromBlockID,
				"to_block_id":           item.ToBlockID,
				"message_ids":           item.MessageIDs,
				"summary":               item.Summary,
				"outcome":               normalizeOutcome(item.Outcome, 0),
				"semantic_action_codes": item.SemanticActionCodes,
			})
		}
		return items
	}
	items := []map[string]any{}
	order := 1
	for _, message := range messages {
		if message.Author != "learner" || message.SemanticActionCode == nil {
			continue
		}
		event := deterministicEventFromMessage(message)
		block, _ := blockByID(doc.Blocks, event.BlockID)
		transitionID := nullableString(event.TransitionID)
		toBlockID := ""
		if event.TransitionID != "" {
			if transition, ok := transitionByID(block.Transitions, event.TransitionID); ok {
				toBlockID = transition.TargetBlockID
			}
		}
		score := deterministicMessageScore(event)
		items = append(items, map[string]any{
			"order":                 order,
			"block_id":              event.BlockID,
			"block_name":            block.Name,
			"transition_id":         transitionID,
			"from_block_id":         event.BlockID,
			"to_block_id":           toBlockID,
			"message_ids":           []string{message.ID},
			"summary":               deterministicSummary(message, event),
			"outcome":               normalizeOutcome("", score),
			"semantic_action_codes": []string{*message.SemanticActionCode},
		})
		order++
	}
	return items
}

func analysisMessageFeedback(messages []entity.TrainingMessage, extras *trainingAnalysisExtras) []map[string]any {
	if extras != nil {
		items := make([]map[string]any, 0, len(extras.MessageFeedback))
		for _, item := range extras.MessageFeedback {
			items = append(items, map[string]any{
				"message_id":           item.MessageID,
				"block_id":             item.BlockID,
				"semantic_action_code": item.SemanticActionCode,
				"score":                floorScore(float64(item.Score)),
				"level":                normalizeLevel(item.Level, item.Score),
				"summary":              item.Summary,
				"strengths":            item.Strengths,
				"improvements":         item.Improvements,
				"related_cluster_ids":  item.RelatedClusterIDs,
			})
		}
		return items
	}
	items := []map[string]any{}
	for _, message := range messages {
		if message.Author != "learner" || message.SemanticActionCode == nil {
			continue
		}
		event := deterministicEventFromMessage(message)
		score := deterministicMessageScore(event)
		strengths, improvements := deterministicMessageNotes(message, event, score)
		items = append(items, map[string]any{
			"message_id":           message.ID,
			"block_id":             event.BlockID,
			"semantic_action_code": *message.SemanticActionCode,
			"score":                floorScore(score),
			"level":                scoreLevel(score),
			"summary":              deterministicSummary(message, event),
			"strengths":            strengths,
			"improvements":         improvements,
			"related_cluster_ids":  assessmentClusterIDs(event.AssessmentEvents),
		})
	}
	return items
}

func analysisOpponentClassifications(messages []entity.TrainingMessage) []map[string]any {
	items := []map[string]any{}
	for _, message := range messages {
		if message.Author == "learner" || len(message.Classification) == 0 {
			continue
		}
		var classification struct {
			SelectedSemanticActionCode string `json:"selected_semantic_action_code"`
			ExpectedSemanticActionCode string `json:"expected_semantic_action_code"`
			Correct                    bool   `json:"correct"`
		}
		if json.Unmarshal(message.Classification, &classification) != nil || classification.SelectedSemanticActionCode == "" || classification.ExpectedSemanticActionCode == "" {
			continue
		}
		items = append(items, map[string]any{
			"message_id":                    message.ID,
			"selected_semantic_action_code": classification.SelectedSemanticActionCode,
			"expected_semantic_action_code": classification.ExpectedSemanticActionCode,
			"correct":                       classification.Correct,
		})
	}
	return items
}

func analysisHighlights(messageFeedback []map[string]any, results []entity.TrainingClusterResult, extras *trainingAnalysisExtras) entity.TrainingHighlights {
	if extras != nil && (len(extras.Highlights.Strengths) > 0 || len(extras.Highlights.Improvements) > 0) {
		return extras.Highlights
	}
	highlights := entity.TrainingHighlights{}
	for _, item := range messageFeedback {
		score, _ := item["score"].(int)
		messageID, _ := item["message_id"].(string)
		clusterIDs, _ := item["related_cluster_ids"].([]string)
		summary, _ := item["summary"].(string)
		if score >= 75 && len(highlights.Strengths) < 3 {
			highlights.Strengths = append(highlights.Strengths, entity.TrainingHighlight{
				Title:       "Сильный ход в переговорах",
				Description: summary,
				MessageIDs:  []string{messageID},
				ClusterIDs:  clusterIDs,
			})
		}
		if score < 70 && len(highlights.Improvements) < 3 {
			highlights.Improvements = append(highlights.Improvements, entity.TrainingHighlight{
				Title:       "Что улучшить",
				Description: summary,
				MessageIDs:  []string{messageID},
				ClusterIDs:  clusterIDs,
			})
		}
	}
	if len(highlights.Strengths) == 0 {
		for _, result := range results {
			if result.Score < 75 {
				continue
			}
			highlights.Strengths = append(highlights.Strengths, entity.TrainingHighlight{Title: "Сильная компетенция", Description: result.Comment, ClusterIDs: []string{result.ClusterID}})
			break
		}
	}
	if len(highlights.Improvements) == 0 {
		for _, result := range results {
			if result.Score >= 75 {
				continue
			}
			highlights.Improvements = append(highlights.Improvements, entity.TrainingHighlight{Title: "Зона роста", Description: competencyRecommendation(result.Score), ClusterIDs: []string{result.ClusterID}})
			break
		}
	}
	return highlights
}

func analysisCompetencies(messages []entity.TrainingMessage, results []entity.TrainingClusterResult, extras *trainingAnalysisExtras) []map[string]any {
	assessments := map[string]entity.TrainingClusterAssessment{}
	if extras != nil {
		for _, item := range extras.ClusterAssessments {
			assessments[item.ClusterID] = item
		}
	}
	items := make([]map[string]any, 0, len(results))
	for _, result := range results {
		assessment, hasAssessment := assessments[result.ClusterID]
		evidence := []map[string]any{}
		summary := cleanComment(result.Comment)
		recommendation := competencyRecommendation(result.Score)
		comment := summary
		if hasAssessment {
			summary = firstNonEmpty(assessment.Summary, assessment.Comment, summary)
			comment = firstNonEmpty(assessment.Comment, summary)
			recommendation = firstNonEmpty(assessment.Recommendation, recommendation)
			for _, episode := range assessment.EvidenceEpisodes {
				evidence = append(evidence, map[string]any{"message_ids": episode.MessageIDs, "summary": episode.Summary, "semantic_action_codes": episode.SemanticActionCodes})
			}
		} else {
			evidence = deterministicClusterEvidence(messages, result.ClusterID)
		}
		items = append(items, map[string]any{
			"cluster_id":      result.ClusterID,
			"score":           floorScore(result.Score),
			"level":           scoreLevel(result.Score),
			"summary":         summary,
			"comment":         comment,
			"evidence":        evidence,
			"recommendation":  recommendation,
			"evidence_weight": result.EvidenceWeight,
		})
	}
	return items
}

func legacyProcessFromTimeline(timeline []map[string]any, extras *trainingAnalysisExtras) []map[string]any {
	if extras != nil && len(extras.Process) > 0 {
		items := make([]map[string]any, 0, len(extras.Process))
		for _, item := range extras.Process {
			items = append(items, map[string]any{"order": item.Order, "summary": item.Summary, "semantic_action_codes": item.SemanticActionCodes})
		}
		return items
	}
	process := make([]map[string]any, 0, len(timeline))
	for _, item := range timeline {
		process = append(process, map[string]any{"order": item["order"], "summary": item["summary"], "semantic_action_codes": item["semantic_action_codes"]})
	}
	return process
}

func deterministicClusterResults(messages []entity.TrainingMessage) []entity.TrainingClusterResult {
	type aggregate struct {
		weightedScore float64
		weight        float64
		comments      []string
	}
	aggregates := map[string]*aggregate{}
	for _, message := range messages {
		event := deterministicEventFromMessage(message)
		for _, assessment := range event.AssessmentEvents {
			if assessment.ClusterID == "" || assessment.Weight <= 0 {
				continue
			}
			item := aggregates[assessment.ClusterID]
			if item == nil {
				item = &aggregate{}
				aggregates[assessment.ClusterID] = item
			}
			item.weight += assessment.Weight
			item.weightedScore += assessment.Score * assessment.Weight
			if strings.TrimSpace(assessment.Comment) != "" {
				item.comments = append(item.comments, strings.TrimSpace(assessment.Comment))
			}
		}
	}
	results := make([]entity.TrainingClusterResult, 0, len(aggregates))
	for clusterID, item := range aggregates {
		if item.weight <= 0 {
			continue
		}
		results = append(results, entity.TrainingClusterResult{ClusterID: clusterID, Score: item.weightedScore / item.weight, EvidenceWeight: math.Min(1, item.weight/3.0), Comment: cleanComment(strings.Join(item.comments, " "))})
	}
	return results
}

func deterministicMessageScore(event deterministicEvent) float64 {
	if len(event.AssessmentEvents) == 0 {
		if event.Correct {
			return 80
		}
		return 45
	}
	sum := 0.0
	weight := 0.0
	for _, assessment := range event.AssessmentEvents {
		w := assessment.Weight
		if w <= 0 {
			w = 1
		}
		sum += assessment.Score * w
		weight += w
	}
	if weight <= 0 {
		return 0
	}
	return sum / weight
}

func deterministicSummary(message entity.TrainingMessage, event deterministicEvent) string {
	if strings.TrimSpace(event.TransitionFeedbackComment) != "" {
		return strings.TrimSpace(event.TransitionFeedbackComment)
	}
	text := messageText(message.Content, entity.TrainingChannelChat)
	switch {
	case event.Correct && text != "":
		return "Выбранная реплика соответствует ожидаемому переговорному действию и помогает продвинуть диалог: " + shortenText(text, 160)
	case !event.Correct && text != "":
		return "Реплика понятна, но выбранная semantic-метка не совпала с ожидаемой. Стоит точнее определять действие: " + shortenText(text, 160)
	case event.Correct:
		return "Выбранная реплика соответствует ожидаемому переговорному действию."
	default:
		return "Выбранная semantic-метка не совпала с ожидаемой."
	}
}

func deterministicMessageNotes(message entity.TrainingMessage, event deterministicEvent, score float64) ([]string, []string) {
	strengths := []string{}
	improvements := []string{}
	for _, assessment := range event.AssessmentEvents {
		comment := strings.TrimSpace(assessment.Comment)
		if comment == "" {
			continue
		}
		if assessment.Score >= 75 {
			strengths = append(strengths, comment)
		} else {
			improvements = append(improvements, comment)
		}
	}
	if len(strengths) == 0 && score >= 75 {
		strengths = append(strengths, "Реплика продвигает переговоры и соответствует задаче текущего состояния.")
	}
	if len(improvements) == 0 && score < 70 {
		improvements = append(improvements, "Можно точнее связать реплику с интересами, границами или встречной ценностью.")
	}
	return uniqueStrings(strengths), uniqueStrings(improvements)
}

func deterministicClusterEvidence(messages []entity.TrainingMessage, clusterID string) []map[string]any {
	evidence := []map[string]any{}
	for _, message := range messages {
		if message.Author != "learner" || message.SemanticActionCode == nil {
			continue
		}
		event := deterministicEventFromMessage(message)
		for _, assessment := range event.AssessmentEvents {
			if assessment.ClusterID != clusterID {
				continue
			}
			evidence = append(evidence, map[string]any{
				"message_ids":           []string{message.ID},
				"summary":               firstNonEmpty(strings.TrimSpace(assessment.Comment), deterministicSummary(message, event)),
				"semantic_action_codes": []string{*message.SemanticActionCode},
			})
		}
	}
	return evidence
}

func assessmentClusterIDs(events []assessmentEvent) []string {
	ids := make([]string, 0, len(events))
	for _, event := range events {
		if event.ClusterID != "" {
			ids = append(ids, event.ClusterID)
		}
	}
	return uniqueStrings(ids)
}

func virtualPartnerClusterResults(assessment entity.TrainingAssessment, targetClusterIDs []string) []entity.TrainingClusterResult {
	targets := stringSet(targetClusterIDs)
	results := make([]entity.TrainingClusterResult, 0, len(assessment.ClusterAssessments))
	for _, item := range assessment.ClusterAssessments {
		if !targets[item.ClusterID] {
			continue
		}
		if item.Status == "NOT_OBSERVED" {
			continue
		}
		if item.Status != "ASSESSED" {
			continue
		}
		evidenceStrength := virtualPartnerEvidenceStrength(len(item.EvidenceEpisodes))
		results = append(results, entity.TrainingClusterResult{
			ClusterID:      item.ClusterID,
			Score:          levelToScore(item.Level),
			EvidenceWeight: evidenceStrength * 0.8,
			Comment:        strings.TrimSpace(item.Comment),
		})
	}
	return results
}

func ensureTerminatedVirtualPartnerResults(doc entity.ScenarioDocument, state trainingState, results []entity.TrainingClusterResult) []entity.TrainingClusterResult {
	block, ok := blockByID(doc.Blocks, state.BlockID)
	if ok && block.Type == "end" {
		return results
	}
	if len(doc.TargetClusterIDs) == 0 {
		return results
	}
	weightedClusters := map[string]bool{}
	for _, result := range results {
		if result.EvidenceWeight > 0 {
			weightedClusters[result.ClusterID] = true
		}
	}
	items := append([]entity.TrainingClusterResult(nil), results...)
	for _, clusterID := range doc.TargetClusterIDs {
		if weightedClusters[clusterID] {
			continue
		}
		items = append(items, entity.TrainingClusterResult{
			ClusterID:      clusterID,
			Score:          20,
			EvidenceWeight: 0.3,
			Comment:        "Тренировка завершилась досрочно: поведение участника сорвало переговоры до достижения финального состояния.",
		})
	}
	return items
}

func virtualPartnerEvidenceStrength(episodes int) float64 {
	switch {
	case episodes <= 0:
		return 0
	case episodes == 1:
		return 0.4
	case episodes == 2:
		return 0.7
	default:
		return 1
	}
}

func analysisExtrasFromAssessment(assessment entity.TrainingAssessment) *trainingAnalysisExtras {
	return &trainingAnalysisExtras{
		Process:            assessment.Process,
		Timeline:           assessment.Timeline,
		MessageFeedback:    assessment.MessageFeedback,
		Highlights:         assessment.Highlights,
		Strengths:          assessment.Strengths,
		Improvements:       assessment.Improvements,
		ClusterAssessments: assessment.ClusterAssessments,
	}
}

func validateTrainingAssessment(assessment entity.TrainingAssessment, targetClusterIDs []string) error {
	for _, item := range assessment.Timeline {
		if strings.TrimSpace(item.Summary) == "" {
			return entity.ErrAIInvalidStructuredOutput
		}
	}
	for _, item := range assessment.MessageFeedback {
		if strings.TrimSpace(item.MessageID) == "" || strings.TrimSpace(item.Summary) == "" {
			return entity.ErrAIInvalidStructuredOutput
		}
	}
	targets := stringSet(targetClusterIDs)
	seen := map[string]bool{}
	for _, item := range assessment.ClusterAssessments {
		if !targets[item.ClusterID] || seen[item.ClusterID] {
			return entity.ErrAIInvalidStructuredOutput
		}
		seen[item.ClusterID] = true
		switch item.Status {
		case "ASSESSED":
			if item.Level < 0 || item.Level > 4 || len(item.EvidenceEpisodes) == 0 || strings.TrimSpace(item.Summary) == "" || strings.TrimSpace(item.Comment) == "" || strings.TrimSpace(item.Recommendation) == "" {
				return entity.ErrAIInvalidStructuredOutput
			}
		case "NOT_OBSERVED":
			if item.Level != 0 || len(item.EvidenceEpisodes) != 0 || strings.TrimSpace(item.Comment) == "" {
				return entity.ErrAIInvalidStructuredOutput
			}
		default:
			return entity.ErrAIInvalidStructuredOutput
		}
	}
	for _, clusterID := range targetClusterIDs {
		if !seen[clusterID] {
			return entity.ErrAIInvalidStructuredOutput
		}
	}
	return nil
}

func scenarioAssessmentInput(doc entity.ScenarioDocument) map[string]any {
	return map[string]any{
		"scenario_id":        doc.ScenarioID,
		"title":              doc.Title,
		"description":        doc.Description,
		"expanded_context":   doc.ExpandedContext,
		"difficulty":         doc.Difficulty,
		"target_cluster_ids": doc.TargetClusterIDs,
		"roles":              doc.Roles,
		"blocks":             blocksForRuntimeAI(doc.Blocks),
	}
}

func blocksForRuntimeAI(blocks []entity.ScenarioBlock) []map[string]any {
	items := make([]map[string]any, 0, len(blocks))
	for _, block := range blocks {
		items = append(items, blockForRuntimeAI(block))
	}
	return items
}

func blockForRuntimeAI(block entity.ScenarioBlock) map[string]any {
	return map[string]any{
		"id":               block.ID,
		"type":             block.Type,
		"name":             block.Name,
		"description":      block.Description,
		"expanded_context": block.ExpandedContext,
		"lead_role_id":     block.LeadRoleID,
		"role_goals":       block.RoleGoals,
		"role_behaviors":   block.RoleBehaviors,
		"behavior":         block.Behavior,
		"action_rules":     block.ActionRules,
		"constraints":      block.Constraints,
	}
}

func blockForVirtualPartnerRuntimeAI(doc entity.ScenarioDocument, block entity.ScenarioBlock, partnerRoleID string) map[string]any {
	return map[string]any{
		"name":             block.Name,
		"description":      block.Description,
		"expanded_context": block.ExpandedContext,
		"lead_role_id":     block.LeadRoleID,
		"role_goals":       block.RoleGoals,
		"role_behavior":    roleBehaviorForRuntimeAI(block, partnerRoleID),
		"transitions":      transitionsForRuntimeAI(doc, block),
	}
}

func roleBehaviorForRuntimeAI(block entity.ScenarioBlock, roleID string) string {
	for _, behavior := range block.RoleBehaviors {
		if behavior.RoleID == roleID {
			return strings.TrimSpace(behavior.Behavior)
		}
	}
	return strings.TrimSpace(block.Behavior)
}

func privateRoleForRuntimeAI(role entity.ScenarioRole) map[string]any {
	return map[string]any{
		"id":                  role.ID,
		"name":                role.Name,
		"description":         role.Description,
		"goal":                role.Goal,
		"interests":           role.Interests,
		"constraints":         role.Constraints,
		"private_information": role.PrivateInformation,
	}
}

func transitionsForRuntimeAI(doc entity.ScenarioDocument, block entity.ScenarioBlock) []map[string]any {
	items := []map[string]any{}
	for _, transition := range block.Transitions {
		if transition.Forbidden {
			continue
		}
		targetType := ""
		if target, ok := blockByID(doc.Blocks, transition.TargetBlockID); ok {
			targetType = target.Type
		}
		items = append(items, map[string]any{
			"id":              transition.ID,
			"condition":       transition.Condition,
			"target_block_id": transition.TargetBlockID,
			"target_type":     targetType,
		})
	}
	return items
}

func historyForRuntimeAI(messages []entity.TrainingMessage, channel entity.TrainingChannel, limit int) []map[string]any {
	if limit > 0 && len(messages) > limit {
		messages = messages[len(messages)-limit:]
	}
	items := make([]map[string]any, 0, len(messages))
	for _, message := range messages {
		item := map[string]any{"author": message.Author, "text": messageText(message.Content, channel)}
		if message.SemanticActionCode != nil {
			item["semantic_action_code"] = *message.SemanticActionCode
		}
		items = append(items, item)
	}
	return items
}

func messagesForAssessmentAI(messages []entity.TrainingMessage, channel entity.TrainingChannel) []map[string]any {
	items := make([]map[string]any, 0, len(messages))
	for _, message := range messages {
		item := map[string]any{
			"message_id":           message.ID,
			"id":                   message.ID,
			"session_id":           message.SessionID,
			"author":               message.Author,
			"author_label":         assessmentAuthorLabel(message.Author),
			"content":              message.Content,
			"text":                 messageText(message.Content, channel),
			"semantic_action_code": message.SemanticActionCode,
			"semantic_annotations": message.SemanticAnnotations,
			"expected_action_code": message.ExpectedActionCode,
			"classification":       message.Classification,
			"created_at":           message.CreatedAt,
		}
		items = append(items, item)
	}
	return items
}

func assessmentAuthorLabel(author string) string {
	switch author {
	case "learner":
		return "ученик"
	case "virtual_partner":
		return "виртуальный партнёр"
	default:
		return author
	}
}

func messageText(raw json.RawMessage, channel entity.TrainingChannel) string {
	var payload map[string]any
	if json.Unmarshal(raw, &payload) != nil {
		return ""
	}
	switch channel {
	case entity.TrainingChannelEmail:
		return strings.TrimSpace(stringFromAny(payload["subject"]) + "\n" + stringFromAny(payload["body"]))
	case entity.TrainingChannelVoiceMessage:
		text := strings.TrimSpace(stringFromAny(payload["transcript"]))
		if text == "" {
			text = strings.TrimSpace(stringFromAny(payload["spoken_text"]))
		}
		return text
	default:
		return strings.TrimSpace(stringFromAny(payload["text"]))
	}
}

func notesToMaps(notes []entity.TrainingAnalysisNote) []map[string]string {
	items := make([]map[string]string, 0, len(notes))
	for _, note := range notes {
		items = append(items, map[string]string{"title": note.Title, "description": note.Description})
	}
	return items
}

func legacyNotesFromHighlights(highlights []entity.TrainingHighlight, fallbackTitle string) []map[string]string {
	items := make([]map[string]string, 0, len(highlights))
	for _, item := range highlights {
		title := strings.TrimSpace(item.Title)
		if title == "" {
			title = fallbackTitle
		}
		items = append(items, map[string]string{"title": title, "description": strings.TrimSpace(item.Description)})
	}
	return items
}

func stringSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}

func overallTrainingScore(results []entity.TrainingClusterResult) float64 {
	if len(results) == 0 {
		return 0
	}
	sum := 0.0
	for _, result := range results {
		sum += result.Score
	}
	return sum / float64(len(results))
}

func levelToScore(level int) float64 {
	switch {
	case level <= 0:
		return 0
	case level == 1:
		return 25
	case level == 2:
		return 50
	case level == 3:
		return 75
	default:
		return 100
	}
}

func floorScore(value float64) int {
	if value < 0 || math.IsNaN(value) {
		return 0
	}
	if value > 100 {
		return 100
	}
	return int(math.Floor(value))
}

func scoreLevel(score float64) string {
	switch {
	case score >= 85:
		return "excellent"
	case score >= 70:
		return "strong"
	case score >= 50:
		return "needs_attention"
	default:
		return "weak"
	}
}

func normalizeLevel(level string, score int) string {
	switch strings.TrimSpace(level) {
	case "excellent", "strong", "needs_attention", "weak":
		return level
	default:
		return scoreLevel(float64(score))
	}
}

func normalizeOutcome(outcome string, score float64) string {
	switch strings.TrimSpace(outcome) {
	case "positive", "neutral", "negative":
		return outcome
	}
	switch {
	case score >= 75:
		return "positive"
	case score >= 50:
		return "neutral"
	default:
		return "negative"
	}
}

func competencyRecommendation(score float64) string {
	switch {
	case score >= 85:
		return "Сохраните этот подход: он хорошо работает в подобных переговорах."
	case score >= 70:
		return "Продолжайте развивать навык: добавляйте больше конкретики и фиксируйте договорённости."
	case score >= 50:
		return "В следующей тренировке заранее формулируйте цель хода и проверяйте, какую ценность вы отдаёте или получаете."
	default:
		return "Начните с базового шага: проясните интересы, обозначьте границы и только потом переходите к предложениям."
	}
}

func nullableString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func shortenText(value string, max int) string {
	value = strings.Join(strings.Fields(value), " ")
	if max <= 0 || len([]rune(value)) <= max {
		return value
	}
	runes := []rune(value)
	return string(runes[:max]) + "…"
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func cleanComment(comment string) string {
	parts := strings.Split(comment, ".")
	unique := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || seen[part] {
			continue
		}
		seen[part] = true
		unique = append(unique, part)
	}
	if len(unique) == 0 {
		return strings.TrimSpace(comment)
	}
	return strings.Join(unique, ". ") + "."
}

func trainingStrengths(results []entity.TrainingClusterResult) []map[string]string {
	items := []map[string]string{}
	for _, result := range results {
		if result.Score >= 75 {
			items = append(items, map[string]string{"title": "Сильное проявление компетенции", "description": result.Comment})
		}
	}
	return items
}

func trainingImprovements(results []entity.TrainingClusterResult) []map[string]string {
	items := []map[string]string{}
	for _, result := range results {
		if result.Score < 75 {
			description := result.Comment
			if description == "" {
				description = "В этой компетенции есть пространство для улучшения."
			}
			items = append(items, map[string]string{"title": "Зона роста", "description": description})
		}
	}
	return items
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
