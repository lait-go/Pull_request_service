package usecase

import (
	"context"
	"revivers/internal/adapters/postgres"
	"revivers/internal/domain"
)

// PostgresRepository определяет интерфейс для работы с базой данных
type PostgresRepository interface {
	// Team методы
	TeamGet(ctx context.Context, teamName string) (domain.Team, error)
	TeamExists(ctx context.Context, teamName string) (bool, error)
	UpdateTeam(ctx context.Context, data domain.Team) error
	CreateTeam(ctx context.Context, data domain.Team) error

	// PullRequest методы
	PullRequestCreate(ctx context.Context, prID, prName, authorUserID string, reviewerUserIDs []string) (*domain.PullRequest, error)
	PullRequestGet(ctx context.Context, prID string) (*domain.PullRequest, error)
	PullRequestMerge(ctx context.Context, prID string) error
	PullRequestReassignReviewer(ctx context.Context, prID, oldUserID, newUserID string) error
	PullRequestsGetByReviewer(ctx context.Context, userID string) ([]domain.PullRequestShort, error)

	// User методы
	UserSetActive(ctx context.Context, userID string, isActive bool) error
	UserGetByID(ctx context.Context, userID string) (int, int, error) // возвращает internal ID, teamID, error
	ActiveUsersGetByTeamID(ctx context.Context, teamID int, excludeUserID int) ([]struct {
		UserID     string
		InternalID int
	}, error)
}

// Service реализует все usecase интерфейсы
type Service struct {
	repo PostgresRepository
}

// NewService создает новый сервис usecase
func NewService(repo *postgres.Pool) *Service {
	return &Service{
		repo: repo,
	}
}
