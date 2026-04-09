package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL     string
	OpenRouterKey   string
	OpenRouterModel string
}

func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		OpenRouterKey:   os.Getenv("OPENROUTER_API_KEY"),
		OpenRouterModel: os.Getenv("OPENROUTER_MODEL"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.OpenRouterKey == "" {
		return Config{}, fmt.Errorf("OPENROUTER_API_KEY is required")
	}
	if cfg.OpenRouterModel == "" {
		cfg.OpenRouterModel = "openai/gpt-5.4-mini"
	}

	return cfg, nil
}
