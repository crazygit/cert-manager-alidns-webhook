# syntax=docker/dockerfile:1.4

# Build stage
FROM golang:1.25.3 AS build

WORKDIR /workspace

# Download dependencies with BuildKit cache mount
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code
COPY main.go ./
COPY pkg/ ./pkg/

# Build the webhook binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a \
    -ldflags '-w -s -extldflags "-static"' \
    -trimpath \
    -o cert-manager-alidns-webhook .

# Final stage
FROM alpine:3.23

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -g 1000 cert-manager && \
    adduser -u 1000 -G cert-manager -D -h /home/cert-manager cert-manager

USER cert-manager

COPY --from=build /workspace/cert-manager-alidns-webhook /usr/local/bin/cert-manager-alidns-webhook

ENTRYPOINT ["/usr/local/bin/cert-manager-alidns-webhook"]
