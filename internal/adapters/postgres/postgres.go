package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Config struct {
	Source string `envconfig:"DB_SOURCE" required:"true"`
}

type Pool struct {
	DB *sqlx.DB
}

func New(ctx context.Context, cfg Config) (*Pool, error) {
	db, err := sqlx.ConnectContext(ctx, "postgres", cfg.Source)
	if err != nil {
		return nil, fmt.Errorf("failed connection to bd: %w", err)
	}

	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed pinging database: %w", err)
	}

	return &Pool{
		DB: db,
	}, nil
}

func (p *Pool) Close() {
	if err := p.DB.Close(); err != nil {
		fmt.Printf("failed to close DB: %v\n", err)
	}
}
