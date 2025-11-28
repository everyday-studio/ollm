package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/mondayy1/llm-games/internal/config"
)

func main() {
	// YAML Load
	env := flag.String("env", "dev", "Environment (dev, prod)")
	flag.Parse()

	validEnvs := map[string]bool{"dev": true, "prod": true}
	if !validEnvs[*env] {
		log.Fatalf("Invalid environment: %s", *env)
	}

	cfg, err := config.LoadConfig(*env)
	if err != nil {
		log.Fatalf("Config load error: %v", err)
	}

	fmt.Printf("config: %+v\n", cfg)
	e := echo.New()

	log.Printf("Server started at %s", fmt.Sprintf(":%d", cfg.App.Port))
	if err := e.Start(fmt.Sprintf(":%d", cfg.App.Port)); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
