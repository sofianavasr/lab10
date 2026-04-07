package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sofia/lab10/internal/llm"
	db "github.com/sofia/lab10/internal/sqlc"
)

type ClothesRepository struct {
	queries clothesQuerier
}

type clothesQuerier interface {
	GetOneByCategoryStyleWeather(ctx context.Context, arg db.GetOneByCategoryStyleWeatherParams) (db.Clothe, error)
}

func NewClothesRepository(queries clothesQuerier) *ClothesRepository {
	return &ClothesRepository{queries: queries}
}

func (r *ClothesRepository) OneByCategoryStyleWeather(
	ctx context.Context,
	category, style, weather string,
) (llm.ClothingItem, error) {
	item, err := r.queries.GetOneByCategoryStyleWeather(ctx, db.GetOneByCategoryStyleWeatherParams{
		Category: category,
		Style:    style,
		Weather:  weather,
	})
	if err != nil {
		return llm.ClothingItem{}, mapQueryErr(err)
	}
	return toServiceItem(item)
}


func mapQueryErr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return llm.ErrNotFound
	}
	return fmt.Errorf("query clothes: %w", err)
}

func toServiceItem(item db.Clothe) (llm.ClothingItem, error) {
	price, err := numericToFloat(item.Price)
	if err != nil {
		return llm.ClothingItem{}, fmt.Errorf("convert price: %w", err)
	}
	return llm.ClothingItem{
		ID:       item.ID,
		Name:     item.Name,
		Price:    price,
		Color:    item.Color,
		Category: item.Category,
		Style:    item.Style,
		Weather:  item.Weather,
	}, nil
}

func numericToFloat(n pgtype.Numeric) (float64, error) {
	v, err := n.Float64Value()
	if err != nil {
		return 0, err
	}
	return v.Float64, nil
}
