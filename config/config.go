package config

import (
	"errors"
	"revivers/internal/adapters/postgres"
	"revivers/pkg/http_server"
	"revivers/pkg/logger"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Postgres postgres.Config
	Http     http_server.Config
	Logger   logger.Config
}

func InitConfig() (*Config, error) {
	var cfg Config

	// Загружаем .env файл, если он существует (не критично, если его нет)
	// Переменные окружения из docker-compose.yml имеют приоритет
	_ = godotenv.Load()

	err := envconfig.Process("", &cfg)
	if err != nil {
		return &Config{}, errors.New("Error value: " + err.Error())
	}

	return &cfg, nil
}
