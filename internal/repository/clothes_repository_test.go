package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sofia/lab10/internal/llm"
	db "github.com/sofia/lab10/internal/sqlc"
)

type querierStub struct {
	cswItem db.Clothe
	cswErr  error
}

func (q querierStub) GetOneByCategoryStyleWeather(_ context.Context, _ db.GetOneByCategoryStyleWeatherParams) (db.Clothe, error) {
	return q.cswItem, q.cswErr
}

func TestMapQueryErr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want error
	}{
		{
			name: "maps pgx no rows to service not found",
			err:  pgx.ErrNoRows,
			want: llm.ErrNotFound,
		},
		{
			name: "wraps unknown errors",
			err:  errors.New("boom"),
			want: errors.New("query clothes: boom"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := mapQueryErr(tt.err)
			if tt.name == "maps pgx no rows to service not found" {
				if !errors.Is(got, tt.want) {
					t.Fatalf("expected mapped error %v, got %v", tt.want, got)
				}
				return
			}
			if got.Error() != tt.want.Error() {
				t.Fatalf("expected %q, got %q", tt.want.Error(), got.Error())
			}
		})
	}
}

func TestToServiceItem(t *testing.T) {
	t.Parallel()

	var price pgtype.Numeric
	if err := price.Scan("49.99"); err != nil {
		t.Fatalf("setup numeric: %v", err)
	}

	got, err := toServiceItem(db.Clothe{
		ID:       7,
		Name:     "Oxford Shirt",
		Price:    price,
		Color:    "white",
		Category: "tops",
		Style:    "business_casual",
		Weather:  "cold",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != 7 || got.Name != "Oxford Shirt" || got.Category != "tops" {
		t.Fatalf("unexpected mapping result: %#v", got)
	}
	if got.Price != 49.99 {
		t.Fatalf("unexpected price: %.2f", got.Price)
	}
}

func TestRepositoryMethods(t *testing.T) {
	t.Parallel()

	price := mustNumeric(t, "19.95")
	base := db.Clothe{
		ID:       10,
		Name:     "Test Item",
		Price:    price,
		Color:    "black",
		Category: "tops",
		Style:    "casual",
		Weather:  "hot",
	}

	t.Run("one by category style weather", func(t *testing.T) {
		t.Parallel()
		repo := NewClothesRepository(querierStub{cswItem: base})
		got, err := repo.OneByCategoryStyleWeather(t.Context(), "tops", "casual", "hot")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != 10 {
			t.Fatalf("unexpected item id: %d", got.ID)
		}
	})

	t.Run("maps no rows to not found in category-style-weather", func(t *testing.T) {
		t.Parallel()
		repo := NewClothesRepository(querierStub{cswErr: pgx.ErrNoRows})
		_, err := repo.OneByCategoryStyleWeather(t.Context(), "tops", "casual", "hot")
		if !errors.Is(err, llm.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

}

func mustNumeric(t *testing.T, value string) pgtype.Numeric {
	t.Helper()

	var price pgtype.Numeric
	if err := price.Scan(value); err != nil {
		t.Fatalf("setup numeric: %v", err)
	}
	return price
}
