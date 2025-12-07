# Build Stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install git and make (for potential Makefile usage)
RUN apk add --no-cache git make

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build arguments for versioning
ARG VERSION=dev
ARG COMMIT_SHA=none
ARG BUILD_TIME=unknown

# Build the applications
# We build static binaries to ensure they work in distroless
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X github.com/sanjeevsethi/sre-platform-app/internal/metadata.Version=${VERSION} -X github.com/sanjeevsethi/sre-platform-app/internal/metadata.CommitSHA=${COMMIT_SHA} -X github.com/sanjeevsethi/sre-platform-app/internal/metadata.BuildTime=${BUILD_TIME}" -o /bin/api-service ./cmd/api-service
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X github.com/sanjeevsethi/sre-platform-app/internal/metadata.Version=${VERSION} -X github.com/sanjeevsethi/sre-platform-app/internal/metadata.CommitSHA=${COMMIT_SHA} -X github.com/sanjeevsethi/sre-platform-app/internal/metadata.BuildTime=${BUILD_TIME}" -o /bin/worker-service ./cmd/worker-service
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/healthcheck ./cmd/platform-healthcheck

# ------------------------------------------------------------------------------
# API Service Image
# ------------------------------------------------------------------------------
FROM gcr.io/distroless/static:nonroot AS api-service

WORKDIR /

# Copy binary and healthcheck
COPY --from=builder /bin/api-service /api-service
COPY --from=builder /bin/healthcheck /healthcheck

# Configuration
ENV API_PORT=8080
ENV GIN_MODE=release

# Expose port
EXPOSE 8080

# Healthcheck
HEALTHCHECK --interval=5s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/healthcheck", "http://localhost:8080/healthz"]

USER nonroot:nonroot

ENTRYPOINT ["/api-service"]

# ------------------------------------------------------------------------------
# Worker Service Image
# ------------------------------------------------------------------------------
FROM gcr.io/distroless/static:nonroot AS worker-service

WORKDIR /

# Copy binary and healthcheck
COPY --from=builder /bin/worker-service /worker-service
COPY --from=builder /bin/healthcheck /healthcheck

# Configuration
ENV WORKER_PORT=8081
ENV REDIS_ADDR=redis:6379

# Expose metrics port
EXPOSE 8081

# Healthcheck
HEALTHCHECK --interval=5s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/healthcheck", "http://localhost:8081/healthz"]

USER nonroot:nonroot

ENTRYPOINT ["/worker-service"]