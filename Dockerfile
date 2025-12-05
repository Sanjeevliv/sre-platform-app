# Build stage
FROM golang:1.25-alpine AS build

# Set working directory
WORKDIR /app

# Copy dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 go build -o /api-service ./cmd/api-service/main.go

# Final stage
FROM scratch

# Copy binary from build stage
COPY --from=build /api-service /api-service

# Expose port
EXPOSE 8080

# Start the application
ENTRYPOINT ["/api-service"]