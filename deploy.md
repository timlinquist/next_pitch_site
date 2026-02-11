# Render Deployment Checklist

- [ ] Fix `main.go` to use `PORT`
- [ ] Update static serving to use `frontend/dist` in production
- [ ] Replace hardcoded Auth0 domain in `auth.go` with `AUTH0_DOMAIN`
- [ ] Add `frontend/.env.example` with `VITE_AUTH0_CLIENT_DOMAIN`, `VITE_AUTH0_CLIENT_ID`, `VITE_API_URL`
- [ ] Create `render.yaml` (or configure equivalent in the dashboard)
- [ ] Add all env vars in Render
- [ ] Update Auth0 URLs for production
- [ ] Ensure `.env` is in `.gitignore`
- [ ] (Optional) Install Render CLI for local deploys and logs
