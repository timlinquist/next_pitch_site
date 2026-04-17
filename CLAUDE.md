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

## Development
- Frontend dev server: `cd frontend && npm run dev`
- Backend: `cd backend && go run .`
- Build frontend: `cd frontend && npm run build`
- Run backend tests: `cd backend && go test ./...`
