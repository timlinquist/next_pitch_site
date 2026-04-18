---
name: deploy
description: Deploy the Next Pitch site to Render (build, push, verify)
user_invocable: true
---

# Deploy to Render

Deploy the Next Pitch site. The app runs as a single Render Web Service using Docker: Vite frontend build + Go backend + DB migrations on startup.

- **Repo**: timlinquist/next_pitch_site
- **Deploy branch**: `main`
- **Service ID**: `srv-d65uhnrnv86c73e04mk0`
- **Workspace**: `tea-d65ttqsr85hc73d1uld0` (My Workspace)

## Pre-deploy checks

1. Ensure backend compiles: `cd backend && go build ./...`
2. Ensure backend tests pass: `cd backend && go test ./...`
3. Ensure frontend builds: `cd frontend && npm run build`
4. Confirm all changes are committed: `git status`

## Push

Push to the `render` branch on GitHub. The remote uses a custom SSH alias that may not resolve — fall back to HTTPS via gh CLI:

```sh
git push origin main
# If SSH fails:
gh auth setup-git && git -c url."https://github.com/".insteadOf="git@github-personal:" push -u origin main
```

Render auto-deploys when the `main` branch is pushed. To trigger a manual deploy instead, use the Render CLI (see below).

## Render CLI

Before running ANY `render` CLI command, read `~/.render/cli.yaml` and verify:
1. `workspace` field exists and is non-empty
2. `api.key` field exists and is non-empty
3. `api.expires_at` is a Unix timestamp in the future (`date +%s` to compare)

If any check fails, tell the user to run `render login` and/or `render workspace set` in their terminal.

Common commands:
```sh
render services list
render deploys list --service srv-d65uhnrnv86c73e04mk0
render logs --service srv-d65uhnrnv86c73e04mk0
render deploys create --service srv-d65uhnrnv86c73e04mk0   # manual deploy
```

When debugging build/deploy failures, use `render logs` or `render deploys list` to fetch logs directly.

## Architecture

- Multi-stage `Dockerfile` at project root (Node build → Go build → Alpine runtime)
- Vite Auth0 env vars (`VITE_AUTH0_CLIENT_DOMAIN`, `VITE_AUTH0_CLIENT_ID`) are baked into the JS bundle at build time (set as **Build** visibility on Render)
- Go binary runs DB migrations on startup (`./migrate -action up`), then serves API + static frontend
- `PORT` is set automatically by Render
- Docker container working directory is `/app/backend` — all runtime paths are relative to that:
  - Frontend dist: `../frontend/dist`
  - Email templates: `templates/email/`
  - DB migrations: `db/migrations/`

## Required Render env vars

| Variable | Visibility | Notes |
|---|---|---|
| `DATABASE_URL` | Runtime | Auto-populated when Render Postgres is linked |
| `PORT` | Runtime | Auto-set by Render |
| `AUTH0_DOMAIN` | Runtime | `dev-cx0z71mw7lq3og41.us.auth0.com` |
| `VITE_AUTH0_CLIENT_DOMAIN` | Build | Same as AUTH0_DOMAIN |
| `VITE_AUTH0_CLIENT_ID` | Build | Auth0 client ID |
| `AWS_ACCESS_KEY_ID` | Runtime | S3 video uploads |
| `AWS_SECRET_ACCESS_KEY` | Runtime | S3 video uploads |
| `AWS_S3_BUCKET` | Runtime | `nextpitch-videos` |
| `AWS_REGION` | Runtime | `us-east-2` |
| `SMTP_HOST` | Runtime | `smtp.zoho.com` |
| `SMTP_PORT` | Runtime | `587` |
| `SMTP_USERNAME` | Runtime | `coachtim@thenextpitch.org` |
| `SMTP_PASSWORD` | Runtime | Zoho app password |
| `STRIPE_SECRET_KEY` | Runtime | Stripe secret key |
| `STRIPE_WEBHOOK_SECRET` | Runtime | Stripe webhook signing secret |
| `PAYPAL_CLIENT_ID` | Runtime | PayPal client ID |
| `PAYPAL_CLIENT_SECRET` | Runtime | PayPal client secret |
| `PAYPAL_API_BASE` | Runtime | `https://api-m.paypal.com` for prod |
| `VITE_STRIPE_PUBLISHABLE_KEY` | Build | Stripe publishable key |
| `VITE_PAYPAL_CLIENT_ID` | Build | PayPal client ID |

## Post-deploy verification

1. `https://<app>.onrender.com` — React app loads
2. `https://<app>.onrender.com/api/ping` — returns `{"message":"pong"}`
3. Test Auth0 login flow
4. Test camp listing at `/camps`

## Auth0 URL updates (only needed for new domains)

In Auth0 dashboard, add the Render/custom domain URL to:
- Allowed Callback URLs
- Allowed Logout URLs
- Allowed Web Origins
- Allowed Origins (CORS)
