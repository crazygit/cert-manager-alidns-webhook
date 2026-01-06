# Build stage
FROM golang:1.25.3 AS build

WORKDIR /workspace

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build the webhook binary
COPY . .
RUN CGO_ENABLED=0 go build \
    -ldflags '-w -extldflags "-static"' \
    -o cert-manager-alidns-webhook .

# Final stage
FROM alpine:3.23

RUN apk add --no-cache ca-certificates

# Non-root user
RUN addgroup -g 1000 cert-manager && \
    adduser -u 1000 -G cert-manager -D -h /home/cert-manager cert-manager

USER cert-manager

COPY --from=build /workspace/cert-manager-alidns-webhook /usr/local/bin/cert-manager-alidns-webhook

ENTRYPOINT ["/usr/local/bin/cert-manager-alidns-webhook"]
