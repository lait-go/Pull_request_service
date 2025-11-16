package usecase

import (
	"context"
	"revivers/internal/domain"
)

func (s *Service) CreatePullRequest(ctx context.Context, req domain.PostPullRequestCreateJSONBody) (*domain.PullRequest, error) {
	// TODO: реализовать бизнес-логику создания PR
	return nil, nil
}

func (s *Service) MergePullRequest(ctx context.Context, req domain.PostPullRequestMergeJSONBody) error {
	// TODO: реализовать бизнес-логику мержа PR
	return nil
}

func (s *Service) ReassignReviewer(ctx context.Context, req domain.PostPullRequestReassignJSONBody) error {
	// TODO: реализовать бизнес-логику переназначения ревьювера
	return nil
}

func (s *Service) GetPullRequestsByReviewer(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	// TODO: реализовать бизнес-логику получения PR по ревьюверу
	return nil, nil
}

// Реализация TeamUseCase

func (s *Service) CreateOrUpdateTeam(ctx context.Context, team domain.Team) error {
	exists, err := s.repo.TeamExists(ctx, team.TeamName)
	if err != nil {
		return err
	}

	if exists {
		if err := s.repo.UpdateTeam(ctx, team); err != nil {
			return err
		}
	} else {
		if err := s.repo.CreateTeam(ctx, team); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	// TODO: реализовать бизнес-логику получения команды
	return nil, nil
}

// Реализация UserUseCase

func (s *Service) SetUserActive(ctx context.Context, userID string, isActive bool) error {
	// TODO: реализовать бизнес-логику установки активности пользователя
	return nil
}
