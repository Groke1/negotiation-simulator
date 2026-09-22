package curator

import (
	"context"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
)

type repository interface {
	CreateAssignment(ctx context.Context, params entity.CreateAssignment) (*entity.Assignment, error)
	ListAssignments(ctx context.Context, filter entity.AssignmentFilter) ([]entity.Assignment, error)
	ListTeam(ctx context.Context, curatorID string, limit, offset int32) ([]entity.CuratorTeamMember, error)
	ListTeamSessions(ctx context.Context, curatorID string, limit, offset int32) ([]entity.TeamSessionListItem, error)
	TeamAnalytics(ctx context.Context, curatorID string) (*entity.TeamAnalytics, error)
}

type userRepository interface {
	GetUserByID(ctx context.Context, id string) (*entity.User, error)
	ListProfile(ctx context.Context, userID string) ([]entity.ProfileCluster, error)
}

type scenarioRepository interface {
	GetScenarioByID(ctx context.Context, id string) (*entity.Scenario, error)
}

type Service struct {
	repository         repository
	userRepository     userRepository
	scenarioRepository scenarioRepository
}

func NewService(repository repository, userRepository userRepository, scenarioRepository scenarioRepository) *Service {
	return &Service{repository: repository, userRepository: userRepository, scenarioRepository: scenarioRepository}
}

func (s *Service) ListTeam(ctx context.Context, claims entity.AccessClaims, limit, offset int32) ([]entity.CuratorTeamMember, error) {
	if err := requireCurator(claims); err != nil {
		return nil, err
	}
	members, err := s.repository.ListTeam(ctx, claims.UserID, limit, offset)
	if err != nil {
		return nil, err
	}
	for i := range members {
		profile, err := s.userRepository.ListProfile(ctx, members[i].UserID)
		if err != nil {
			return nil, err
		}
		members[i].Clusters = profile
	}
	return members, nil
}

func (s *Service) ListTeamSessions(ctx context.Context, claims entity.AccessClaims, limit, offset int32) ([]entity.TeamSessionListItem, error) {
	if err := requireCurator(claims); err != nil {
		return nil, err
	}
	return s.repository.ListTeamSessions(ctx, claims.UserID, limit, offset)
}

func (s *Service) ListAssignments(ctx context.Context, claims entity.AccessClaims, filter entity.AssignmentFilter) ([]entity.Assignment, error) {
	if err := requireCurator(claims); err != nil {
		return nil, err
	}
	filter.CuratorID = claims.UserID
	return s.repository.ListAssignments(ctx, filter)
}

func (s *Service) CreateAssignment(ctx context.Context, claims entity.AccessClaims, learnerID, scenarioID string) (*entity.Assignment, error) {
	if err := requireCurator(claims); err != nil {
		return nil, err
	}
	curator, err := s.userRepository.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	learner, err := s.userRepository.GetUserByID(ctx, learnerID)
	if err != nil {
		return nil, err
	}
	if curator.OrganizationID == nil || learner.OrganizationID == nil || *curator.OrganizationID != *learner.OrganizationID {
		return nil, entity.ErrAssignmentAccessDenied
	}
	scenario, err := s.scenarioRepository.GetScenarioByID(ctx, scenarioID)
	if err != nil {
		return nil, err
	}
	if scenario.Status != entity.ScenarioStatusPublished {
		return nil, entity.ErrScenarioNotAvailable
	}
	assignment, err := s.repository.CreateAssignment(ctx, entity.CreateAssignment{CuratorID: claims.UserID, LearnerID: learnerID, ScenarioID: scenarioID})
	if err != nil {
		return nil, err
	}
	assignment.ScenarioTitle = scenario.Title
	return assignment, nil
}

func (s *Service) TeamAnalytics(ctx context.Context, claims entity.AccessClaims) (*entity.TeamAnalytics, error) {
	if err := requireCurator(claims); err != nil {
		return nil, err
	}
	analytics, err := s.repository.TeamAnalytics(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	members, err := s.ListTeam(ctx, claims, 100, 0)
	if err != nil {
		return nil, err
	}
	if len(members) == 0 {
		return analytics, nil
	}
	byCluster := map[string]entity.TeamAnalyticsCluster{}
	counts := map[string]int{}
	for _, member := range members {
		for _, cluster := range member.Clusters {
			item := byCluster[cluster.ClusterID]
			item.ClusterID = cluster.ClusterID
			item.Code = cluster.Code
			item.Name = cluster.Name
			item.AverageScore += cluster.Score
			byCluster[cluster.ClusterID] = item
			counts[cluster.ClusterID]++
		}
	}
	for id, item := range byCluster {
		if counts[id] > 0 {
			item.AverageScore /= counts[id]
		}
		analytics.Clusters = append(analytics.Clusters, item)
	}
	return analytics, nil
}

func requireCurator(claims entity.AccessClaims) error {
	if claims.ActiveRole != entity.RoleCurator {
		return entity.ErrCuratorRoleRequired
	}
	return nil
}
