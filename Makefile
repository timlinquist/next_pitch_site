SHELL := /bin/bash
.PHONY: dev\:start dev\:stop migrate backend frontend test test-backend test-frontend build

NVM_INIT := export NVM_DIR="$$HOME/.nvm" && . "$$NVM_DIR/nvm.sh"

dev\:start: migrate
	@echo "Starting backend and frontend..."
	@trap 'kill 0' EXIT; \
	(cd backend && go run .) & \
	($(NVM_INIT) && cd frontend && npm run dev) & \
	wait

dev\:stop:
	@for port in 8080 5173; do \
		pids=$$(lsof -ti:$$port 2>/dev/null); \
		if [ -n "$$pids" ]; then \
			echo "Stopping processes on port $$port (PIDs: $$pids)"; \
			echo "$$pids" | xargs kill 2>/dev/null; \
		fi; \
	done
	@sleep 1
	@all_clear=true; \
	for port in 8080 5173; do \
		if lsof -ti:$$port >/dev/null 2>&1; then \
			echo "WARNING: port $$port still in use"; \
			all_clear=false; \
		fi; \
	done; \
	if $$all_clear; then echo "Ports 8080 and 5173 are free"; fi

migrate:
	cd backend && go run cmd/migrate/main.go -action up

backend:
	cd backend && go run .

frontend:
	$(NVM_INIT) && cd frontend && npm run dev

test: test-backend test-frontend

test-backend:
	cd backend && go test ./...

test-frontend:
	$(NVM_INIT) && cd frontend && npm run test

build:
	cd backend && go build ./...
	$(NVM_INIT) && cd frontend && npm run build
