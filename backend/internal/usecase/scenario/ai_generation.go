package scenario

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/ai"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
)

type aiGenerator interface {
	GenerateJSON(ctx context.Context, role ai.Role, input any, output any, validate func() error) error
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
		Block entity.BlockContentProposal `json:"block"`
	}
	input := map[string]any{"scenario": scenarioAIInput(doc), "roles": doc.Roles, "block": block, "neighbors": blockNeighbors(doc.Blocks, blockID), "semantic_actions": actions, "instructions": options.Instructions}
	if err := s.ai.GenerateJSON(ctx, ai.RoleBlockDesigner, input, &output, func() error {
		return validateBlockProposal(output.Block, doc, actions)
	}); err != nil {
		return nil, nil, err
	}
	blocks := append([]entity.ScenarioBlock(nil), doc.Blocks...)
	blocks[index].Name = output.Block.Name
	blocks[index].Description = output.Block.Description
	blocks[index].ExpandedContext = output.Block.ExpandedContext
	blocks[index].LeadRoleID = output.Block.LeadRoleID
	blocks[index].LeadProbability = output.Block.LeadProbability
	blocks[index].RoleGoals = output.Block.RoleGoals
	blocks[index].Behavior = output.Block.Behavior
	blocks[index].ActionRules = output.Block.ActionRules
	blocks[index].Constraints = output.Block.Constraints
	copyScenario := *scenario
	nodes, _ := json.Marshal(blocks)
	copyScenario.Nodes = nodes
	validation, err := validateScenario(&copyScenario)
	if err != nil {
		return nil, nil, err
	}
	return &output.Block, validation, nil
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
	input := map[string]any{"scenario": scenarioAIInput(doc), "roles": doc.Roles, "block": block, "transitions": block.Transitions, "learner_role_id": learnerRoleID, "semantic_actions": actions, "instructions": options.Instructions}
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
	blocks[index].DeterministicDialogue = dialogue
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
	LeadProbability int                         `json:"lead_probability"`
	RoleGoals       []entity.BlockRoleGoal      `json:"role_goals"`
	Behavior        string                      `json:"behavior"`
	ActionRules     []entity.SemanticActionRule `json:"action_rules"`
	Constraints     []string                    `json:"constraints"`
	Transitions     []architectTransitionOutput `json:"transitions"`
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
	StartNodeKey string            `json:"start_node_key"`
	Nodes        []localDialogNode `json:"nodes"`
}

type localDialogNode struct {
	Key             string                       `json:"key"`
	IncomingMessage *entity.DeterministicMessage `json:"incoming_message"`
	ReplyOptions    []localReplyOption           `json:"reply_options"`
}

type localReplyOption struct {
	Key                string `json:"key"`
	RoleID             string `json:"role_id"`
	Text               string `json:"text"`
	SemanticActionCode string `json:"semantic_action_code"`
	NextNodeKey        string `json:"next_node_key"`
	TransitionID       string `json:"transition_id"`
}

func validateArchitectDraft(output architectDraftOutput, doc entity.ScenarioDocument, actions []entity.SemanticActionDefinition) error {
	if strings.TrimSpace(output.ExpandedContext) == "" || len(output.Blocks) == 0 {
		return fmt.Errorf("empty draft")
	}
	roleIDs := roleIDSet(doc)
	actionCodes := actionCodeSet(actions)
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
		}
		if block.LeadRoleID != "" && !roleIDs[block.LeadRoleID] {
			return fmt.Errorf("unknown lead role")
		}
		for _, rule := range block.ActionRules {
			if !actionCodes[rule.SemanticActionCode] {
				return fmt.Errorf("unknown semantic action")
			}
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
			LeadProbability: block.LeadProbability,
			RoleGoals:       block.RoleGoals,
			Behavior:        block.Behavior,
			ActionRules:     block.ActionRules,
			Constraints:     block.Constraints,
			Transitions:     transitions,
		})
	}
	return blocks, nil
}

func validateBlockProposal(block entity.BlockContentProposal, doc entity.ScenarioDocument, actions []entity.SemanticActionDefinition) error {
	if strings.TrimSpace(block.Name) == "" {
		return fmt.Errorf("name is empty")
	}
	if block.LeadRoleID != "" && !roleIDSet(doc)[block.LeadRoleID] {
		return fmt.Errorf("unknown lead role")
	}
	codes := actionCodeSet(actions)
	for _, rule := range block.ActionRules {
		if !codes[rule.SemanticActionCode] {
			return fmt.Errorf("unknown semantic action")
		}
	}
	return nil
}

func (output dialogueOutput) effectiveDialogue() localDialogue {
	return output.DeterministicDialogue
}

func validateDialogueOutput(output localDialogue, learnerRoleID string, block entity.ScenarioBlock, actions []entity.SemanticActionDefinition) error {
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
		for _, option := range node.ReplyOptions {
			if option.RoleID != learnerRoleID {
				return fmt.Errorf("reply option role mismatch")
			}
			if !codes[option.SemanticActionCode] {
				return fmt.Errorf("unknown reply semantic action")
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
				NextNodeID:         keyToID[strings.TrimSpace(option.NextNodeKey)],
				TransitionID:       option.TransitionID,
			})
		}
		nodes = append(nodes, entity.DeterministicDialogueNode{ID: keyToID[nodeKey], IncomingMessage: node.IncomingMessage, ReplyOptions: options})
	}
	return entity.DeterministicDialogue{StartNodeID: keyToID[strings.TrimSpace(input.StartNodeKey)], Nodes: nodes}, nil
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
