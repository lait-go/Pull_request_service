package usecase

import (
	"context"
	"revivers/internal/adapters/postgres"
	"revivers/internal/domain"
)

// PostgresRepository определяет интерфейс для работы с базой данных
type PostgresRepository interface {
	TeamGet(ctx context.Context, teamName string) (domain.Team, error)
	TeamExists(ctx context.Context, teamName string) (bool, error)
	UpdateTeam(ctx context.Context, data domain.Team) error
	CreateTeam(ctx context.Context, data domain.Team) error
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
