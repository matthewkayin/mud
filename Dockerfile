# Builds the game server and web client into one image. The editor is part of the web client and the server,
# but it's only enabled when the server runs with --edit, which the image's CMD doesn't pass.

# Build the web client. It's written to /build/backend/web/dist (see frontend/vite.config.ts).
FROM node:24-bookworm-slim AS frontend
WORKDIR /build/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Build the game server
FROM golang:1.26-bookworm AS backend
WORKDIR /build/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
COPY --from=frontend /build/backend/web/dist ./web/dist
# The embed tag builds the web client into the binary.
# The deploy workflow writes the release version to backend/VERSION. Builds without it report "dev".
RUN go build -tags embed -ldflags "-X main.version=$(cat VERSION 2>/dev/null || echo dev)" -o /build/mud .

# The server reads ./data and ./banner.txt and writes ./saves and ./logs, relative to /app
FROM debian:bookworm-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=backend /build/mud ./mud
COPY backend/data ./data
COPY backend/banner.txt ./banner.txt
ENV PORT=8080
EXPOSE 8080
CMD ["/app/mud"]
