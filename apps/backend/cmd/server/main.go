package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	"github.com/mondayy1/llm-games/internal/config"
	"github.com/mondayy1/llm-games/internal/db"
	"github.com/mondayy1/llm-games/internal/handler"
	"github.com/mondayy1/llm-games/internal/middleware"
	repository "github.com/mondayy1/llm-games/internal/repository/postgres"
	"github.com/mondayy1/llm-games/internal/usecase"
)

func main() {
	app := fx.New(
		fx.Provide(
			NewConfig,
			NewDB,
			echo.New,
		),
		fx.Provide(
			usecase.NewUserUseCase,
		),
		fx.Provide(
			repository.NewUserRepository,
		),
		fx.Invoke(
			middleware.Setup,
			handler.NewUserHandler,
		),
		fx.Invoke(StartServer),
	)

	app.Run()
}

func NewConfig() *config.Config {
	env := flag.String("env", "dev", "Environment (dev, prod)")
	flag.Parse()

	validEnvs := map[string]bool{"dev": true, "prod": true}
	if !validEnvs[*env] {
		log.Fatalf("Invalid environment: %s. Valid environments are: dev, prod", *env)
	}

	cfg, err := config.LoadConfig(*env)
	if err != nil {
		log.Fatalf("Config load error: %v", err)
	}
	fmt.Printf("config: %+v\n", cfg)

	return cfg
}

func NewDB(lc fx.Lifecycle, cfg *config.Config) *sql.DB {
	dbConn, err := db.NewDBConnection(cfg)
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return dbConn.Close()
		},
	})

	return dbConn
}

func StartServer(lc fx.Lifecycle, e *echo.Echo, cfg *config.Config) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := e.Start(fmt.Sprintf(":%d", cfg.App.Port)); err != nil {
					log.Fatal("Shutting down the server due to:", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return e.Shutdown(ctx)
		},
	})
}
