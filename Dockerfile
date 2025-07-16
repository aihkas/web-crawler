# Stage 1: Build the Go binary
FROM golang:1.22-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source code
COPY . .

# Build the application, disabling CGO for a static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Stage 2: Create the final, minimal image
FROM alpine:latest

# It's good practice to run as a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

WORKDIR /home/appuser/

# Copy the built binary from the builder stage
COPY --from=builder /app/main .

# Expose the port the server runs on
EXPOSE 8080

# The command to run the application
CMD ["./main"]
