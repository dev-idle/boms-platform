.PHONY: help tidy build run dev test lint fmt vet clean docker-up docker-down sqlc sqlc-vet ci

APP_NAME := boms-api
MAIN_PKG := ./cmd/api
SQLC   := go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-18s\033[0m %s\n", $$1, $$2}'

tidy: ## go mod tidy + verify (run when dependencies change)
	go mod tidy
	go mod verify

build: ## Build API binary (fast — no tidy)
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/$(APP_NAME) $(MAIN_PKG)

run: ## Run API (requires .env)
	go run $(MAIN_PKG)

dev: ## Hot reload with Air
	air -c air.toml

test: ## Tests (no -race; CI uses -race on Linux)
	go test ./... -count=1

lint: ## golangci-lint (install separately)
	golangci-lint run ./...

fmt: ## go fmt
	go fmt ./...

vet: ## go vet
	go vet ./...

clean: ## Remove bin/ tmp/ coverage
	rm -rf bin/ tmp/ coverage.out

docker-up: ## Redis (Postgres = Neon in .env)
	docker compose -f scripts/docker-compose.dev.yml up -d

docker-down: ## Stop Redis
	docker compose -f scripts/docker-compose.dev.yml down

sqlc: ## sqlc generate
	$(SQLC) generate

sqlc-vet: ## sqlc vet
	$(SQLC) vet

ci: ## verify + sqlc vet + vet (tests: make test; CI runs -race once)
	go mod verify
	$(SQLC) vet
	go vet ./...
