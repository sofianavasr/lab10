package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sofia/lab10/internal/config"
	"github.com/sofia/lab10/internal/llm"
	"github.com/sofia/lab10/internal/repository"
	db "github.com/sofia/lab10/internal/sqlc"
)

var buildRecommender = defaultBuildRecommender
var openDBTX = defaultOpenDBTX
var newClothingAgent = func(apiKey, model string, repo llm.ClothingQuerier) (llm.Recommender, error) {
	return llm.NewClothingAgent(apiKey, model, repo)
}

func main() {
	ctx := context.Background()
	if err := runApp(ctx, os.Args[1:], os.Stdout); err != nil {
		slog.Error("command failed", "error", err)
		os.Exit(1)
	}
}

func runApp(ctx context.Context, args []string, out io.Writer) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	recService, cleanup, err := buildRecommender(ctx, cfg)
	if err != nil {
		return fmt.Errorf("build runtime: %w", err)
	}
	defer cleanup()

	return run(ctx, args, out, recService)
}

func defaultBuildRecommender(ctx context.Context, cfg config.Config) (llm.Recommender, func(), error) {
	conn, cleanup, err := openDBTX(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("connect postgres: %w", err)
	}

	queries := db.New(conn)
	repo := repository.NewClothesRepository(queries)

	agent, err := newClothingAgent(cfg.OpenRouterKey, cfg.OpenRouterModel, repo)
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("build clothing agent: %w", err)
	}

	return agent, cleanup, nil
}

func defaultOpenDBTX(ctx context.Context, databaseURL string) (db.DBTX, func(), error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, nil, err
	}
	return pool, pool.Close, nil
}

func run(ctx context.Context, args []string, out io.Writer, rec llm.Recommender) error {
	prompt := strings.TrimSpace(strings.Join(args, " "))
	if prompt == "" {
		return fmt.Errorf("prompt is required")
	}

	items, err := rec.Recommend(ctx, prompt)
	if err != nil {
		return fmt.Errorf("generate recommendations: %w", err)
	}

	payload, err := json.Marshal(items)
	if err != nil {
		return fmt.Errorf("marshal recommendations: %w", err)
	}

	if _, err := out.Write(payload); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}
