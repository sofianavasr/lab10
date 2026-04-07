package main

import (
	"math/rand"
	"testing"
)

func TestGenerateSeedRows(t *testing.T) {
	t.Parallel()

	rows, err := generateSeedRows(seedTotalRows, rand.New(rand.NewSource(42)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != seedTotalRows {
		t.Fatalf("expected %d rows, got %d", seedTotalRows, len(rows))
	}

	covered := make(map[string]struct{})
	for _, row := range rows {
		key := row.Category + "|" + row.Style + "|" + row.Weather
		covered[key] = struct{}{}
		if !isAllowedColor(row.Color) {
			t.Fatalf("row has invalid color %q", row.Color)
		}
	}

	if len(covered) != len(categories)*len(styles)*len(weathers) {
		t.Fatalf("expected %d unique category-style-weather combinations, got %d", len(categories)*len(styles)*len(weathers), len(covered))
	}
}

func TestGenerateSeedRows_ErrorsForTooSmallTotal(t *testing.T) {
	t.Parallel()

	_, err := generateSeedRows(80, rand.New(rand.NewSource(42)))
	if err == nil {
		t.Fatal("expected error when total rows is smaller than required combinations")
	}
}

func isAllowedColor(got string) bool {
	for _, color := range colors {
		if got == color {
			return true
		}
	}
	return false
}
