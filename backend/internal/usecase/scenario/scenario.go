package scenario

import (
	"context"
	"encoding/json"
	"sync"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
)

type scenarioRepository interface {
	CreateScenario(ctx context.Context, scenario entity.Scenario) (*entity.Scenario, error)
	GetScenarioByID(ctx context.Context, id string) (*entity.Scenario, error)
	ListScenariosByOrganization(ctx context.Context, organizationID string, params entity.ListScenarios) ([]entity.Scenario, error)
	UpdateScenario(ctx context.Context, scenario entity.Scenario) (*entity.Scenario, error)
	SaveScenarioValidation(ctx context.Context, id string, result []byte) (*entity.Scenario, error)
	SetScenarioStatus(ctx context.Context, id string, status entity.ScenarioStatus) (*entity.Scenario, error)
	AddFavorite(ctx context.Context, userID, scenarioID string) error
	RemoveFavorite(ctx context.Context, userID, scenarioID string) error
	GetPersonalStats(ctx context.Context, userID, scenarioID string) (*entity.ScenarioPersonalStats, error)
}

type profileRepository interface {
	ListCompetencyClusters(ctx context.Context) ([]entity.CompetencyCluster, error)
	ListSemanticActions(ctx context.Context) ([]entity.SemanticActionDefinition, error)
}

type userRepository interface {
	GetUserByID(ctx context.Context, id string) (*entity.User, error)
}

type transactor interface {
	WithTx(ctx context.Context, f func(ctx context.Context) error) error
}

type scenarioService struct {
	scenarioRepository scenarioRepository
	profileRepository  profileRepository
	userRepository     userRepository
	transactor         transactor
	ai                 aiGenerator
	generationMu       sync.RWMutex
	generationRequests map[string]*entity.AIGenerationRequest
}

func NewScenarioService(scenarioRepository scenarioRepository, profileRepository profileRepository, userRepository userRepository, transactor transactor, ai aiGenerator) *scenarioService {
	return &scenarioService{scenarioRepository: scenarioRepository, profileRepository: profileRepository, userRepository: userRepository, transactor: transactor, ai: ai, generationRequests: map[string]*entity.AIGenerationRequest{}}
}

func (s *scenarioService) ListCompetencyClusters(ctx context.Context) ([]entity.CompetencyCluster, error) {
	return s.profileRepository.ListCompetencyClusters(ctx)
}

func (s *scenarioService) ListSemanticActions(ctx context.Context) ([]entity.SemanticActionDefinition, error) {
	return s.profileRepository.ListSemanticActions(ctx)
}

func (s *scenarioService) ListScenarios(ctx context.Context, claims entity.AccessClaims, params entity.ListScenarios) ([]entity.Scenario, error) {
	user, err := s.userRepository.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if user.OrganizationID == nil {
		return nil, entity.ErrOrganizationRequired
	}
	params.UserID = claims.UserID
	if claims.ActiveRole != entity.RoleMethodist {
		status := entity.ScenarioStatusPublished
		params.Status = &status
	}
	return s.scenarioRepository.ListScenariosByOrganization(ctx, *user.OrganizationID, params)
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
	if normalizedNodes, err := normalizeScenarioNodes(scenario.Nodes); err != nil {
		return err
	} else {
		scenario.Nodes = normalizedNodes
	}
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

func (s *scenarioService) PublishScenario(ctx context.Context, claims entity.AccessClaims, scenarioID string) (*entity.Scenario, *entity.ScenarioValidation, error) {
	if err := requireMethodist(claims); err != nil {
		return nil, nil, err
	}
	var published *entity.Scenario
	var validation *entity.ScenarioValidation
	err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		scenario, err := s.scenarioRepository.GetScenarioByID(txCtx, scenarioID)
		if err != nil {
			return err
		}
		if err := ensureOwnedBy(claims.UserID, scenario); err != nil {
			return err
		}
		if scenario.Status != entity.ScenarioStatusDraft {
			return entity.ErrInvalidScenarioStatus
		}
		validation, err = validateScenario(scenario)
		if err != nil {
			return err
		}
		if !validation.Valid {
			return entity.ErrScenarioValidationFailed
		}
		published, err = s.scenarioRepository.SetScenarioStatus(txCtx, scenario.ID, entity.ScenarioStatusPublished)
		return err
	})
	if err != nil {
		return nil, nil, err
	}
	return published, validation, nil
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

func (s *scenarioService) CloneScenario(ctx context.Context, claims entity.AccessClaims, scenarioID string) (*entity.Scenario, error) {
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
	source, err := s.scenarioRepository.GetScenarioByID(ctx, scenarioID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureReadable(ctx, claims, source); err != nil {
		return nil, err
	}
	id, err := entity.NewUUID()
	if err != nil {
		return nil, err
	}
	copy := *source
	copy.ID = id
	copy.OrganizationID = user.OrganizationID
	copy.CreatedBy = &user.ID
	copy.Status = entity.ScenarioStatusDraft
	copy.ValidatedRevision = nil
	copy.ValidationResult = nil
	copy.PublishedAt = nil
	return s.scenarioRepository.CreateScenario(ctx, copy)
}

func (s *scenarioService) AddFavorite(ctx context.Context, claims entity.AccessClaims, scenarioID string) error {
	scenario, err := s.GetScenario(ctx, claims, scenarioID)
	if err != nil {
		return err
	}
	if scenario.Status != entity.ScenarioStatusPublished {
		return entity.ErrInvalidScenarioStatus
	}
	return s.scenarioRepository.AddFavorite(ctx, claims.UserID, scenario.ID)
}

func (s *scenarioService) RemoveFavorite(ctx context.Context, claims entity.AccessClaims, scenarioID string) error {
	scenario, err := s.GetScenario(ctx, claims, scenarioID)
	if err != nil {
		return err
	}
	return s.scenarioRepository.RemoveFavorite(ctx, claims.UserID, scenario.ID)
}

func (s *scenarioService) GetPersonalStats(ctx context.Context, userID, scenarioID string) (*entity.ScenarioPersonalStats, error) {
	return s.scenarioRepository.GetPersonalStats(ctx, userID, scenarioID)
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

func normalizeScenarioNodes(raw json.RawMessage) (json.RawMessage, error) {
	blocks := []entity.ScenarioBlock{}
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return nil, entity.ErrInvalidScenarioJSON
	}
	entity.NormalizeScenarioBlocks(blocks)
	for i := range blocks {
		dialogue := blocks[i].DeterministicDialogue
		if dialogue == nil {
			continue
		}
		if len(dialogue.Nodes) == 0 && dialogue.StartNodeID == "" && dialogue.LearnerRoleID == "" {
			blocks[i].DeterministicDialogue = nil
			continue
		}
	}
	normalized, err := json.Marshal(blocks)
	if err != nil {
		return nil, err
	}
	return normalized, nil
}

func validateScenario(s *entity.Scenario) (*entity.ScenarioValidation, error) {
	doc := entity.ScenarioDocumentFromScenario(*s)
	issues := []entity.ScenarioValidationIssue{}
	if doc.Description == "" {
		issues = append(issues, validationIssue("error", "MISSING_DESCRIPTION", "Нужно заполнить описание сценария.", "scenario", nil))
	}
	if len(doc.Roles) < 2 {
		issues = append(issues, validationIssue("error", "MISSING_ROLES", "В сценарии должны быть как минимум две роли.", "scenario", nil))
	}
	if len(doc.SupportedRuntimes) == 0 {
		issues = append(issues, validationIssue("error", "MISSING_RUNTIME", "Нужно выбрать хотя бы один режим тренировки.", "scenario", nil))
	}
	if len(doc.SupportedChannels) == 0 {
		issues = append(issues, validationIssue("error", "MISSING_CHANNEL", "Нужно выбрать хотя бы один канал общения.", "scenario", nil))
	}
	if len(doc.Blocks) == 0 {
		issues = append(issues, validationIssue("error", "MISSING_BLOCKS", "Нужно добавить хотя бы один блок сценария.", "scenario", nil))
	}
	roleIDs := make(map[string]bool, len(doc.Roles))
	for _, role := range doc.Roles {
		roleIDs[role.ID] = true
	}
	targetClusterIDs := make(map[string]bool, len(doc.TargetClusterIDs))
	for _, clusterID := range doc.TargetClusterIDs {
		targetClusterIDs[clusterID] = true
	}
	for _, block := range doc.Blocks {
		if block.Type == "end" {
			if block.Result == nil || block.Result.Summary == "" {
				id := block.ID
				issues = append(issues, validationIssue("error", "END_BLOCK_RESULT_MISSING", "Финальный блок должен содержать результат переговоров.", "block", &id))
			}
		} else if block.Result != nil {
			id := block.ID
			issues = append(issues, validationIssue("error", "NON_END_BLOCK_RESULT_UNEXPECTED", "Результат переговоров можно задавать только в финальном блоке.", "block", &id))
		}
		if !containsString(doc.SupportedRuntimes, "deterministic") || block.Type == "start" {
			continue
		}
		if block.DeterministicDialogue == nil || len(block.DeterministicDialogue.Nodes) == 0 {
			if block.Type == "end" {
				continue
			}
			id := block.ID
			issues = append(issues, validationIssue("error", "MISSING_DIALOGUE_NODES", "Для детерминированного режима блок должен содержать дерево реплик.", "block", &id))
			continue
		}
		if block.DeterministicDialogue.LearnerRoleID == "" || !roleIDs[block.DeterministicDialogue.LearnerRoleID] {
			id := block.ID
			issues = append(issues, validationIssue("error", "INVALID_DIALOGUE_LEARNER_ROLE", "Детерминированное дерево должно ссылаться на существующую роль пользователя.", "block", &id))
		}
		for _, node := range block.DeterministicDialogue.Nodes {
			targets := map[string]bool{}
			for _, target := range node.AssessmentTargets {
				if !targetClusterIDs[target.ClusterID] {
					id := block.ID
					issues = append(issues, validationIssue("error", "ASSESSMENT_TARGET_CLUSTER_NOT_TARGETED", "Оцениваемый кластер должен входить в целевые кластеры сценария.", "block", &id))
				}
				if targets[target.ClusterID] {
					id := block.ID
					issues = append(issues, validationIssue("error", "ASSESSMENT_TARGET_DUPLICATED", "Оцениваемый кластер продублирован в узле диалога.", "block", &id))
				}
				targets[target.ClusterID] = true
				if !validAssessmentWeight(target.Weight) {
					id := block.ID
					issues = append(issues, validationIssue("error", "ASSESSMENT_TARGET_WEIGHT_INVALID", "Вес оцениваемого кластера должен быть 0.5, 1 или 1.5.", "block", &id))
				}
			}
			for _, option := range node.ReplyOptions {
				if option.RoleID != block.DeterministicDialogue.LearnerRoleID {
					id := block.ID
					issues = append(issues, validationIssue("error", "INVALID_DIALOGUE_REPLY_ROLE", "Роль варианта ответа должна совпадать с learner_role_id детерминированного дерева.", "block", &id))
					break
				}
				assessments := map[string]bool{}
				for _, assessment := range option.Assessments {
					if assessment.Level < 0 || assessment.Level > 4 {
						id := block.ID
						issues = append(issues, validationIssue("error", "REPLY_OPTION_ASSESSMENT_LEVEL_INVALID", "Уровень оценки варианта ответа должен быть от 0 до 4.", "block", &id))
					}
					if !targets[assessment.ClusterID] {
						id := block.ID
						issues = append(issues, validationIssue("error", "REPLY_OPTION_ASSESSMENT_UNEXPECTED", "Вариант ответа оценивает кластер, который не указан среди целей узла.", "block", &id))
					}
					if assessments[assessment.ClusterID] {
						id := block.ID
						issues = append(issues, validationIssue("error", "REPLY_OPTION_ASSESSMENT_DUPLICATED", "Оценка кластера продублирована в варианте ответа.", "block", &id))
					}
					assessments[assessment.ClusterID] = true
				}
				for clusterID := range targets {
					if !assessments[clusterID] {
						id := block.ID
						issues = append(issues, validationIssue("error", "REPLY_OPTION_ASSESSMENT_MISSING", "Вариант ответа должен содержать оценку по каждому целевому кластеру узла.", "block", &id))
					}
				}
			}
		}
	}
	issues = append(issues, blockGraphIssues(doc.Blocks)...)
	errorsCount := 0
	warningsCount := 0
	for _, issue := range issues {
		if issue.Severity == "warning" {
			warningsCount++
		} else {
			errorsCount++
		}
	}
	return &entity.ScenarioValidation{Valid: errorsCount == 0, ErrorsCount: errorsCount, WarningsCount: warningsCount, Issues: issues}, nil
}

func validAssessmentWeight(value float64) bool {
	return value == 0.5 || value == 1 || value == 1.5
}

func blockGraphIssues(blocks []entity.ScenarioBlock) []entity.ScenarioValidationIssue {
	issues := []entity.ScenarioValidationIssue{}
	startIDs := []string{}
	endCount := 0
	for _, block := range blocks {
		switch block.Type {
		case "start":
			startIDs = append(startIDs, block.ID)
		case "end":
			endCount++
		}
	}
	if len(startIDs) != 1 {
		issues = append(issues, validationIssue("error", "START_BLOCK_COUNT_INVALID", "В сценарии должен быть ровно один начальный блок.", "scenario", nil))
	}
	if len(blocks) > 0 && endCount == 0 {
		issues = append(issues, validationIssue("error", "END_BLOCK_MISSING", "В сценарии должен быть хотя бы один финальный блок.", "scenario", nil))
	}
	graph := make(map[string][]string, len(blocks))
	known := make(map[string]bool, len(blocks))
	for _, block := range blocks {
		known[block.ID] = true
	}
	for _, block := range blocks {
		for _, transition := range block.Transitions {
			if transition.Forbidden || !known[transition.TargetBlockID] {
				continue
			}
			graph[block.ID] = append(graph[block.ID], transition.TargetBlockID)
		}
	}
	if len(startIDs) == 1 {
		reachable := reachableBlocks(graph, startIDs[0])
		for _, block := range blocks {
			if !reachable[block.ID] {
				id := block.ID
				issues = append(issues, validationIssue("error", "BLOCK_UNREACHABLE", "Блок недостижим из начального состояния.", "block", &id))
			}
		}
	}
	issues = append(issues, blockGraphCycleIssues(graph, blocks)...)
	return issues
}

func reachableBlocks(graph map[string][]string, startID string) map[string]bool {
	reachable := map[string]bool{}
	var visit func(string)
	visit = func(blockID string) {
		if reachable[blockID] {
			return
		}
		reachable[blockID] = true
		for _, nextID := range graph[blockID] {
			visit(nextID)
		}
	}
	visit(startID)
	return reachable
}

func blockGraphCycleIssues(graph map[string][]string, blocks []entity.ScenarioBlock) []entity.ScenarioValidationIssue {
	const (
		unvisited = 0
		visiting  = 1
		visited   = 2
	)
	state := make(map[string]int, len(blocks))
	reported := map[string]bool{}
	issues := []entity.ScenarioValidationIssue{}
	var visit func(string) bool
	visit = func(blockID string) bool {
		switch state[blockID] {
		case visiting:
			return true
		case visited:
			return false
		}
		state[blockID] = visiting
		for _, nextID := range graph[blockID] {
			if visit(nextID) {
				if !reported[blockID] {
					id := blockID
					issues = append(issues, validationIssue("error", "BLOCK_GRAPH_CYCLE", "В графе блоков обнаружен цикл. Сценарий должен вести к финальному блоку без возвратов к уже пройденным состояниям.", "block", &id))
					reported[blockID] = true
				}
				state[blockID] = visited
				return true
			}
		}
		state[blockID] = visited
		return false
	}
	for _, block := range blocks {
		if state[block.ID] == unvisited {
			visit(block.ID)
		}
	}
	return issues
}

func validationIssue(severity, code, message, entityType string, entityID *string) entity.ScenarioValidationIssue {
	return entity.ScenarioValidationIssue{Severity: severity, Code: code, Message: message, EntityType: entityType, EntityID: entityID}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
