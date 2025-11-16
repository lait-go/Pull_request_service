package usecase

import (
	"context"
	"revivers/internal/domain"
)

// UseCase определяет всю бизнес-логику приложения
type UseCase interface {
	// PullRequest методы

	// CreatePullRequest создает PR и автоматически назначает до 2 ревьюверов из команды автора
	CreatePullRequest(ctx context.Context, req domain.PostPullRequestCreateJSONBody) (*domain.PullRequest, error)

	// MergePullRequest помечает PR как MERGED (идемпотентная операция)
	MergePullRequest(ctx context.Context, req domain.PostPullRequestMergeJSONBody) (*domain.PullRequest, error)

	// ReassignReviewer переназначает конкретного ревьювера на другого из его команды
	// Возвращает PR и ID нового ревьювера
	ReassignReviewer(ctx context.Context, req domain.PostPullRequestReassignJSONBody) (*domain.PullRequest, string, error)

	// GetPullRequestsByReviewer возвращает PR'ы, где пользователь назначен ревьювером
	GetPullRequestsByReviewer(ctx context.Context, userID string) ([]domain.PullRequestShort, error)

	// Team методы

	// CreateOrUpdateTeam создает команду с участниками (создаёт/обновляет пользователей)
	CreateOrUpdateTeam(ctx context.Context, team domain.Team) error

	// GetTeam возвращает команду с участниками
	GetTeam(ctx context.Context, teamName string) (*domain.Team, error)

	// SetUserActive устанавливает флаг активности пользователя
	// Возвращает обновленного пользователя
	SetUserActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
}
