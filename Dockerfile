# syntax=docker/dockerfile:1
#
# Single Dockerfile at repo root (platform requirement).
# Builds the React SPA, builds the Go server, then assembles a slim runtime
# image that serves both (the Go server serves the built SPA in non-dev mode).
# Listens on port 80; health check at GET /up.

# ---------- Stage 1: build the frontend (React SPA) ----------
FROM node:24-alpine AS frontend-build
WORKDIR /app/frontend
# deps first (layer cache)
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
# SPA build (ssr:false) -> build/client
ENV NODE_ENV=production
RUN npm run build

# ---------- Stage 2: build the Go backend ----------
FROM golang:1.27-alpine AS backend-build
WORKDIR /app/backend
# cache go modules
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
# The SPA is served from frontend/dist relative to WORKDIR, so copy the built
# client tree into the expected location inside the build context.
COPY --from=frontend-build /app/frontend/build/client /app/frontend/dist
# static build, no CGO
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server ./cmd/server

# ---------- Stage 3: runtime ----------
# The app must bind port 80 (platform requirement). Binds to a privileged port
# require root, so the runtime image runs as root — standard for single-service
# app containers (the container runtime already isolates the process).
FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=backend-build /out/server /app/server
# SPA assets (relative to WORKDIR /app -> /app/frontend/dist)
COPY --from=backend-build /app/frontend/dist /app/frontend/dist
ENV PORT=80
EXPOSE 80
# Entrypoint: run migrations + serve (migrations are embedded in the binary
# via the migration runner; DATABASE_URL must be provided or the app fails hard).
ENTRYPOINT ["/app/server"]
