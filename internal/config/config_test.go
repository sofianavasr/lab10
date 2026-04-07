package config

import "testing"

func TestLoad(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test")
	t.Setenv("OPENROUTER_API_KEY", "key")
	t.Setenv("OPENROUTER_MODEL", "anthropic/claude-3.5-haiku")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DatabaseURL != "postgres://test" {
		t.Fatalf("unexpected database url: %q", cfg.DatabaseURL)
	}
	if cfg.OpenRouterKey != "key" {
		t.Fatalf("unexpected key: %q", cfg.OpenRouterKey)
	}
	if cfg.OpenRouterModel != "anthropic/claude-3.5-haiku" {
		t.Fatalf("unexpected model: %q", cfg.OpenRouterModel)
	}
}

func TestLoad_DefaultModel(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test")
	t.Setenv("OPENROUTER_API_KEY", "key")
	t.Setenv("OPENROUTER_MODEL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.OpenRouterModel != "openai/gpt-5.4-mini" {
		t.Fatalf("expected default model, got %q", cfg.OpenRouterModel)
	}
}

func TestLoad_MissingEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("OPENROUTER_API_KEY", "")
	t.Setenv("OPENROUTER_MODEL", "")

	if _, err := Load(); err == nil {
		t.Fatal("expected error when required env vars are missing")
	}
}
