package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	"github.com/everyday-studio/ollm/internal/config"
	"github.com/everyday-studio/ollm/internal/db"
	"github.com/everyday-studio/ollm/internal/handler"
	"github.com/everyday-studio/ollm/internal/middleware"
	repository "github.com/everyday-studio/ollm/internal/repository/postgres"
	"github.com/everyday-studio/ollm/internal/usecase"
)

func main() {
	app := fx.New(
		fx.Provide(
			NewConfig,
			NewLogger,
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

func NewLogger(cfg *config.Config) *slog.Logger {
	logLevel := slog.LevelInfo
	switch strings.ToLower(cfg.App.LogLevel) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	logHandler := slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: logLevel,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					return slog.String(a.Key, time.Now().Format("2006-01-02T15:04:05.000Z07:00"))
				}
				return a
			},
		},
	)
	return slog.New(logHandler)
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
