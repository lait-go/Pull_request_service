package config

import (
	"errors"
	kafkaconsumer "revivers/internal/adapters/kafka_consumer"
	"revivers/internal/adapters/postgres"
	"revivers/internal/adapters/postgres/migrations"
	"revivers/pkg/http_server"
	"revivers/pkg/logger"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Migrate  migrations.Config
	Postgres postgres.Config
	Kafka    kafkaconsumer.Config
	Http     http_server.Config
	Logger   logger.Config
}

func InitConfig() (*Config, error) {
	var cfg Config

	if err := godotenv.Load(); err != nil {
		return &Config{}, errors.New("Error loading .env file")
	}

	err := envconfig.Process("", &cfg)
	if err != nil {
		return &Config{}, errors.New("Error value: " + err.Error())
	}

	return &cfg, nil
}
