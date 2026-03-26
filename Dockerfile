# syntax=docker/dockerfile:1

# ── Dev: Go toolchain + hot-reload tools (source mounted at /app) ─────────────
FROM golang:1.25-alpine AS dev
RUN apk add --no-cache ca-certificates tzdata git make && \
    go install github.com/air-verse/air@latest && \
    go install github.com/swaggo/swag/cmd/swag@latest
WORKDIR /app

# ── Stage 1: Build frontend ───────────────────────────────────────────────────
FROM node:22-alpine AS frontend
WORKDIR /build/web
COPY web/package*.json ./
RUN npm ci --ignore-scripts
COPY web/ ./
RUN npm run build
# Output: /build/web/dist/

# ── Stage 2: Build backend (with embedded frontend) ───────────────────────────
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Overlay the built frontend into the embed path before compiling.
COPY --from=frontend /build/web/dist ./internal/ui/dist
RUN CGO_ENABLED=0 GOOS=linux go build \
      -ldflags="-s -w" \
      -o /out/manhes \
      ./cmd/manhes

# ── Stage 3: Runtime ─────────────────────────────────────────────────────────
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /out/manhes /usr/local/bin/manhes
WORKDIR /app
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/manhes"]
