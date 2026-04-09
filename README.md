# Lab10 Clothes Recommender (Go + Postgres + LangChainGo)

CLI application that recommends an outfit from a PostgreSQL `clothes` table using natural-language input. The LLM can call a **weather tool** backed by [Open-Meteo](https://open-meteo.com/) (geocoding + current conditions) so recommendations can match real conditions for a named city when you mention one.

The app always returns exactly 3 clothing items:
- 1 `tops`
- 1 `bottoms`
- 1 `shoes`

## Stack

- Go 1.22+
- PostgreSQL (pgx)
- sqlc for type-safe queries
- LangChainGo with a direct OpenRouter-backed LLM parser

## Project structure

- `cmd/clothes-cli`: CLI entrypoint
- `internal/service`: recommendation business logic
- `internal/repository`: database access layer
- `internal/llm`: OpenRouter intent parsing adapter
- `sqlc/queries`: SQL query definitions
- `internal/sqlc`: generated sqlc code
- `migrations`: schema migrations

## Setup

1. Copy env vars:
   - `cp .env.example .env`
2. Fill required values in `.env`:
   - `DATABASE_URL`
   - `OPENROUTER_API_KEY`
   - `OPENROUTER_MODEL`
3. Start Postgres:
   - `make docker-up`
4. Apply migrations:
   - `make migrate-up`
5. Generate sqlc code (after editing queries):
   - `make sqlc`
6. Seed sample data (600 rows):
   - `make seed`

## Run

```bash
make run ARGS="I want to wear a casual outfit for a day out in cdmx"
```

Output is a JSON array with exactly 3 items (`tops`, `bottoms`, `shoes`).

## Weather API usage

The clothing agent exposes a `get_weather` tool that resolves a city to coordinates and reads current temperature, WMO weather code, and wind speed from Open-Meteo’s public endpoints (no API key; see [Open-Meteo API](https://open-meteo.com/en/docs)). The tool returns a short line such as `weather: rainy, temperature: 18.5°C`. The model maps that to one of the database `weather` values (`cold`, `hot`, `rainy`, `snowy`, `windy`, `humid`) when calling `search_clothing`.

**Requirements:** outbound HTTPS to `geocoding-api.open-meteo.com` and `api.open-meteo.com` while the CLI runs.

**Example prompts** (include a place so the model can fetch conditions):

```bash
make run ARGS="Casual outfit for today in Mexico City"
make run ARGS="What should I wear to work in Oslo this afternoon?"
```

If the prompt does not imply a location, the model may skip the weather tool and pick a `weather` value from context alone.

## Development commands

- `make fmt` - format code
- `go vet ./...` - vet checks
- `make lint` - golangci-lint
- `make staticcheck` - staticcheck
- `go test ./...` - run unit tests
- `go test ./cmd/clothes-cli ./cmd/seed-clothes ./internal/config ./internal/llm ./internal/repository ./internal/service -coverprofile=coverage.out`
- `make run ARGS="..."` - run the CLI with a prompt
- `make seed` - insert 600 sample clothes rows

## Database notes

The `clothes` table is created in `migrations/001_create_clothes_table.up.sql` with constrained values for:
- `color`
- `category`
- `style`
- `weather`

The seed command (`go run ./cmd/seed-clothes` or `make seed`) inserts 600 rows and guarantees every `(category, style, weather)` combination appears at least once (90 unique triples total), while assigning random valid colors.

Running the seed command multiple times appends more rows. To reseed from scratch, clear the table first:

```bash
psql "$DATABASE_URL" -c "TRUNCATE TABLE clothes RESTART IDENTITY;"
make seed
```

Use `make migrate-down` to rollback one migration.
