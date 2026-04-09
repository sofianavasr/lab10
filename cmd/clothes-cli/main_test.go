package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sofia/lab10/internal/config"
	"github.com/sofia/lab10/internal/llm"
	db "github.com/sofia/lab10/internal/sqlc"
)

type recommenderStub struct {
	items []llm.ClothingItem
	err   error
}

func (r recommenderStub) Recommend(_ context.Context, _ string) ([]llm.ClothingItem, error) {
	return r.items, r.err
}

func TestRun(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		service recommenderStub
		wantErr string
	}{
		{
			name: "prints json output for prompt",
			args: []string{"casual outfit for day out"},
		service: recommenderStub{
			items: []llm.ClothingItem{
				{Name: "Tee", Category: "tops"},
				{Name: "Jeans", Category: "bottoms"},
				{Name: "Sneakers", Category: "shoes"},
			},
		},
		},
		{
			name:    "returns error when prompt missing",
			args:    nil,
			service: recommenderStub{},
			wantErr: "prompt is required",
		},
		{
			name: "returns error when service fails",
			args: []string{"casual outfit"},
			service: recommenderStub{
				err: errors.New("service failed"),
			},
			wantErr: "generate recommendations",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var out bytes.Buffer
			err := run(context.Background(), tt.args, &out, tt.service)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			raw := out.String()
			if idx := strings.Index(raw, "["); idx >= 0 {
				raw = raw[idx:]
			}
			var got []llm.ClothingItem
			if err := json.Unmarshal([]byte(raw), &got); err != nil {
				t.Fatalf("output is not valid JSON array: %v; output=%s", err, out.String())
			}
			if len(got) != 3 {
				t.Fatalf("expected 3 recommendations, got %d", len(got))
			}

			wantCategories := map[string]bool{
				"tops":    false,
				"bottoms": false,
				"shoes":   false,
			}
			for _, item := range got {
				if _, ok := wantCategories[item.Category]; ok {
					wantCategories[item.Category] = true
				}
			}
			for category, found := range wantCategories {
				if !found {
					t.Fatalf("missing required category %s in output %#v", category, got)
				}
			}
		})
	}
}

type failingWriter struct{}

func (failingWriter) Write(_ []byte) (int, error) {
	return 0, io.ErrClosedPipe
}

func TestRun_WriteError(t *testing.T) {
	t.Parallel()

	err := run(
		context.Background(),
		[]string{"casual"},
		failingWriter{},
		recommenderStub{
			items: []llm.ClothingItem{
				{Name: "Shirt", Category: "tops"},
				{Name: "Pants", Category: "bottoms"},
				{Name: "Sneakers", Category: "shoes"},
			},
		},
	)
	if err == nil || !strings.Contains(err.Error(), "write output") {
		t.Fatalf("expected write output error, got %v", err)
	}
}

func TestRunApp(t *testing.T) {
	t.Run("returns config error when env missing", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "")
		t.Setenv("OPENROUTER_API_KEY", "")
		t.Setenv("OPENROUTER_MODEL", "")

		var out bytes.Buffer
		err := runApp(context.Background(), []string{"casual"}, &out)
		if err == nil || !strings.Contains(err.Error(), "load config") {
			t.Fatalf("expected load config error, got %v", err)
		}
	})

	t.Run("uses buildRecommender and prints output", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "postgres://test")
		t.Setenv("OPENROUTER_API_KEY", "key")
		t.Setenv("OPENROUTER_MODEL", "openai/gpt-4o-mini")

		original := buildRecommender
		t.Cleanup(func() {
			buildRecommender = original
		})

		buildRecommender = func(_ context.Context, _ config.Config) (llm.Recommender, func(), error) {
			return recommenderStub{
					items: []llm.ClothingItem{
						{Name: "Shirt", Category: "tops"},
						{Name: "Pants", Category: "bottoms"},
						{Name: "Loafers", Category: "shoes"},
					},
				},
				func() {},
				nil
		}

		var out bytes.Buffer
		if err := runApp(context.Background(), []string{"formal office look"}, &out); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out.String(), `"tops"`) {
			t.Fatalf("unexpected output: %s", out.String())
		}
	})

	t.Run("returns build runtime error", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "postgres://test")
		t.Setenv("OPENROUTER_API_KEY", "key")
		t.Setenv("OPENROUTER_MODEL", "openai/gpt-4o-mini")

		original := buildRecommender
		t.Cleanup(func() {
			buildRecommender = original
		})

		buildRecommender = func(_ context.Context, _ config.Config) (llm.Recommender, func(), error) {
			return nil, nil, errors.New("build failed")
		}

		var out bytes.Buffer
		err := runApp(context.Background(), []string{"casual"}, &out)
		if err == nil || !strings.Contains(err.Error(), "build runtime") {
			t.Fatalf("expected build runtime error, got %v", err)
		}
	})

	t.Run("default build returns error on bad database url", func(t *testing.T) {
		_, cleanup, err := defaultBuildRecommender(context.Background(), config.Config{
			DatabaseURL:     "::invalid-dsn::",
			OpenRouterKey:   "key",
			OpenRouterModel: "openai/gpt-4o-mini",
		})
		if cleanup != nil {
			cleanup()
		}
		if err == nil || !strings.Contains(err.Error(), "connect postgres") {
			t.Fatalf("expected connect postgres error, got %v", err)
		}
	})

	t.Run("default build returns service with mocked dependencies", func(t *testing.T) {
		originalOpen := openDBTX
		originalAgent := newClothingAgent
		t.Cleanup(func() {
			openDBTX = originalOpen
			newClothingAgent = originalAgent
		})

		openDBTX = func(_ context.Context, _ string) (db.DBTX, func(), error) {
			return fakeDBTX{}, func() {}, nil
		}
		newClothingAgent = func(_, _ string, _ llm.ClothingQuerier, _ *http.Client) (llm.Recommender, error) {
			return fakeRecommender{}, nil
		}

		rec, cleanup, err := defaultBuildRecommender(context.Background(), config.Config{
			DatabaseURL:     "postgres://any",
			OpenRouterKey:   "key",
			OpenRouterModel: "openai/gpt-4o-mini",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cleanup == nil {
			t.Fatal("expected cleanup function")
		}
		cleanup()
		if rec == nil {
			t.Fatal("expected recommender service")
		}
	})
}

type fakeDBTX struct{}

func (fakeDBTX) Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (fakeDBTX) Query(context.Context, string, ...interface{}) (pgx.Rows, error) {
	return nil, errors.New("not implemented")
}

func (fakeDBTX) QueryRow(context.Context, string, ...interface{}) pgx.Row {
	return nil
}

type fakeRecommender struct{}

func (fakeRecommender) Recommend(_ context.Context, _ string) ([]llm.ClothingItem, error) {
	return []llm.ClothingItem{}, nil
}
