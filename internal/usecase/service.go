package usecase

import (
	"context"
	"fmt"
	"math/rand"
	"revivers/internal/domain"
)

func (s *Service) CreatePullRequest(ctx context.Context, req domain.PostPullRequestCreateJSONBody) (*domain.PullRequest, error) {
	// Получаем автора и его команду
	authorInternalID, teamID, err := s.repo.UserGetByID(ctx, req.AuthorId)
	if err != nil {
		return nil, err
	}

	// Проверяем, что автор в команде
	if teamID == 0 {
		return nil, &domain.NoCandidateError{
			TeamName: "author has no team",
		}
	}

	// Получаем активных пользователей команды (исключая автора)
	activeUsers, err := s.repo.ActiveUsersGetByTeamID(ctx, teamID, authorInternalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}

	// Согласно ТЗ: "Если доступных кандидатов меньше двух, назначается доступное количество (0/1)"
	// Если кандидатов 0, создаем PR без ревьюверов
	reviewerUserIDs := make([]string, 0)
	if len(activeUsers) > 0 {
		// Выбираем до 2 ревьюверов случайным образом
		maxReviewers := 2
		if len(activeUsers) < maxReviewers {
			maxReviewers = len(activeUsers)
		}

		// Перемешиваем список для случайного выбора
		shuffled := make([]struct {
			UserID     string
			InternalID int
		}, len(activeUsers))
		copy(shuffled, activeUsers)
		rand.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})

		// Выбираем ревьюверов
		reviewerUserIDs = make([]string, 0, maxReviewers)
		for i := 0; i < maxReviewers; i++ {
			reviewerUserIDs = append(reviewerUserIDs, shuffled[i].UserID)
		}
	}

	// Создаем PR с назначенными ревьюверами
	pr, err := s.repo.PullRequestCreate(ctx, req.PullRequestId, req.PullRequestName, req.AuthorId, reviewerUserIDs)
	if err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *Service) MergePullRequest(ctx context.Context, req domain.PostPullRequestMergeJSONBody) (*domain.PullRequest, error) {
	if err := s.repo.PullRequestMerge(ctx, req.PullRequestId); err != nil {
		return nil, err
	}
	// Получаем обновленный PR
	pr, err := s.repo.PullRequestGet(ctx, req.PullRequestId)
	if err != nil {
		return nil, err
	}
	return pr, nil
}

func (s *Service) ReassignReviewer(ctx context.Context, req domain.PostPullRequestReassignJSONBody) (*domain.PullRequest, string, error) {
	// Получаем PR для проверки автора
	pr, err := s.repo.PullRequestGet(ctx, req.PullRequestId)
	if err != nil {
		return nil, "", err
	}

	// Проверяем, что PR не MERGED
	if pr.Status == domain.PullRequestStatusMERGED {
		return nil, "", &domain.PullRequestMergedError{
			PullRequestID: req.PullRequestId,
		}
	}

	// Получаем старого ревьювера и его команду
	oldReviewerInternalID, teamID, err := s.repo.UserGetByID(ctx, req.OldUserId)
	if err != nil {
		return nil, "", err
	}

	// Проверяем, что старый ревьювер в команде
	if teamID == 0 {
		return nil, "", &domain.NoCandidateError{
			TeamName: "reviewer has no team",
		}
	}

	// Получаем автора PR для проверки, что новый ревьювер не является автором
	authorInternalID, _, err := s.repo.UserGetByID(ctx, pr.AuthorId)
	if err != nil {
		return nil, "", err
	}

	// Согласно ТЗ: "Переназначение заменяет одного ревьювера на случайного активного участника из команды заменяемого ревьювера"
	// Получаем активных пользователей команды заменяемого ревьювера (исключая автора PR)
	activeUsers, err := s.repo.ActiveUsersGetByTeamID(ctx, teamID, authorInternalID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get active users: %w", err)
	}

	// Фильтруем, исключая старого ревьювера и автора PR
	candidates := make([]struct {
		UserID     string
		InternalID int
	}, 0)
	for _, user := range activeUsers {
		// Исключаем старого ревьювера и автора PR
		if user.InternalID != oldReviewerInternalID && user.InternalID != authorInternalID {
			candidates = append(candidates, user)
		}
	}

	// Проверяем, что есть кандидаты для переназначения
	if len(candidates) == 0 {
		return nil, "", &domain.NoCandidateError{
			TeamName: "no active candidates for reassignment",
		}
	}

	// Выбираем случайного нового ревьювера
	newReviewer := candidates[rand.Intn(len(candidates))]

	// Переназначаем ревьювера через репозиторий
	if err := s.repo.PullRequestReassignReviewer(ctx, req.PullRequestId, req.OldUserId, newReviewer.UserID); err != nil {
		return nil, "", err
	}

	// Получаем обновленный PR
	updatedPR, err := s.repo.PullRequestGet(ctx, req.PullRequestId)
	if err != nil {
		return nil, "", err
	}

	return updatedPR, newReviewer.UserID, nil
}

func (s *Service) GetPullRequestsByReviewer(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	prs, err := s.repo.PullRequestsGetByReviewer(ctx, userID)
	if err != nil {
		return nil, err
	}
	return prs, nil
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
	team, err := s.repo.TeamGet(ctx, teamName)
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// Реализация UserUseCase

func (s *Service) SetUserActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	if err := s.repo.UserSetActive(ctx, userID, isActive); err != nil {
		return nil, err
	}
	// Получаем обновленного пользователя
	user, err := s.repo.UserGet(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}
