# Render Deployment Guide

## Architecture
Single Render **Web Service** running a Docker container that:
1. Builds the Vite frontend (bakes Auth0 env vars into the JS bundle)
2. Builds the Go backend
3. Runs DB migrations on startup
4. Serves the API and static frontend from one process

## Code Changes (already applied)
- [x] `main.go` uses `PORT` env var (Render sets this automatically)
- [x] `main.go` serves built frontend from `../frontend/dist`
- [x] `auth.go` reads `AUTH0_DOMAIN` from env (no more hardcoded domain)
- [x] `email_service.go` uses direct relative path for templates
- [x] `db.go` + `cmd/migrate` support `DATABASE_URL` (Render Postgres format)
- [x] `.gitignore` excludes all `.env` files
- [x] Multi-stage `Dockerfile` at project root
- [x] `.dockerignore` to keep build context lean

## Environment Variables Required on Render

### Set automatically by Render
- `PORT` — assigned by Render, no action needed

### Database (link your Render Postgres instance)
- `DATABASE_URL` — auto-populated when you link the DB to the service

### Auth0
- `AUTH0_DOMAIN` = `dev-cx0z71mw7lq3og41.us.auth0.com`

### Vite build args (set as env vars with "Build" visibility in Render)
- `VITE_AUTH0_CLIENT_DOMAIN` = `dev-cx0z71mw7lq3og41.us.auth0.com`
- `VITE_AUTH0_CLIENT_ID` = your Auth0 client ID

### AWS (S3 video uploads)
- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`
- `AWS_S3_BUCKET` = `nextpitch-videos`
- `AWS_REGION` = `us-east-2`

### Email (SMTP)
- `SMTP_HOST` = `smtp.zoho.com`
- `SMTP_PORT` = `587`
- `SMTP_USERNAME` = `coachtim@thenextpitch.org`
- `SMTP_PASSWORD`

## Deployment Steps

### 1. Push code to GitHub
```bash
git add -A
git commit -m "Add Dockerfile and production config for Render deployment"
git push origin render   # or main, whichever branch Render watches
```

### 2. Create Web Service on Render
- Go to https://dashboard.render.com → New → Web Service
- Connect your GitHub repo
- Select the branch (e.g. `render` or `main`)
- Runtime: **Docker**
- Render will auto-detect the `Dockerfile` at the repo root

### 3. Configure Environment Variables
- In the service's Environment tab, add all env vars listed above
- For `DATABASE_URL`: link your existing Render Postgres instance
- For `VITE_AUTH0_CLIENT_DOMAIN` and `VITE_AUTH0_CLIENT_ID`: set visibility to **Build** (they're baked into the frontend JS at build time, not needed at runtime)
- For all others: visibility can be **Runtime** or **Build & Runtime**

### 4. Update Auth0 Settings
In your Auth0 dashboard (https://manage.auth0.com):
- **Allowed Callback URLs**: add your Render URL (e.g. `https://your-app.onrender.com`)
- **Allowed Logout URLs**: add your Render URL
- **Allowed Web Origins**: add your Render URL
- **Allowed Origins (CORS)**: add your Render URL

### 5. Deploy
Render will automatically build and deploy when you push. You can also trigger a manual deploy from the Render dashboard.

### 6. Verify
- Visit `https://your-app.onrender.com` — should load the React app
- Visit `https://your-app.onrender.com/api/ping` — should return `{"message":"pong"}`
- Test login flow with Auth0
- Test scheduling, contact form, video upload

## Custom Domain (optional)
1. In Render service settings → Custom Domain
2. Add your domain (e.g. `app.thenextpitch.org`)
3. Update DNS records as instructed by Render
4. Update Auth0 URLs to include the custom domain
