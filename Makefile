## forge — backend fundamentals, from the metal up.
## Run `make help` to list targets.

.DEFAULT_GOAL := help
.PHONY: help up down reset migrate psql build run test image verify-image

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
	dotnet test

image: ## build the container image (paso-00b; needs your Dockerfile)
	docker build -t forge-api:paso-00b .

verify-image: ## paso-00b: run the executable spec for your Dockerfile + compose wiring
	./scripts/verify-image.sh
