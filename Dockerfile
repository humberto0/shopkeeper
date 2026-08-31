# syntax=docker/dockerfile:1

# --- build stage ---
FROM golang:1.27-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./

# Cache the module download separately from the source — re-downloaded only
# when go.mod/go.sum change. --mount=type=cache keeps the module cache across
# builds without committing it to the layer.
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

# --mount=type=cache keeps the Go build cache across rebuilds, cutting
# incremental build times significantly.
# CGO_ENABLED=0 → static binary (no C runtime in the final image).
# -trimpath strips local paths; -s -w drops symbol/debug tables.
# -buildvcs=false avoids a warning: .git is excluded via .dockerignore.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -buildvcs=false \
    -ldflags="-s -w" \
    -o /bin/api \
    ./cmd/api

# --- runtime stage ---
FROM alpine:3.21

# ca-certificates for outbound TLS; tzdata for time zone handling.
RUN apk add --no-cache ca-certificates tzdata

# Non-root user: a compromised process inherits an unprivileged account.
RUN adduser -D -u 10001 appuser
USER appuser

COPY --from=builder /bin/api /bin/api

EXPOSE 8080

ENTRYPOINT ["/bin/api"]
