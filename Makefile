## forge — backend fundamentals, from the metal up.
## Run `make help` to list targets.

.DEFAULT_GOAL := help
.PHONY: help up down reset migrate psql build run test image verify-image go-test go-race go-vet

help: ## list targets
	@grep -E '^[a-z-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-13s\033[0m %s\n", $$1, $$2}'

up: ## start Postgres (docker compose) and wait until healthy
	docker compose up -d
	@echo "waiting for db healthcheck…"; \
	until [ "$$(docker inspect -f '{{.State.Health.Status}}' forge-db 2>/dev/null)" = "healthy" ]; do sleep 1; done; \
	echo "✓ db healthy"

down: ## stop Postgres (keeps the data volume)
	docker compose down

reset: ## wipe the data volume, start fresh, and migrate
	docker compose down -v
	$(MAKE) up
	$(MAKE) migrate

migrate: ## apply db/migrations/*.sql
	./scripts/migrate.sh

psql: ## open a psql shell
	./scripts/psql.sh

build: ## build the solution
	dotnet build

run: ## run the API (needs: make up && make migrate)
	@if [ -f .env ]; then set -a; . ./.env; set +a; fi; dotnet run --project src/Forge.Api

test: ## run tests (integration tests skip if the db is down)
	@if [ -f .env ]; then set -a; . ./.env; set +a; fi; dotnet test

image: ## build the container image (paso-00b; needs your Dockerfile)
	docker build -t forge-api:paso-00b .

verify-image: ## paso-00b: run the executable spec for your Dockerfile + compose wiring
	./scripts/verify-image.sh

## ── Go line (paso-02 onward; see go/README.md) ──────────────────────────────
## Every target sources .env, so FORGE_DB_PORT points at forge's Postgres and
## not at whatever else is holding 5432 on this machine.

go-test: ## run the Go tests (integration tests skip if the db is down)
	@if [ -f .env ]; then set -a; . ./.env; set +a; fi; cd go && go test ./...

go-race: ## run the Go tests under the race detector (read paso-02 on what it does NOT catch)
	@if [ -f .env ]; then set -a; . ./.env; set +a; fi; cd go && go test -race -count=1 ./...

go-vet: ## vet + gofmt check
	@cd go && go vet ./... && test -z "$$(gofmt -l .)" && echo "✓ vet + gofmt clean"
