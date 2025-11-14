package usecase

import (
	"revivers/internal/adapters/postgres"
)

type Postgres interface {
	// CreateOrder(ctx context.Context, sub domain.Order) error
	// GetOrder(ctx context.Context, user_id string) (domain.Order, error)
}

type Unimplemented struct {
	Postgres Postgres
}

func NewProfile(postgres *postgres.Pool) *Unimplemented {
	return &Unimplemented{
		Postgres: postgres,
	}
}
