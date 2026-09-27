package scenario

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/ai"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
)

type aiGenerator interface {
	GenerateJSON(ctx context.Context, role ai.Role, input any, output any, validate func() error) error
}

type aiGenerationRunner func(ctx context.Context) (json.RawMessage, error)

func (s *scenarioService) StartExpandContext(ctx context.Context, claims entity.AccessClaims, scenarioID string, options entity.AIGenerationOptions) (*entity.AIGenerationRequest, error) {
	if err := s.ensureGenerationRequestAllowed(ctx, claims, scenarioID, ""); err != nil {
		return nil, err
	}
	return s.startAIGenerationRequest(claims, scenarioID, "", "", entity.AIGenerationKindExpandContext, func(ctx context.Context) (json.RawMessage, error) {
		expanded, err := s.ExpandContext(ctx, claims, scenarioID, options)
		if err != nil {
			return nil, err
		}
		return json.Marshal(map[string]any{"expanded_context": expanded})
	})
}

func (s *scenarioService) StartGenerateDraft(ctx context.Context, claims entity.AccessClaims, scenarioID string, options entity.AIGenerationOptions) (*entity.AIGenerationRequest, error) {
	if err := s.ensureGenerationRequestAllowed(ctx, claims, scenarioID, ""); err != nil {
		return nil, err
	}
	return s.startAIGenerationRequest(claims, scenarioID, "", "", entity.AIGenerationKindGenerateDraft, func(ctx context.Context) (json.RawMessage, error) {
		proposal, validation, err := s.GenerateDraft(ctx, claims, scenarioID, options)
		if err != nil {
			return nil, err
		}
		return json.Marshal(map[string]any{"proposal": map[string]any{"expanded_context": proposal.ExpandedContext, "blocks": blocksForGenerationAPI(proposal.Blocks)}, "validation": validation})
	})
}

func (s *scenarioService) StartGenerateBlock(ctx context.Context, claims entity.AccessClaims, scenarioID, blockID string, options entity.AIGenerationOptions) (*entity.AIGenerationRequest, error) {
	if err := s.ensureGenerationRequestAllowed(ctx, claims, scenarioID, blockID); err != nil {
		return nil, err
	}
	return s.startAIGenerationRequest(claims, scenarioID, blockID, "", entity.AIGenerationKindGenerateBlock, func(ctx context.Context) (json.RawMessage, error) {
		proposal, validation, err := s.GenerateBlock(ctx, claims, scenarioID, blockID, options)
		if err != nil {
			return nil, err
		}
		return json.Marshal(map[string]any{"block_id": blockID, "proposal": blockProposalForGenerationAPI(*proposal), "validation": validation})
	})
}

func (s *scenarioService) StartGenerateDialogue(ctx context.Context, claims entity.AccessClaims, scenarioID, blockID, learnerRoleID string, options entity.AIGenerationOptions) (*entity.AIGenerationRequest, error) {
	if err := s.ensureGenerationRequestAllowed(ctx, claims, scenarioID, blockID); err != nil {
		return nil, err
	}
	return s.startAIGenerationRequest(claims, scenarioID, blockID, learnerRoleID, entity.AIGenerationKindGenerateDialogue, func(ctx context.Context) (json.RawMessage, error) {
		dialogue, validation, err := s.GenerateDialogue(ctx, claims, scenarioID, blockID, learnerRoleID, options)
		if err != nil {
			return nil, err
		}
		return json.Marshal(map[string]any{"block_id": blockID, "learner_role_id": learnerRoleID, "proposal": dialogue, "validation": validation})
	})
}

func (s *scenarioService) GetAIGenerationRequest(ctx context.Context, claims entity.AccessClaims, scenarioID, requestID string) (*entity.AIGenerationRequest, error) {
	if err := requireMethodist(claims); err != nil {
		return nil, err
	}
	scenario, err := s.scenarioRepository.GetScenarioByID(ctx, scenarioID)
	if err != nil {
		return nil, err
	}
	if err := ensureOwnedBy(claims.UserID, scenario); err != nil {
		return nil, err
	}
	s.generationMu.RLock()
	job, ok := s.generationRequests[requestID]
	s.generationMu.RUnlock()
	if !ok || job.ScenarioID != scenarioID || job.UserID != claims.UserID {
		return nil, entity.ErrAIGenerationRequestNotFound
	}
	return cloneAIGenerationRequest(job), nil
}

func (s *scenarioService) ensureGenerationRequestAllowed(ctx context.Context, claims entity.AccessClaims, scenarioID, blockID string) error {
	if err := requireMethodist(claims); err != nil {
		return err
	}
	scenario, err := s.scenarioRepository.GetScenarioByID(ctx, scenarioID)
	if err != nil {
		return err
	}
	if err := ensureOwnedBy(claims.UserID, scenario); err != nil {
		return err
	}
	if scenario.Status != entity.ScenarioStatusDraft {
		return entity.ErrScenarioNotDraft
	}
	if blockID == "" {
		return nil
	}
	doc := entity.ScenarioDocumentFromScenario(*scenario)
	if _, _, ok := findBlock(doc.Blocks, blockID); !ok {
		return entity.ErrBlockNotFound
	}
	return nil
}

func (s *scenarioService) startAIGenerationRequest(claims entity.AccessClaims, scenarioID, blockID, learnerRoleID string, kind entity.AIGenerationKind, runner aiGenerationRunner) (*entity.AIGenerationRequest, error) {
	id, err := entity.NewUUID()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	job := &entity.AIGenerationRequest{ID: id, ScenarioID: scenarioID, UserID: claims.UserID, BlockID: blockID, LearnerRoleID: learnerRoleID, Kind: kind, Status: entity.AIGenerationStatusPending, CreatedAt: now, UpdatedAt: now}
	s.generationMu.Lock()
	s.generationRequests[id] = job
	s.generationMu.Unlock()
	response := cloneAIGenerationRequest(job)
	go s.runAIGenerationRequest(id, runner)
	return response, nil
}

func (s *scenarioService) runAIGenerationRequest(id string, runner aiGenerationRunner) {
	s.updateAIGenerationRequest(id, func(job *entity.AIGenerationRequest) {
		job.Status = entity.AIGenerationStatusRunning
	})
	result, err := runner(context.Background())
	s.updateAIGenerationRequest(id, func(job *entity.AIGenerationRequest) {
		if err != nil {
			job.Status = entity.AIGenerationStatusFailed
			job.ErrorCode = aiGenerationJobErrorCode(err)
			job.ErrorMessage = err.Error()
			job.Result = nil
			return
		}
		job.Status = entity.AIGenerationStatusSucceeded
		job.Result = result
		job.ErrorCode = ""
		job.ErrorMessage = ""
	})
}

func (s *scenarioService) updateAIGenerationRequest(id string, update func(*entity.AIGenerationRequest)) {
	s.generationMu.Lock()
	defer s.generationMu.Unlock()
	job := s.generationRequests[id]
	if job == nil {
		return
	}
	update(job)
	job.UpdatedAt = time.Now()
}

func cloneAIGenerationRequest(job *entity.AIGenerationRequest) *entity.AIGenerationRequest {
	if job == nil {
		return nil
	}
	clone := *job
	if len(job.Result) > 0 {
		clone.Result = append(json.RawMessage(nil), job.Result...)
	}
	return &clone
}

func aiGenerationJobErrorCode(err error) string {
	switch {
	case errors.Is(err, entity.ErrAIGenerationContextIncomplete):
		return "AI_GENERATION_CONTEXT_INCOMPLETE"
	case errors.Is(err, entity.ErrAIInvalidStructuredOutput):
		return "AI_INVALID_STRUCTURED_OUTPUT"
	case errors.Is(err, entity.ErrAIProviderUnavailable):
		return "AI_PROVIDER_UNAVAILABLE"
	case errors.Is(err, entity.ErrAIGenerationTimeout):
		return "AI_GENERATION_TIMEOUT"
	default:
		return "AI_GENERATION_FAILED"
	}
}

func blocksForGenerationAPI(blocks []entity.ScenarioBlock) []map[string]any {
	items := make([]map[string]any, 0, len(blocks))
	for _, block := range blocks {
		item := map[string]any{
			"id":             block.ID,
			"type":           block.Type,
			"name":           block.Name,
			"description":    block.Description,
			"role_goals":     block.RoleGoals,
			"role_behaviors": block.RoleBehaviors,
			"transitions":    block.Transitions,
		}
		if strings.TrimSpace(block.ExpandedContext) != "" {
			item["expanded_context"] = block.ExpandedContext
		}
		if strings.TrimSpace(block.LeadRoleID) != "" {
			item["lead_role_id"] = block.LeadRoleID
		}
		if block.Result != nil {
			item["result"] = block.Result
		}
		if block.DeterministicDialogue != nil {
			item["deterministic_dialogue"] = block.DeterministicDialogue
		}
		items = append(items, item)
	}
	return items
}

func blockProposalForGenerationAPI(block entity.BlockContentProposal) map[string]any {
	item := map[string]any{
		"name":           block.Name,
		"description":    block.Description,
		"role_goals":     block.RoleGoals,
		"role_behaviors": block.RoleBehaviors,
	}
	if strings.TrimSpace(block.ExpandedContext) != "" {
		item["expanded_context"] = block.ExpandedContext
	}
	if strings.TrimSpace(block.LeadRoleID) != "" {
		item["lead_role_id"] = block.LeadRoleID
	}
	if block.Result != nil {
		item["result"] = block.Result
	}
	return item
}

func (s *scenarioService) ExpandContext(ctx context.Context, claims entity.AccessClaims, scenarioID string, options entity.AIGenerationOptions) (string, error) {
	scenario, doc, actions, err := s.aiContext(ctx, claims, scenarioID, "")
	if err != nil {
		return "", err
	}
	if missing := missingScenarioAIFields(doc, false); len(missing) > 0 {
		return "", fmt.Errorf("%w: %s", entity.ErrAIGenerationContextIncomplete, strings.Join(missing, ", "))
	}
	var output struct {
		ExpandedContext string `json:"expanded_context"`
	}
	input := map[string]any{"task": "expand_context", "scenario": scenarioAIInput(doc), "roles": doc.Roles, "semantic_actions": actions, "instructions": options.Instructions}
	if err := s.ai.GenerateJSON(ctx, ai.RoleScenarioArchitect, input, &output, func() error {
		if strings.TrimSpace(output.ExpandedContext) == "" {
			return fmt.Errorf("expanded_context is empty")
		}
		return nil
	}); err != nil {
		return "", err
	}
	_ = scenario
	return strings.TrimSpace(output.ExpandedContext), nil
}

func (s *scenarioService) GenerateDraft(ctx context.Context, claims entity.AccessClaims, scenarioID string, options entity.AIGenerationOptions) (*entity.ScenarioDraftProposal, *entity.ScenarioValidation, error) {
	scenario, doc, actions, err := s.aiContext(ctx, claims, scenarioID, "")
	if err != nil {
		return nil, nil, err
	}
	if missing := missingScenarioAIFields(doc, true); len(missing) > 0 {
		return nil, nil, fmt.Errorf("%w: %s", entity.ErrAIGenerationContextIncomplete, strings.Join(missing, ", "))
	}
	var output architectDraftOutput
	input := map[string]any{"task": "generate_draft", "scenario": scenarioAIInput(doc), "roles": doc.Roles, "semantic_actions": actions, "instructions": options.Instructions}
	if err := s.ai.GenerateJSON(ctx, ai.RoleScenarioArchitect, input, &output, func() error {
		return validateArchitectDraft(output, doc, actions)
	}); err != nil {
		return nil, nil, err
	}
	blocks, err := normalizeArchitectBlocks(output.Blocks)
	if err != nil {
		return nil, nil, err
	}
	entity.NormalizeScenarioBlocks(blocks)
	proposal := &entity.ScenarioDraftProposal{ExpandedContext: strings.TrimSpace(output.ExpandedContext), Blocks: blocks}
	copyScenario := *scenario
	copyScenario.Context = entity.ScenarioContextJSON(doc.Description, proposal.ExpandedContext, doc.Tags)
	nodes, _ := json.Marshal(proposal.Blocks)
	copyScenario.Nodes = nodes
	validation, err := validateScenario(&copyScenario)
	if err != nil {
		return nil, nil, err
	}
	return proposal, validation, nil
}

func (s *scenarioService) GenerateBlock(ctx context.Context, claims entity.AccessClaims, scenarioID, blockID string, options entity.AIGenerationOptions) (*entity.BlockContentProposal, *entity.ScenarioValidation, error) {
	scenario, doc, actions, err := s.aiContext(ctx, claims, scenarioID, blockID)
	if err != nil {
		return nil, nil, err
	}
	block, index, ok := findBlock(doc.Blocks, blockID)
	if !ok {
		return nil, nil, entity.ErrBlockNotFound
	}
	var output struct {
		Block blockContentOutput `json:"block"`
	}
	input := map[string]any{"task": "generate_block", "difficulty": doc.Difficulty, "scenario": scenarioAIInput(doc), "roles": doc.Roles, "block": block, "neighbors": blockNeighbors(doc.Blocks, blockID), "semantic_actions": actions, "instructions": options.Instructions}
	if err := s.ai.GenerateJSON(ctx, ai.RoleBlockDesigner, input, &output, func() error {
		proposal := output.Block.toProposal()
		if block.Type == "end" && (proposal.Result == nil || strings.TrimSpace(proposal.Result.Summary) == "") {
			return fmt.Errorf("end block result is required")
		}
		return validateBlockProposal(proposal, doc, actions)
	}); err != nil {
		return nil, nil, err
	}
	proposal := output.Block.toProposal()
	blocks := append([]entity.ScenarioBlock(nil), doc.Blocks...)
	blocks[index].Name = proposal.Name
	blocks[index].Description = proposal.Description
	blocks[index].ExpandedContext = proposal.ExpandedContext
	blocks[index].LeadRoleID = proposal.LeadRoleID
	blocks[index].RoleGoals = proposal.RoleGoals
	blocks[index].RoleBehaviors = proposal.RoleBehaviors
	blocks[index].Result = proposal.Result
	copyScenario := *scenario
	nodes, _ := json.Marshal(blocks)
	copyScenario.Nodes = nodes
	validation, err := validateScenario(&copyScenario)
	if err != nil {
		return nil, nil, err
	}
	return &proposal, validation, nil
}

func (s *scenarioService) GenerateDialogue(ctx context.Context, claims entity.AccessClaims, scenarioID, blockID, learnerRoleID string, options entity.AIGenerationOptions) (*entity.DeterministicDialogue, *entity.ScenarioValidation, error) {
	scenario, doc, actions, err := s.aiContext(ctx, claims, scenarioID, blockID)
	if err != nil {
		return nil, nil, err
	}
	block, index, ok := findBlock(doc.Blocks, blockID)
	if !ok {
		return nil, nil, entity.ErrBlockNotFound
	}
	if _, ok := roleIDSet(doc)[learnerRoleID]; !ok {
		return nil, nil, entity.ErrAIGenerationContextIncomplete
	}
	var output dialogueOutput
	input := map[string]any{"task": "generate_dialogue", "difficulty": doc.Difficulty, "scenario": scenarioAIInput(doc), "roles": doc.Roles, "block": block, "transitions": block.Transitions, "learner_role_id": learnerRoleID, "semantic_actions": actions, "instructions": options.Instructions}
	if err := s.ai.GenerateJSON(ctx, ai.RoleDialogueBuilder, input, &output, func() error {
		return validateDialogueOutput(output.effectiveDialogue(), learnerRoleID, block, actions)
	}); err != nil {
		return nil, nil, err
	}
	dialogue, err := normalizeDialogue(output.effectiveDialogue())
	if err != nil {
		return nil, nil, err
	}
	blocks := append([]entity.ScenarioBlock(nil), doc.Blocks...)
	blocks[index].DeterministicDialogue = &dialogue
	copyScenario := *scenario
	nodes, _ := json.Marshal(blocks)
	copyScenario.Nodes = nodes
	validation, err := validateScenario(&copyScenario)
	if err != nil {
		return nil, nil, err
	}
	return &dialogue, validation, nil
}

func (s *scenarioService) aiContext(ctx context.Context, claims entity.AccessClaims, scenarioID, blockID string) (*entity.Scenario, entity.ScenarioDocument, []entity.SemanticActionDefinition, error) {
	if err := requireMethodist(claims); err != nil {
		return nil, entity.ScenarioDocument{}, nil, err
	}
	scenario, err := s.scenarioRepository.GetScenarioByID(ctx, scenarioID)
	if err != nil {
		return nil, entity.ScenarioDocument{}, nil, err
	}
	if err := ensureOwnedBy(claims.UserID, scenario); err != nil {
		return nil, entity.ScenarioDocument{}, nil, err
	}
	if scenario.Status != entity.ScenarioStatusDraft {
		return nil, entity.ScenarioDocument{}, nil, entity.ErrScenarioNotDraft
	}
	doc := entity.ScenarioDocumentFromScenario(*scenario)
	if blockID != "" {
		if _, _, ok := findBlock(doc.Blocks, blockID); !ok {
			return nil, entity.ScenarioDocument{}, nil, entity.ErrBlockNotFound
		}
	}
	actions, err := s.ListSemanticActions(ctx)
	if err != nil {
		return nil, entity.ScenarioDocument{}, nil, err
	}
	return scenario, doc, actions, nil
}

func scenarioAIInput(doc entity.ScenarioDocument) map[string]any {
	return map[string]any{
		"scenario_id":        doc.ScenarioID,
		"title":              doc.Title,
		"description":        doc.Description,
		"expanded_context":   doc.ExpandedContext,
		"difficulty":         doc.Difficulty,
		"tags":               doc.Tags,
		"supported_runtimes": doc.SupportedRuntimes,
		"supported_channels": doc.SupportedChannels,
		"target_cluster_ids": doc.TargetClusterIDs,
	}
}

func missingScenarioAIFields(doc entity.ScenarioDocument, requireRoles bool) []string {
	missing := []string{}
	if strings.TrimSpace(doc.Title) == "" {
		missing = append(missing, "title")
	}
	if strings.TrimSpace(doc.Description) == "" {
		missing = append(missing, "description")
	}
	if requireRoles && len(doc.Roles) != 2 {
		missing = append(missing, "roles")
	}
	return missing
}

type architectDraftOutput struct {
	ExpandedContext string                 `json:"expanded_context"`
	Blocks          []architectBlockOutput `json:"blocks"`
}

type architectBlockOutput struct {
	Key             string                      `json:"key"`
	Type            string                      `json:"type"`
	Name            string                      `json:"name"`
	Description     string                      `json:"description"`
	ExpandedContext string                      `json:"expanded_context"`
	LeadRoleID      string                      `json:"lead_role_id"`
	RoleGoals       []entity.BlockRoleGoal      `json:"role_goals"`
	RoleBehaviors   []entity.BlockRoleBehavior  `json:"role_behaviors"`
	Result          *entity.ScenarioEndResult   `json:"result"`
	Transitions     []architectTransitionOutput `json:"transitions"`
}

type blockContentOutput struct {
	Name            string                     `json:"name"`
	Description     string                     `json:"description"`
	ExpandedContext string                     `json:"expanded_context"`
	LeadRoleID      string                     `json:"lead_role_id"`
	RoleGoals       []entity.BlockRoleGoal     `json:"role_goals"`
	RoleBehaviors   []entity.BlockRoleBehavior `json:"role_behaviors"`
	Result          *entity.ScenarioEndResult  `json:"result"`
}

func (b blockContentOutput) toProposal() entity.BlockContentProposal {
	return entity.BlockContentProposal{
		Name:            b.Name,
		Description:     b.Description,
		ExpandedContext: b.ExpandedContext,
		LeadRoleID:      b.LeadRoleID,
		RoleGoals:       b.RoleGoals,
		RoleBehaviors:   b.RoleBehaviors,
		Result:          b.Result,
	}
}

type architectTransitionOutput struct {
	TargetBlockKey string `json:"target_block_key"`
	Condition      string `json:"condition"`
	Note           string `json:"note"`
	Weight         int    `json:"weight"`
	Forbidden      bool   `json:"forbidden"`
}

type dialogueOutput struct {
	DeterministicDialogue localDialogue `json:"deterministic_dialogue"`
}

type localDialogue struct {
	LearnerRoleID string            `json:"learner_role_id"`
	StartNodeKey  string            `json:"start_node_key"`
	Nodes         []localDialogNode `json:"nodes"`
}

type localDialogNode struct {
	Key               string                                 `json:"key"`
	IncomingMessage   *entity.DeterministicMessage           `json:"incoming_message"`
	AssessmentTargets []entity.DeterministicAssessmentTarget `json:"assessment_targets"`
	ReplyOptions      []localReplyOption                     `json:"reply_options"`
}

type localReplyOption struct {
	Key                string                                `json:"key"`
	RoleID             string                                `json:"role_id"`
	Text               string                                `json:"text"`
	SemanticActionCode string                                `json:"semantic_action_code"`
	Assessments        []entity.DeterministicReplyAssessment `json:"assessments"`
	NextNodeKey        string                                `json:"next_node_key"`
	TransitionID       string                                `json:"transition_id"`
}

func validateArchitectDraft(output architectDraftOutput, doc entity.ScenarioDocument, _ []entity.SemanticActionDefinition) error {
	if strings.TrimSpace(output.ExpandedContext) == "" || len(output.Blocks) == 0 {
		return fmt.Errorf("empty draft")
	}
	roleIDs := roleIDSet(doc)
	keys := map[string]bool{}
	starts, ends := 0, 0
	for _, block := range output.Blocks {
		if block.Key == "" || keys[block.Key] {
			return fmt.Errorf("duplicate or empty block key")
		}
		keys[block.Key] = true
		if block.Type == "start" {
			starts++
		}
		if block.Type == "end" {
			ends++
			if block.Result == nil || strings.TrimSpace(block.Result.Summary) == "" {
				return fmt.Errorf("end block result is required")
			}
		}
		if block.LeadRoleID != "" && !roleIDs[block.LeadRoleID] {
			return fmt.Errorf("unknown lead role")
		}
		seenBehaviors := map[string]bool{}
		for _, behavior := range block.RoleBehaviors {
			if !roleIDs[behavior.RoleID] {
				return fmt.Errorf("unknown behavior role")
			}
			if seenBehaviors[behavior.RoleID] {
				return fmt.Errorf("duplicate behavior role")
			}
			seenBehaviors[behavior.RoleID] = true
		}
	}
	if starts != 1 || ends < 1 {
		return fmt.Errorf("draft must contain one start and at least one end")
	}
	for _, block := range output.Blocks {
		for _, transition := range block.Transitions {
			if !keys[transition.TargetBlockKey] {
				return fmt.Errorf("unknown target block key")
			}
		}
	}
	return nil
}

func normalizeArchitectBlocks(input []architectBlockOutput) ([]entity.ScenarioBlock, error) {
	keyToID := make(map[string]string, len(input))
	for _, block := range input {
		id, err := entity.NewUUID()
		if err != nil {
			return nil, err
		}
		keyToID[block.Key] = id
	}
	blocks := make([]entity.ScenarioBlock, 0, len(input))
	for _, block := range input {
		transitions := make([]entity.ScenarioTransition, 0, len(block.Transitions))
		for _, transition := range block.Transitions {
			id, err := entity.NewUUID()
			if err != nil {
				return nil, err
			}
			transitions = append(transitions, entity.ScenarioTransition{ID: id, TargetBlockID: keyToID[transition.TargetBlockKey], Condition: transition.Condition, Note: transition.Note, Weight: transition.Weight, Forbidden: transition.Forbidden})
		}
		blocks = append(blocks, entity.ScenarioBlock{
			ID:              keyToID[block.Key],
			Type:            block.Type,
			Name:            block.Name,
			Description:     block.Description,
			ExpandedContext: block.ExpandedContext,
			LeadRoleID:      block.LeadRoleID,
			RoleGoals:       block.RoleGoals,
			RoleBehaviors:   block.RoleBehaviors,
			Result:          block.Result,
			Transitions:     transitions,
		})
	}
	return blocks, nil
}

func validateBlockProposal(block entity.BlockContentProposal, doc entity.ScenarioDocument, _ []entity.SemanticActionDefinition) error {
	if strings.TrimSpace(block.Name) == "" {
		return fmt.Errorf("name is empty")
	}
	if block.LeadRoleID != "" && !roleIDSet(doc)[block.LeadRoleID] {
		return fmt.Errorf("unknown lead role")
	}
	roles := roleIDSet(doc)
	seenBehaviors := map[string]bool{}
	for _, behavior := range block.RoleBehaviors {
		if !roles[behavior.RoleID] {
			return fmt.Errorf("unknown behavior role")
		}
		if seenBehaviors[behavior.RoleID] {
			return fmt.Errorf("duplicate behavior role")
		}
		seenBehaviors[behavior.RoleID] = true
	}
	return nil
}

func (output dialogueOutput) effectiveDialogue() localDialogue {
	return output.DeterministicDialogue
}

func validateDialogueOutput(output localDialogue, learnerRoleID string, block entity.ScenarioBlock, actions []entity.SemanticActionDefinition) error {
	if strings.TrimSpace(output.LearnerRoleID) != learnerRoleID {
		return fmt.Errorf("learner_role_id mismatch")
	}
	startKey := strings.TrimSpace(output.StartNodeKey)
	if startKey == "" || len(output.Nodes) == 0 {
		return fmt.Errorf("empty dialogue")
	}
	nodeKeys := map[string]bool{}
	codes := actionCodeSet(actions)
	transitions := map[string]bool{}
	for _, transition := range block.Transitions {
		if !transition.Forbidden {
			transitions[transition.ID] = true
		}
	}
	for _, node := range output.Nodes {
		nodeKey := strings.TrimSpace(node.Key)
		if nodeKey == "" || nodeKeys[nodeKey] {
			return fmt.Errorf("duplicate or empty node key")
		}
		nodeKeys[nodeKey] = true
		if node.IncomingMessage != nil && !codes[node.IncomingMessage.SemanticActionCode] {
			return fmt.Errorf("unknown incoming semantic action")
		}
		targets := map[string]bool{}
		for _, target := range node.AssessmentTargets {
			if strings.TrimSpace(target.ClusterID) == "" {
				return fmt.Errorf("empty assessment target cluster")
			}
			if targets[target.ClusterID] {
				return fmt.Errorf("duplicate assessment target")
			}
			targets[target.ClusterID] = true
			if target.Weight != 0.5 && target.Weight != 1 && target.Weight != 1.5 {
				return fmt.Errorf("invalid assessment target weight")
			}
		}
		for _, option := range node.ReplyOptions {
			if option.RoleID != learnerRoleID {
				return fmt.Errorf("reply option role mismatch")
			}
			if !codes[option.SemanticActionCode] {
				return fmt.Errorf("unknown reply semantic action")
			}
			assessments := map[string]bool{}
			for _, assessment := range option.Assessments {
				if !targets[assessment.ClusterID] {
					return fmt.Errorf("reply assessment target mismatch")
				}
				if assessments[assessment.ClusterID] {
					return fmt.Errorf("duplicate reply assessment")
				}
				assessments[assessment.ClusterID] = true
				if assessment.Level < 0 || assessment.Level > 4 {
					return fmt.Errorf("invalid reply assessment level")
				}
			}
			for clusterID := range targets {
				if !assessments[clusterID] {
					return fmt.Errorf("missing reply assessment")
				}
			}
			nextNodeKey := strings.TrimSpace(option.NextNodeKey)
			if (nextNodeKey == "") == (option.TransitionID == "") {
				return fmt.Errorf("reply must have exactly one next_node_key or transition_id")
			}
			if option.TransitionID != "" && !transitions[option.TransitionID] {
				return fmt.Errorf("unknown transition")
			}
		}
	}
	if !nodeKeys[startKey] {
		return fmt.Errorf("start node not found")
	}
	for _, node := range output.Nodes {
		for _, option := range node.ReplyOptions {
			nextNodeKey := strings.TrimSpace(option.NextNodeKey)
			if nextNodeKey != "" && !nodeKeys[nextNodeKey] {
				return fmt.Errorf("next node not found")
			}
		}
	}
	return nil
}

func normalizeDialogue(input localDialogue) (entity.DeterministicDialogue, error) {
	keyToID := map[string]string{}
	for _, node := range input.Nodes {
		nodeKey := strings.TrimSpace(node.Key)
		id, err := entity.NewUUID()
		if err != nil {
			return entity.DeterministicDialogue{}, err
		}
		keyToID[nodeKey] = id
	}
	nodes := make([]entity.DeterministicDialogueNode, 0, len(input.Nodes))
	for _, node := range input.Nodes {
		nodeKey := strings.TrimSpace(node.Key)
		options := make([]entity.DeterministicReplyOption, 0, len(node.ReplyOptions))
		for _, option := range node.ReplyOptions {
			id, err := entity.NewUUID()
			if err != nil {
				return entity.DeterministicDialogue{}, err
			}
			options = append(options, entity.DeterministicReplyOption{
				ID:                 id,
				RoleID:             option.RoleID,
				Text:               option.Text,
				SemanticActionCode: option.SemanticActionCode,
				Assessments:        option.Assessments,
				NextNodeID:         keyToID[strings.TrimSpace(option.NextNodeKey)],
				TransitionID:       option.TransitionID,
			})
		}
		nodes = append(nodes, entity.DeterministicDialogueNode{ID: keyToID[nodeKey], IncomingMessage: node.IncomingMessage, AssessmentTargets: node.AssessmentTargets, ReplyOptions: options})
	}
	return entity.DeterministicDialogue{LearnerRoleID: strings.TrimSpace(input.LearnerRoleID), StartNodeID: keyToID[strings.TrimSpace(input.StartNodeKey)], Nodes: nodes}, nil
}

func findBlock(blocks []entity.ScenarioBlock, id string) (entity.ScenarioBlock, int, bool) {
	for i, block := range blocks {
		if block.ID == id {
			return block, i, true
		}
	}
	return entity.ScenarioBlock{}, 0, false
}

func blockNeighbors(blocks []entity.ScenarioBlock, id string) map[string]any {
	incoming := []entity.ScenarioBlock{}
	outgoing := []entity.ScenarioBlock{}
	targets := map[string]bool{}
	for _, block := range blocks {
		if block.ID == id {
			for _, transition := range block.Transitions {
				targets[transition.TargetBlockID] = true
			}
			continue
		}
		for _, transition := range block.Transitions {
			if transition.TargetBlockID == id {
				incoming = append(incoming, block)
			}
		}
	}
	for _, block := range blocks {
		if targets[block.ID] {
			outgoing = append(outgoing, block)
		}
	}
	return map[string]any{"incoming_blocks": incoming, "outgoing_blocks": outgoing}
}

func roleIDSet(doc entity.ScenarioDocument) map[string]bool {
	items := map[string]bool{}
	for _, role := range doc.Roles {
		items[role.ID] = true
	}
	return items
}

func actionCodeSet(actions []entity.SemanticActionDefinition) map[string]bool {
	items := map[string]bool{}
	for _, action := range actions {
		items[action.Code] = true
	}
	return items
}
