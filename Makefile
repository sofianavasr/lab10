-include .env
export

APP_NAME=clothes-cli
MIGRATIONS_DIR=./migrations

.PHONY: fmt vet test coverage lint staticcheck check sqlc migrate-up migrate-down migrate-create docker-up docker-down

fmt:
	gofmt -w $$(go list -f '{{.Dir}}' ./...)

vet:
	go vet ./...

test:
	go test ./...

coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

lint:
	golangci-lint run ./...

staticcheck:
	staticcheck ./...

check: fmt vet lint staticcheck test

sqlc:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$$DATABASE_URL" up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$$DATABASE_URL" down 1

migrate-create:
	@if [ -z "$(name)" ]; then echo "usage: make migrate-create name=your_name"; exit 1; fi
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

docker-up:
	docker compose up -d

docker-down:
	docker compose down
