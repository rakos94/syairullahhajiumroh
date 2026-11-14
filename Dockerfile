# syntax=docker/dockerfile:1

# 1) Build the React frontend
FROM node:22-alpine AS web
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY web/ ./
RUN npm run build

# 2) Build the Go server (embeds web/dist)
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /src/web/dist ./web/dist
# -p 1 compiles one package at a time so the build fits in ~2GB RAM (small VPS / Docker VM).
RUN CGO_ENABLED=0 go build -p 1 -trimpath -ldflags="-s -w" -o /out/server ./cmd/main.go

# 3) Small runtime image
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget \
    && adduser -D -H -u 10001 app \
    && mkdir -p /app/uploads && chown app:app /app/uploads
WORKDIR /app
COPY --from=build /out/server /app/server
USER app
ENV GIN_MODE=release \
    APP_PORT=8080 \
    UPLOAD_DIR=/app/uploads \
    TZ=Asia/Jakarta
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/api/health || exit 1
CMD ["/app/server"]
