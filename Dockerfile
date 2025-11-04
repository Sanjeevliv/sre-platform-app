# --- Build Stage ---
# Start with the official Go image.
# We use a specific version for reproducible builds.
FROM golang:1.25-alpine AS build

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the api-service binary
# CGO_ENABLED=0 is critical for static binaries
# -o /api-service specifies the output file name
RUN CGO_ENABLED=0 go build -o /api-service ./cmd/api-service/main.go

# --- Final Stage ---
# Start from a minimal 'scratch' image
# 'scratch' is an empty image, which is the most secure and smallest base
FROM scratch

# Copy the binary from the 'build' stage
COPY --from=build /api-service /api-service

# Expose the port the service runs on
EXPOSE 8080

# The command to run when the container starts
ENTRYPOINT ["/api-service"]