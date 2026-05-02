# ── Build stage ──────────────────────────────
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /build/api ./cmd/api

# ── Runtime stage ────────────────────────────
FROM alpine:3.23

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S linkrift && adduser -S linkrift -G linkrift

COPY --from=builder /build/api /usr/local/bin/api

WORKDIR /app

RUN mkdir -p /app/data/uploads && chown -R linkrift:linkrift /app/data

USER linkrift

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["api"]
