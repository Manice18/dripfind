.PHONY: up down tools migrate-up api web install tidy start stop

RUN_DIR := .run

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

# Start Postgres + API + web in the background.
start: up
	@mkdir -p $(RUN_DIR)
	@echo "Waiting for Postgres..."
	@for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30; do \
		docker compose exec -T postgres pg_isready >/dev/null 2>&1 && break; \
		sleep 1; \
	done
	@if lsof -ti tcp:8080 >/dev/null 2>&1; then \
		echo "API already running on :8080"; \
	else \
		(cd backend && go run ./cmd/api > ../$(RUN_DIR)/api.log 2>&1 & echo $$! > ../$(RUN_DIR)/api.pid); \
		echo "API started → $(RUN_DIR)/api.log"; \
	fi
	@if lsof -ti tcp:3000 >/dev/null 2>&1; then \
		echo "Web already running on :3000"; \
	else \
		(cd web && npm run dev > ../$(RUN_DIR)/web.log 2>&1 & echo $$! > ../$(RUN_DIR)/web.pid); \
		echo "Web started → $(RUN_DIR)/web.log"; \
	fi
	@echo "LOOKBOOK: http://localhost:3000  ·  API: http://localhost:8080"

# Stop API, web, and Postgres.
stop:
	@if [ -f $(RUN_DIR)/api.pid ]; then kill $$(cat $(RUN_DIR)/api.pid) 2>/dev/null || true; fi
	@if [ -f $(RUN_DIR)/web.pid ]; then kill $$(cat $(RUN_DIR)/web.pid) 2>/dev/null || true; fi
	@-lsof -ti tcp:8080 | xargs kill 2>/dev/null || true
	@-lsof -ti tcp:3000 | xargs kill 2>/dev/null || true
	@rm -f $(RUN_DIR)/api.pid $(RUN_DIR)/web.pid
	@docker compose down
	@echo "Stopped Postgres, API, and web."
