package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Config struct {
	User     string `envconfig:"POSTGRES_USER"     required:"true"`
	Password string `envconfig:"POSTGRES_PASSWORD" required:"true"`
	Port     string `envconfig:"DB_PORT"     required:"true"`
	Host     string `envconfig:"DB_HOST"     required:"true"`
	DBName   string `envconfig:"DB_NAME"  required:"true"`
}

type Pool struct {
	DB *sqlx.DB
}

func (c *Config) DbKeyInit() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", c.User, c.Password, c.Host, c.Port, c.DBName)
}

func New(ctx context.Context, cfg Config) (*Pool, error) {
	db, err := sqlx.ConnectContext(ctx, "postgres", cfg.DbKeyInit())
	if err != nil {
		return nil, fmt.Errorf("Error connection to bd: %w", err)
	}

	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
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
