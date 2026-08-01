# syntax=docker/dockerfile:1

# ---------- build ----------
FROM golang:1.25-alpine AS build
WORKDIR /src

# Copy manifests first so dependency download is cached independently of source
# changes. Editing a .go file no longer re-downloads the module graph.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 produces a fully static binary, which is what lets the runtime
# stage be plain Alpine with no Go toolchain or libc shims.
# -s -w strips the symbol table and DWARF debug info.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/worker ./cmd/worker

# ---------- runtime ----------
FROM alpine:3.20

# ca-certificates is REQUIRED, not optional. Every OpenAI call and every retailer
# scrape is HTTPS; without a CA bundle they all fail with
# "x509: certificate signed by unknown authority".
# wget is used by the compose healthcheck.
RUN apk add --no-cache ca-certificates tzdata wget \
    && adduser -D -u 10001 app

WORKDIR /app
COPY --from=build /out/api /app/api
COPY --from=build /out/worker /app/worker

# REQUIRED: main.go's findMigrations() resolves ./migrations relative to the
# working directory. Without this the API exits at boot with a migrate error.
# NOTE: delete this line once UPDATE.md 5.2 embeds migrations via go:embed —
# keeping both copies invites schema drift.
COPY --from=build /src/migrations /app/migrations

# Must be owned by the non-root user or uploads fail with a permission error.
RUN mkdir -p /data/images && chown -R app /data/images

USER app

# Absolute path short-circuits the filepath.IsAbs check in main.go, so writes go
# straight to the mounted volume rather than the findBackendRoot() fallback.
ENV IMAGES_DIR=/data/images \
    HTTP_ADDR=:8080

EXPOSE 8080
CMD ["/app/api"]
