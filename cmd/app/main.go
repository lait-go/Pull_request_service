package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	config "revivers/config"
	_ "revivers/docs"
	"revivers/internal/adapters/cache"
	kafkaconsumer "revivers/internal/adapters/kafka_consumer"
	"revivers/internal/adapters/postgres"
	"revivers/internal/adapters/postgres/migrations"
	"revivers/internal/controlers/http"
	"revivers/internal/usecase"
	"revivers/pkg/http_server"
	"revivers/pkg/logger"
)

func main() {
	c, err := config.InitConfig()
	if err != nil {
		log.Fatalf("config init failed: %v", err)
	}

	logger.Init(c.Logger)
	logger.Info("logger initialized")

	logger.Info("starting migration...")
	if err = migrations.RunMigrate("internal/adapters/postgres/migrations", c.Migrate); err != nil {
		logger.Error("migration failed", err)
	} else {
		logger.Info("migration completed")
	}

	logger.Info("starting app...")

	if err = AppRun(context.Background(), c); err != nil {
		logger.Error("application run failed", err)
		os.Exit(1)
	}
}

// @title Subscription API
// @version 1.0
// @description API для управления подписками пользователей
// @host localhost:8081
// @BasePath /
// @schemes http
func AppRun(ctx context.Context, c *config.Config) error {
	post, err := postgres.New(ctx, c.Postgres)
	if err != nil {
		return err
	}
	logger.Info("postgres initialized")

	cache, err := cache.New(ctx, c.Cache, post)
	if err != nil {
		return err
	}
	logger.Info("cache initialized")

	kafka := kafkaconsumer.NewConsumer(c.Kafka)
	logger.Info("kafka initialized")

	profile := usecase.NewProfile(post, cache, kafka)
	logger.Info("usecase layer initialized")

	router := http.Router(profile, c.Http)
	logger.Info("router initialized")

	go func() {
		if err := profile.CreateOrder(ctx); err != nil {
			logger.Error("failed to create subscription", err)
		}
	}()

	server := http_server.New(router, c.Http)
	if err = server.Run(); err != nil {
		return err
	}
	logger.Info("HTTP server started on port: " + c.Http.Port)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig

	post.Close()
	kafka.Close()
	server.Close()

	return nil
}
