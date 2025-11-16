package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	config "revivers/config"
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
	if err = migrations.RunMigrate("internal/adapters/postgres/migrations/", c.Postgres); err != nil {
		logger.Error("migration failed", err)
		os.Exit(1)
	} else {
		logger.Info("migration completed")
	}

	logger.Info("starting app...")

	if err = AppRun(context.Background(), c); err != nil {
		logger.Error("application run failed", err)
		os.Exit(1)
	}
}

func AppRun(ctx context.Context, c *config.Config) error {
	post, err := postgres.New(ctx, c.Postgres)
	if err != nil {
		return err
	}
	logger.Info("postgres initialized")

	// Создаем usecase сервис
	service := usecase.NewService(post)
	logger.Info("usecase layer initialized")

	// Создаем HTTP контроллер, который реализует ServerInterface
	controller := http.NewController(service)
	logger.Info("HTTP controller initialized")

	router := http.HandlerWithBaseURL(controller, "")
	logger.Info("router initialized")

	server := http_server.New(router, c.Http)
	if err = server.Run(); err != nil {
		return err
	}
	logger.Info("HTTP server started on URL: " + c.Http.Url)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig

	post.Close()
	server.Close()

	return nil
}
