package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const seedTotalRows = 600

var colors = []string{"red", "blue", "pink", "brown", "black", "white", "gray", "beige"}
var categories = []string{"tops", "bottoms", "shoes"}
var styles = []string{"casual", "formal", "business_casual", "streetwear", "athleisure"}
var weathers = []string{"cold", "hot", "rainy", "snowy", "windy", "humid"}

type seedRow struct {
	Name     string
	Price    float64
	Color    string
	Category string
	Style    string
	Weather  string
}

type combo struct {
	Category string
	Style    string
	Weather  string
}

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		slog.Error("seed failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rows, err := generateSeedRows(seedTotalRows, rng)
	if err != nil {
		return fmt.Errorf("generate seed rows: %w", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	for _, row := range rows {
		_, err := tx.Exec(ctx, `
			INSERT INTO clothes (name, price, color, category, style, weather)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, row.Name, row.Price, row.Color, row.Category, row.Style, row.Weather)
		if err != nil {
			return fmt.Errorf("insert row %q: %w", row.Name, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	slog.Info("seed completed", "rows", len(rows))
	return nil
}

func generateSeedRows(total int, rng *rand.Rand) ([]seedRow, error) {
	combos := allCombos()
	if total < len(combos) {
		return nil, fmt.Errorf("total rows must be at least %d", len(combos))
	}

	rng.Shuffle(len(combos), func(i, j int) {
		combos[i], combos[j] = combos[j], combos[i]
	})

	rows := make([]seedRow, 0, total)
	for idx, c := range combos {
		rows = append(rows, newSeedRow(idx+1, c, rng))
	}

	for len(rows) < total {
		c := combos[rng.Intn(len(combos))]
		rows = append(rows, newSeedRow(len(rows)+1, c, rng))
	}

	return rows, nil
}

func allCombos() []combo {
	combos := make([]combo, 0, len(categories)*len(styles)*len(weathers))
	for _, category := range categories {
		for _, style := range styles {
			for _, weather := range weathers {
				combos = append(combos, combo{
					Category: category,
					Style:    style,
					Weather:  weather,
				})
			}
		}
	}
	return combos
}

func newSeedRow(index int, c combo, rng *rand.Rand) seedRow {
	price := 15.0 + rng.Float64()*145.0
	return seedRow{
		Name:     fmt.Sprintf("%s-%s-%s-%03d", c.Category, c.Style, c.Weather, index),
		Price:    float64(int(price*100)) / 100,
		Color:    colors[rng.Intn(len(colors))],
		Category: c.Category,
		Style:    c.Style,
		Weather:  c.Weather,
	}
}
