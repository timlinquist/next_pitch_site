# Next Pitch Site

## Project Structure
- `frontend/` — React + Vite SPA
- `backend/` — Go (Gin) API server
- Backend serves the built frontend from `frontend/dist/`

## Git

**Before any `git push`**, always run `/credential-scan` first. Do not push if the scan finds issues.

The remote origin uses a custom SSH alias (`git@github-personal:`) which may not resolve depending on local SSH config. When pushing fails over SSH, use the gh CLI HTTPS workaround:

```sh
gh auth setup-git && git -c url."https://github.com/".insteadOf="git@github-personal:" push -u origin <branch>
```

## Node / nvm

Node is managed via nvm. The Bash tool doesn't source `.zshrc`, so prefix npm/node commands with:

```sh
export NVM_DIR="$HOME/.nvm" && . "$NVM_DIR/nvm.sh" && <command>
```

## Development

To start the full local dev environment (runs migrations, then backend + frontend):

```sh
make dev:start
```

This runs migrations, starts the Go API on :8080, and Vite on :5173 (proxies `/api` to backend). Open `localhost:5173`. Ctrl-C stops both.

Other useful targets:
- `make migrate` — run DB migrations only
- `make test` — run backend + frontend unit tests
- `make test-backend` / `make test-frontend` — run unit tests individually
- `make test-e2e` — run Playwright E2E tests (auto-starts servers, runs migrations, cleans up test data)
- `make test-all` — run all tests (backend + frontend unit + E2E)
- `make build` — production build (backend binary + frontend dist)

## Testing

**After completing any feature or bug fix**, always run `make test-all` before considering the work done. This runs backend unit tests, frontend unit tests, and Playwright E2E tests. The E2E suite handles its own setup (migrations, server startup) and teardown (test data cleanup) automatically.

Stop the dev server (`make dev:stop`) before running E2E tests — the test runner starts its own servers.

`npm run build` is for production only (output to `frontend/dist/`, served by Go).
