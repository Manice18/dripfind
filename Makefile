.PHONY: up down tools migrate migrate-up migrate-revert migrate-info migrate-new api worker web install tidy start stop

RUN_DIR := .run
MIGRATIONS_DIR := backend/migrations
POSTGRES_USER ?= outfit
POSTGRES_DB ?= outfitfinder

up:
	docker compose up -d postgres
	docker compose -f minio/docker-compose.yaml --project-directory minio up -d

down:
	docker compose down
	docker compose -f minio/docker-compose.yaml --project-directory minio down

tools:
	docker compose --profile tools up -d

# Apply pending *.up.sql via the same schema_migrations table the API uses.
# (Also applied automatically on API start when MIGRATE_ON_START=true.)
migrate: migrate-up

migrate-up:
	@echo "Applying pending migrations..."
	@docker compose exec -T postgres pg_isready -U $(POSTGRES_USER) -d $(POSTGRES_DB) >/dev/null \
		|| { echo "ERROR: Postgres is not ready. Run 'make up' first."; exit 1; }
	@docker compose exec -T postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 -c \
		"CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW());" >/dev/null
	@applied=0; \
	for f in $$(ls -1 $(MIGRATIONS_DIR)/*.up.sql 2>/dev/null | sort); do \
		version=$$(basename "$$f" .up.sql); \
		exists=$$(docker compose exec -T postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -Atc \
			"SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version='$$version')"); \
		if [ "$$exists" = "t" ]; then \
			echo "  skip  $$version"; \
			continue; \
		fi; \
		echo "  apply $$version"; \
		docker compose exec -T postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 < "$$f" \
			|| { echo "ERROR: failed applying $$f"; exit 1; }; \
		docker compose exec -T postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 -c \
			"INSERT INTO schema_migrations(version) VALUES('$$version');" >/dev/null \
			|| { echo "ERROR: failed recording $$version"; exit 1; }; \
		applied=$$((applied + 1)); \
	done; \
	if [ "$$applied" -eq 0 ]; then echo "Already up to date."; else echo "Applied $$applied migration(s)."; fi

migrate-revert:
	@echo "Reverting last migration..."
	@docker compose exec -T postgres pg_isready -U $(POSTGRES_USER) -d $(POSTGRES_DB) >/dev/null \
		|| { echo "ERROR: Postgres is not ready. Run 'make up' first."; exit 1; }
	@latest=$$(docker compose exec -T postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -Atc \
		"SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1;" 2>/dev/null); \
	if [ -z "$$latest" ]; then echo "No migrations to revert."; exit 1; fi; \
	down="$(MIGRATIONS_DIR)/$${latest}.down.sql"; \
	if [ ! -f "$$down" ]; then echo "ERROR: missing $$down"; exit 1; fi; \
	echo "  revert $$latest"; \
	docker compose exec -T postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 < "$$down" \
		|| { echo "ERROR: failed applying $$down"; exit 1; }; \
	docker compose exec -T postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 -c \
		"DELETE FROM schema_migrations WHERE version='$$latest';" >/dev/null; \
	echo "Reverted $$latest"

migrate-info:
	@echo "Checking migration status..."
	@docker compose exec -T postgres pg_isready -U $(POSTGRES_USER) -d $(POSTGRES_DB) >/dev/null \
		|| { echo "ERROR: Postgres is not ready. Run 'make up' first."; exit 1; }
	@echo ""
	@echo "Applied (schema_migrations):"
	@docker compose exec -T postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -c \
		"SELECT version, applied_at FROM schema_migrations ORDER BY version;" 2>/dev/null \
		|| echo "  (schema_migrations does not exist yet — run 'make migrate')"
	@echo ""
	@echo "Files in $(MIGRATIONS_DIR)/:"
	@ls -1 $(MIGRATIONS_DIR)/*.up.sql 2>/dev/null | xargs -n1 basename || echo "  (none)"

migrate-new:
	@echo "Creating new migration..."
	@echo ""
	@echo "Usage: make migrate-new NAME='description_of_change'"
	@echo ""
	@if [ -z "$(NAME)" ]; then \
		echo "ERROR: NAME is required"; \
		echo "Example: make migrate-new NAME='add_user_preferences'"; \
		exit 1; \
	fi
	@mkdir -p $(MIGRATIONS_DIR); \
	last=$$(ls -1 $(MIGRATIONS_DIR)/*.up.sql 2>/dev/null | sed 's|.*/||; s/_.*||' | sort -n | tail -1); \
	if [ -z "$$last" ]; then next=1; else next=$$((10#$$last + 1)); fi; \
	ver=$$(printf "%06d" $$next); \
	up="$(MIGRATIONS_DIR)/$${ver}_$(NAME).up.sql"; \
	down="$(MIGRATIONS_DIR)/$${ver}_$(NAME).down.sql"; \
	printf -- "-- +migrate Up\n\n" > "$$up"; \
	printf -- "-- +migrate Down\n\n" > "$$down"; \
	echo ""; \
	echo "Migration created:"; \
	echo "  $$up"; \
	echo "  $$down"; \
	echo ""; \
	echo "Next steps:"; \
	echo "  1. Edit the new migration files in $(MIGRATIONS_DIR)/"; \
	echo "  2. Run 'make migrate' to apply it locally"; \
	echo "  3. Test thoroughly before committing"; \
	echo ""; \
	echo "IMPORTANT: Never modify migrations after committing them!"

api:
	cd backend && go run ./cmd/api

worker:
	cd backend && go run ./cmd/worker

web:
	cd web && npm run dev

install:
	cd backend && go mod tidy
	cd web && npm install

tidy:
	cd backend && go mod tidy

# Start Postgres + MinIO + API + worker + web in the background.
start: up
	@mkdir -p $(RUN_DIR)
	@echo "Waiting for Postgres..."
	@for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30; do \
		docker compose exec -T postgres pg_isready >/dev/null 2>&1 && break; \
		sleep 1; \
	done
	@echo "Waiting for MinIO..."
	@for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30; do \
		curl -sf http://127.0.0.1:9000/minio/health/live >/dev/null 2>&1 && break; \
		sleep 1; \
	done
	@if lsof -ti tcp:8080 >/dev/null 2>&1; then \
		echo "API already running on :8080"; \
	else \
		sh -c 'cd backend || exit 1; nohup go run ./cmd/api >"$$1" 2>&1 & echo $$! >"$$2"' \
			_ $(CURDIR)/$(RUN_DIR)/api.log $(CURDIR)/$(RUN_DIR)/api.pid; \
		echo "API started → $(RUN_DIR)/api.log"; \
	fi
	@echo "Waiting for API..."
	@for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30; do \
		curl -sf http://127.0.0.1:8080/health >/dev/null 2>&1 && break; \
		sleep 1; \
	done
	@if [ -f $(RUN_DIR)/worker.pid ] && kill -0 $$(cat $(RUN_DIR)/worker.pid) 2>/dev/null; then \
		echo "Worker already running (pid $$(cat $(RUN_DIR)/worker.pid))"; \
	else \
		sh -c 'cd backend || exit 1; nohup go run ./cmd/worker >"$$1" 2>&1 & echo $$! >"$$2"' \
			_ $(CURDIR)/$(RUN_DIR)/worker.log $(CURDIR)/$(RUN_DIR)/worker.pid; \
		echo "Worker started → $(RUN_DIR)/worker.log"; \
	fi
	@if lsof -ti tcp:3000 >/dev/null 2>&1; then \
		echo "Web already running on :3000"; \
	else \
		sh -c 'cd web || exit 1; nohup npm run dev >"$$1" 2>&1 & echo $$! >"$$2"' \
			_ $(CURDIR)/$(RUN_DIR)/web.log $(CURDIR)/$(RUN_DIR)/web.pid; \
		echo "Web started → $(RUN_DIR)/web.log"; \
	fi
	@echo "LOOKBOOK: http://localhost:3000  ·  API: http://localhost:8080"

# Stop API, worker, web, and Postgres.
stop:
	@if [ -f $(RUN_DIR)/api.pid ]; then kill $$(cat $(RUN_DIR)/api.pid) 2>/dev/null || true; fi
	@if [ -f $(RUN_DIR)/worker.pid ]; then kill $$(cat $(RUN_DIR)/worker.pid) 2>/dev/null || true; fi
	@if [ -f $(RUN_DIR)/web.pid ]; then kill $$(cat $(RUN_DIR)/web.pid) 2>/dev/null || true; fi
	@-lsof -ti tcp:8080 | xargs kill 2>/dev/null || true
	@-lsof -ti tcp:3000 | xargs kill 2>/dev/null || true
	@-pkill -f 'go run ./cmd/worker' 2>/dev/null || true
	@-pkill -f 'backend/cmd/worker' 2>/dev/null || true
	@rm -f $(RUN_DIR)/api.pid $(RUN_DIR)/worker.pid $(RUN_DIR)/web.pid
	@docker compose down
	@docker compose -f minio/docker-compose.yaml --project-directory minio down
	@echo "Stopped Postgres, MinIO, API, worker, and web."
