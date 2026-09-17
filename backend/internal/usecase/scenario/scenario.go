package scenario

import (
	"context"
	"encoding/json"
	"strings"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
)

type scenarioRepository interface {
	CreateScenario(ctx context.Context, scenario entity.Scenario) (*entity.Scenario, error)
	GetScenarioByID(ctx context.Context, id string) (*entity.Scenario, error)
	ListScenariosByOrganization(ctx context.Context, organizationID string, status *entity.ScenarioStatus, limit, offset int32) ([]entity.Scenario, error)
	UpdateScenario(ctx context.Context, scenario entity.Scenario) (*entity.Scenario, error)
	SaveScenarioValidation(ctx context.Context, id string, result []byte) (*entity.Scenario, error)
	SetScenarioStatus(ctx context.Context, id string, status entity.ScenarioStatus) (*entity.Scenario, error)
	CreateScenarioVersion(ctx context.Context, scenarioID, version string, snapshot []byte) (*entity.ScenarioVersion, error)
	ListScenarioVersions(ctx context.Context, scenarioID string, limit, offset int32) ([]entity.ScenarioVersion, error)
	GetScenarioVersion(ctx context.Context, scenarioID, version string) (*entity.ScenarioVersion, error)
}

type userRepository interface {
	GetUserByID(ctx context.Context, id string) (*entity.User, error)
}

type transactor interface {
	WithTx(ctx context.Context, f func(ctx context.Context) error) error
}

type scenarioService struct {
	scenarioRepository scenarioRepository
	userRepository     userRepository
	transactor         transactor
}

func NewScenarioService(scenarioRepository scenarioRepository, userRepository userRepository, transactor transactor) *scenarioService {
	return &scenarioService{scenarioRepository: scenarioRepository, userRepository: userRepository, transactor: transactor}
}

func (s *scenarioService) ListScenarios(ctx context.Context, claims entity.AccessClaims, params entity.ListScenarios) ([]entity.Scenario, error) {
	user, err := s.userRepository.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if user.OrganizationID == nil {
		return []entity.Scenario{}, nil
	}
	return s.scenarioRepository.ListScenariosByOrganization(ctx, *user.OrganizationID, params.Status, params.Limit, params.Offset)
}

func (s *scenarioService) CreateScenario(ctx context.Context, claims entity.AccessClaims, params entity.CreateScenario) (*entity.Scenario, error) {
	if err := requireMethodist(claims); err != nil {
		return nil, err
	}
	user, err := s.userRepository.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if user.OrganizationID == nil {
		return nil, entity.ErrOrganizationRequired
	}
	id, err := entity.NewUUID()
	if err != nil {
		return nil, err
	}
	return s.scenarioRepository.CreateScenario(ctx, entity.Scenario{
		ID:                   id,
		OrganizationID:       user.OrganizationID,
		CreatedBy:            &user.ID,
		Version:              params.Version,
		Title:                params.Title,
		Status:               entity.ScenarioStatusDraft,
		ModeSupport:          params.ModeSupport,
		Context:              params.Context,
		Roles:                params.Roles,
		Goals:                params.Goals,
		Interests:            params.Interests,
		Constraints:          params.Constraints,
		Parameters:           params.Parameters,
		PrivateInformation:   params.PrivateInformation,
		Difficulty:           params.Difficulty,
		Competencies:         params.Competencies,
		ObservationRules:     params.ObservationRules,
		CompletionConditions: params.CompletionConditions,
		Channels:             params.Channels,
		Nodes:                params.Nodes,
	})
}

func (s *scenarioService) GenerateScenario(context.Context, entity.AccessClaims, string) (*entity.Scenario, error) {
	panic("implements me")
}

func (s *scenarioService) GetScenario(ctx context.Context, claims entity.AccessClaims, scenarioID string) (*entity.Scenario, error) {
	scenario, err := s.scenarioRepository.GetScenarioByID(ctx, scenarioID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureReadable(ctx, claims, scenario); err != nil {
		return nil, err
	}
	return scenario, nil
}

func (s *scenarioService) UpdateScenario(ctx context.Context, claims entity.AccessClaims, params entity.UpdateScenario) error {
	if err := requireMethodist(claims); err != nil {
		return err
	}
	scenario, err := s.scenarioRepository.GetScenarioByID(ctx, params.ScenarioID)
	if err != nil {
		return err
	}
	if err := ensureOwnedBy(claims.UserID, scenario); err != nil {
		return err
	}
	if scenario.Status != entity.ScenarioStatusDraft {
		return entity.ErrInvalidScenarioStatus
	}
	applyUpdate(scenario, params)
	_, err = s.scenarioRepository.UpdateScenario(ctx, *scenario)
	return err
}

func (s *scenarioService) ValidateScenario(ctx context.Context, claims entity.AccessClaims, scenarioID string) (*entity.ScenarioValidation, error) {
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
	validation, err := validateScenario(scenario)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(validation)
	if err != nil {
		return nil, err
	}
	if _, err := s.scenarioRepository.SaveScenarioValidation(ctx, scenario.ID, data); err != nil {
		return nil, err
	}
	return validation, nil
}

func (s *scenarioService) PublishScenario(ctx context.Context, claims entity.AccessClaims, scenarioID string) (*entity.ScenarioVersion, error) {
	if err := requireMethodist(claims); err != nil {
		return nil, err
	}
	var version *entity.ScenarioVersion
	err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		scenario, err := s.scenarioRepository.GetScenarioByID(txCtx, scenarioID)
		if err != nil {
			return err
		}
		if err := ensureOwnedBy(claims.UserID, scenario); err != nil {
			return err
		}
		if scenario.ValidatedRevision == nil || *scenario.ValidatedRevision != scenario.Revision {
			return entity.ErrScenarioNotValidated
		}
		if !validationResultIsValid(scenario.ValidationResult) {
			return entity.ErrScenarioValidationFailed
		}
		snapshot, err := marshalScenarioSnapshot(scenario)
		if err != nil {
			return err
		}
		version, err = s.scenarioRepository.CreateScenarioVersion(txCtx, scenario.ID, scenario.Version, snapshot)
		if err != nil {
			return err
		}
		_, err = s.scenarioRepository.SetScenarioStatus(txCtx, scenario.ID, entity.ScenarioStatusPublished)
		return err
	})
	if err != nil {
		return nil, err
	}
	return version, nil
}

func (s *scenarioService) ArchiveScenario(ctx context.Context, claims entity.AccessClaims, scenarioID string) error {
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
	if scenario.Status == entity.ScenarioStatusArchived {
		return entity.ErrInvalidScenarioStatus
	}
	_, err = s.scenarioRepository.SetScenarioStatus(ctx, scenario.ID, entity.ScenarioStatusArchived)
	return err
}

func (s *scenarioService) ListScenarioVersions(ctx context.Context, claims entity.AccessClaims, scenarioID string, limit, offset int32) ([]entity.ScenarioVersion, error) {
	scenario, err := s.GetScenario(ctx, claims, scenarioID)
	if err != nil {
		return nil, err
	}
	return s.scenarioRepository.ListScenarioVersions(ctx, scenario.ID, limit, offset)
}

func (s *scenarioService) GetScenarioVersion(ctx context.Context, claims entity.AccessClaims, scenarioID, version string) (*entity.ScenarioVersion, error) {
	scenario, err := s.GetScenario(ctx, claims, scenarioID)
	if err != nil {
		return nil, err
	}
	return s.scenarioRepository.GetScenarioVersion(ctx, scenario.ID, version)
}

func requireMethodist(claims entity.AccessClaims) error {
	if claims.ActiveRole != entity.RoleMethodist {
		return entity.ErrMethodistRoleRequired
	}
	return nil
}

func ensureOwnedBy(userID string, scenario *entity.Scenario) error {
	if scenario.CreatedBy == nil || *scenario.CreatedBy != userID {
		return entity.ErrScenarioNotOwned
	}
	return nil
}

func (s *scenarioService) ensureReadable(ctx context.Context, claims entity.AccessClaims, scenario *entity.Scenario) error {
	user, err := s.userRepository.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return err
	}
	if user.OrganizationID != nil && scenario.OrganizationID != nil && *user.OrganizationID == *scenario.OrganizationID {
		return nil
	}
	if scenario.CreatedBy != nil && *scenario.CreatedBy == claims.UserID {
		return nil
	}
	return entity.ErrScenarioNotFound
}

func applyUpdate(scenario *entity.Scenario, update entity.UpdateScenario) {
	if update.Version != nil {
		scenario.Version = *update.Version
	}
	if update.Title != nil {
		scenario.Title = *update.Title
	}
	if update.ModeSupport != nil {
		scenario.ModeSupport = *update.ModeSupport
	}
	if update.Context != nil {
		scenario.Context = *update.Context
	}
	if update.Roles != nil {
		scenario.Roles = *update.Roles
	}
	if update.Goals != nil {
		scenario.Goals = *update.Goals
	}
	if update.Interests != nil {
		scenario.Interests = *update.Interests
	}
	if update.Constraints != nil {
		scenario.Constraints = *update.Constraints
	}
	if update.Parameters != nil {
		scenario.Parameters = *update.Parameters
	}
	if update.PrivateInformation != nil {
		scenario.PrivateInformation = *update.PrivateInformation
	}
	if update.Difficulty != nil {
		scenario.Difficulty = *update.Difficulty
	}
	if update.Competencies != nil {
		scenario.Competencies = *update.Competencies
	}
	if update.ObservationRules != nil {
		scenario.ObservationRules = *update.ObservationRules
	}
	if update.CompletionConditions != nil {
		scenario.CompletionConditions = *update.CompletionConditions
	}
	if update.Channels != nil {
		scenario.Channels = *update.Channels
	}
	if update.Nodes != nil {
		scenario.Nodes = *update.Nodes
	}
}

func validationResultIsValid(raw []byte) bool {
	var validation entity.ScenarioValidation
	return len(raw) > 0 && json.Unmarshal(raw, &validation) == nil && validation.Valid
}

func marshalScenarioSnapshot(s *entity.Scenario) ([]byte, error) {
	return json.Marshal(scenarioPayload{
		ScenarioID:           s.ID,
		Version:              s.Version,
		Title:                s.Title,
		ModeSupport:          s.ModeSupport,
		Context:              rawJSON(s.Context),
		Roles:                rawJSON(s.Roles),
		Goals:                rawJSON(s.Goals),
		Interests:            rawJSON(s.Interests),
		Constraints:          rawJSON(s.Constraints),
		Parameters:           rawJSON(s.Parameters),
		PrivateInformation:   rawJSON(s.PrivateInformation),
		Difficulty:           s.Difficulty,
		Competencies:         s.Competencies,
		ObservationRules:     rawJSON(s.ObservationRules),
		CompletionConditions: rawJSON(s.CompletionConditions),
		Channels:             s.Channels,
		Nodes:                rawJSON(s.Nodes),
	})
}

type scenarioPayload struct {
	ScenarioID           string                    `json:"scenario_id"`
	Version              string                    `json:"version"`
	Title                string                    `json:"title"`
	ModeSupport          []string                  `json:"mode_support"`
	Context              json.RawMessage           `json:"context"`
	Roles                json.RawMessage           `json:"roles"`
	Goals                json.RawMessage           `json:"goals"`
	Interests            json.RawMessage           `json:"interests"`
	Constraints          json.RawMessage           `json:"constraints"`
	Parameters           json.RawMessage           `json:"parameters"`
	PrivateInformation   json.RawMessage           `json:"private_information"`
	Difficulty           entity.ScenarioDifficulty `json:"difficulty"`
	Competencies         []string                  `json:"competencies"`
	ObservationRules     json.RawMessage           `json:"observation_rules"`
	CompletionConditions json.RawMessage           `json:"completion_conditions"`
	Channels             []string                  `json:"channels"`
	Nodes                json.RawMessage           `json:"nodes"`
}

func rawJSON(value json.RawMessage) json.RawMessage {
	if len(value) == 0 {
		return json.RawMessage("null")
	}
	return value
}

func validateScenario(s *entity.Scenario) (*entity.ScenarioValidation, error) {
	var goals map[string]string
	var contextBlock struct {
		ConflictType []string `json:"conflict_type"`
		Description  string   `json:"description"`
	}
	var roles []struct {
		ID string `json:"id"`
	}
	var completionConditions []map[string]any
	var nodes []validationNode
	var rules []struct {
		SemanticAction string `json:"semantic_action"`
		Indicator      string `json:"indicator"`
	}
	if err := unmarshalScenarioJSON(s.Goals, &goals); err != nil {
		return nil, err
	}
	if err := unmarshalScenarioJSON(s.Context, &contextBlock); err != nil {
		return nil, err
	}
	if err := unmarshalScenarioJSON(s.Roles, &roles); err != nil {
		return nil, err
	}
	if err := unmarshalScenarioJSON(s.CompletionConditions, &completionConditions); err != nil {
		return nil, err
	}
	if err := unmarshalScenarioJSON(s.Nodes, &nodes); err != nil {
		return nil, err
	}
	if err := unmarshalScenarioJSON(s.ObservationRules, &rules); err != nil {
		return nil, err
	}

	structure := entity.ValidationStructure{
		HasGoal:                 strings.TrimSpace(goals["learner"]) != "" && strings.TrimSpace(goals["partner"]) != "",
		HasConflict:             len(contextBlock.ConflictType) > 0 && strings.TrimSpace(contextBlock.Description) != "",
		HasRoles:                hasScenarioParty(roles, "learner") && hasScenarioParty(roles, "partner"),
		HasCompletionConditions: len(completionConditions) > 0,
	}
	graph := validateGraph(nodes)
	coverage := validateCoverage(rules, nodes)
	valid := structure.HasGoal && structure.HasConflict && structure.HasRoles && structure.HasCompletionConditions && len(graph.DeadEnds) == 0 && len(graph.UnreachableNodes) == 0 && len(graph.Cycles) == 0 && len(coverage.UnderCoveredIndicators) == 0
	return &entity.ScenarioValidation{Valid: valid, Structure: structure, Graph: graph, Coverage: coverage}, nil
}

func unmarshalScenarioJSON(raw []byte, dst any) error {
	if len(raw) == 0 {
		return entity.ErrInvalidScenarioJSON
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		return entity.ErrInvalidScenarioJSON
	}
	return nil
}

type validationNode struct {
	ID      string             `json:"id"`
	Actions []validationAction `json:"actions"`
	IsEnd   bool               `json:"is_end"`
}

type validationAction struct {
	SemanticAction string  `json:"semantic_action"`
	NextNodeID     *string `json:"next_node_id"`
}

func validateGraph(nodes []validationNode) entity.ValidationGraph {
	graph := entity.ValidationGraph{DeadEnds: []string{}, UnreachableNodes: []string{}, Cycles: [][]string{}}
	if len(nodes) == 0 {
		return graph
	}
	nodeByID := make(map[string]validationNode, len(nodes))
	for _, node := range nodes {
		nodeByID[node.ID] = node
		if !node.IsEnd && len(node.Actions) == 0 {
			graph.DeadEnds = append(graph.DeadEnds, node.ID)
		}
	}
	visited := map[string]bool{}
	var walk func(id string, stack []string)
	walk = func(id string, stack []string) {
		if visited[id] {
			return
		}
		for idx, stackID := range stack {
			if stackID == id {
				graph.Cycles = append(graph.Cycles, append(append([]string{}, stack[idx:]...), id))
				return
			}
		}
		node, ok := nodeByID[id]
		if !ok {
			return
		}
		visited[id] = true
		stack = append(stack, id)
		for _, action := range node.Actions {
			if action.NextNodeID != nil && *action.NextNodeID != "" {
				walk(*action.NextNodeID, stack)
			}
		}
	}
	walk(nodes[0].ID, nil)
	for _, node := range nodes {
		if !visited[node.ID] {
			graph.UnreachableNodes = append(graph.UnreachableNodes, node.ID)
		}
	}
	return graph
}

func validateCoverage(rules []struct {
	SemanticAction string `json:"semantic_action"`
	Indicator      string `json:"indicator"`
}, nodes []validationNode) entity.ValidationCoverage {
	actionCounts := map[string]int{}
	for _, node := range nodes {
		for _, action := range node.Actions {
			actionCounts[action.SemanticAction]++
		}
	}
	coverage := entity.ValidationCoverage{UnderCoveredIndicators: []entity.UnderCoveredIndicator{}}
	seen := map[string]bool{}
	for _, rule := range rules {
		key := rule.SemanticAction + "\x00" + rule.Indicator
		if seen[key] {
			continue
		}
		seen[key] = true
		found := actionCounts[rule.SemanticAction]
		if found < 1 {
			coverage.UnderCoveredIndicators = append(coverage.UnderCoveredIndicators, entity.UnderCoveredIndicator{Indicator: rule.Indicator, ContextsFound: found, Required: 1})
		}
	}
	return coverage
}

func hasScenarioParty(roles []struct {
	ID string `json:"id"`
}, party string) bool {
	for _, role := range roles {
		if role.ID == party {
			return true
		}
	}
	return false
}
