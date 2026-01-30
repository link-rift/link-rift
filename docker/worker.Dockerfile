# ── Build stage ──────────────────────────────
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /build/worker ./cmd/worker

# ── Runtime stage ────────────────────────────
FROM alpine:3.23

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S linkrift && adduser -S linkrift -G linkrift

COPY --from=builder /build/worker /usr/local/bin/worker

USER linkrift

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
    CMD test -f /tmp/worker-healthy || exit 1

ENTRYPOINT ["worker"]
