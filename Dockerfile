# ---- Stage 1: Build the Vite frontend ----
FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci
COPY frontend/ ./

# Build args for Vite env vars (baked into the JS bundle at build time)
ARG VITE_AUTH0_CLIENT_DOMAIN
ARG VITE_AUTH0_CLIENT_ID
ARG VITE_STRIPE_PUBLISHABLE_KEY
ARG VITE_PAYPAL_CLIENT_ID

RUN npm run build

# ---- Stage 2: Build the Go backend ----
FROM golang:1.24-alpine AS backend-builder

WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./

# Build the main server binary
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# Build the migration tool
RUN CGO_ENABLED=0 GOOS=linux go build -o migrate ./cmd/migrate

# ---- Stage 3: Final lean image ----
FROM alpine:3.19

RUN apk add --no-cache ca-certificates

# Mirror the project directory structure so relative paths work:
#   Working dir: /app/backend
#   Frontend:    /app/frontend/dist  (referenced as ../frontend/dist)
#   Templates:   /app/backend/templates/email
#   Migrations:  /app/backend/db/migrations

WORKDIR /app/backend

# Copy Go binaries
COPY --from=backend-builder /app/backend/server .
COPY --from=backend-builder /app/backend/migrate .

# Copy email templates
COPY --from=backend-builder /app/backend/templates ./templates

# Copy migration files
COPY --from=backend-builder /app/backend/db/migrations ./db/migrations

# Copy built frontend
COPY --from=frontend-builder /app/frontend/dist /app/frontend/dist

# Run migrations then start the server
# Render sets the PORT env var automatically
EXPOSE 8080

CMD ["sh", "-c", "./migrate -action up && ./server"]
