package migrations

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Config struct {
	User       string `envconfig:"MIGRATE_USER"     required:"true"`
	Password   string `envconfig:"MIGRATE_PASSWORD" required:"true"`
	UserDb     string `envconfig:"POSTGRES_USER"     required:"true"`
	PasswordDb string `envconfig:"POSTGRES_PASSWORD" required:"true"`
	Port       string `envconfig:"DB_PORT"     required:"true"`
	Host       string `envconfig:"DB_HOST"     required:"true"`
	DBName     string `envconfig:"DB_NAME"  required:"true"`
}

func (c *Config) DbKeyInit() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", c.User, c.Password, c.Host, c.Port, c.DBName)
}

func RunMigrate(path string, cfg Config) error {
	key := cfg.DbKeyInit()
	if key == "" {
		return fmt.Errorf("migration database connection string is empty")
	}
	mig, err := migrate.New(fmt.Sprintf("file://%s", path), key)
	if err != nil {
		return fmt.Errorf("Error to run migration: %w", err)
	}

	if err = mig.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("Error to up migration: %w", err)
	}

	db, err := sql.Open("postgres", key)
	if err != nil {
		return fmt.Errorf("Error opening database connection: %w", err)
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s'", cfg.UserDb, cfg.PasswordDb))
	if err != nil {
		return fmt.Errorf("Error creating user: %w", err)
	}
	_, err = db.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON SCHEMA public TO %s", cfg.UserDb))
	if err != nil {
		return fmt.Errorf("Error granting privileges on schema: %w", err)	
	}
	_, err = db.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO %s", cfg.UserDb))
	if err != nil {
		return fmt.Errorf("Error granting privileges on tables: %w", err)
	}
	_, err = db.Exec(fmt.Sprintf("GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO %s", cfg.UserDb))
	if err != nil {
		return fmt.Errorf("Error granting privileges on sequences: %w", err)
	}

	return nil
}
