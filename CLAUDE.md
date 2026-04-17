# Next Pitch Site

## Project Structure
- `frontend/` — React + Vite SPA
- `backend/` — Go (Gin) API server
- Backend serves the built frontend from `frontend/dist/`

## Git

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

Local dev uses the Vite dev server (HMR + API proxy to Go backend):

1. `cd backend && go run .` — starts API on :8080
2. `cd frontend && npm run dev` — starts Vite on :5173, proxies `/api` to :8080
3. Open `localhost:5173`

`npm run build` is for production only (output to `frontend/dist/`, served by Go).

- Run backend tests: `cd backend && go test ./...`
- Run frontend tests: `cd frontend && npm run test`
