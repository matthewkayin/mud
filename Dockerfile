# Builds the game server and web client into one image. The editor is not included.

# Build the web client
FROM node:24-bookworm-slim AS frontend
WORKDIR /build/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Build the game server. go.work references the editor, which isn't copied, so it's turned off.
FROM golang:1.26-bookworm AS backend
WORKDIR /build/backend
ENV GOWORK=off
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN go build -o /build/mud .

# The server reads ./data and ./banner.txt and writes ./saves and ./logs, relative to /app
FROM debian:bookworm-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=backend /build/mud ./mud
COPY backend/data ./data
COPY backend/banner.txt ./banner.txt
COPY --from=frontend /build/frontend/dist ./static
ENV PORT=8080
ENV STATIC_FOLDER=/app/static
EXPOSE 8080
CMD ["/app/mud"]
