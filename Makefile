.PHONY: up down migrate-up api web install tidy

up:
	docker compose up -d postgres

down:
	docker compose down

tools:
	docker compose --profile tools up -d

migrate-up:
	@echo "Migrations run automatically on API start (MIGRATE_ON_START=true)"

api:
	cd backend && go run ./cmd/api

web:
	cd web && npm run dev

install:
	cd backend && go mod tidy
	cd web && npm install

tidy:
	cd backend && go mod tidy

dev: up
	@echo "Postgres starting. Run 'make api' and 'make web' in separate terminals."
