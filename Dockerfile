# First stage: Build the Go binary
FROM golang:1.22 AS builder

# Set the working directory inside the builder container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code (assuming your main.go is in cmd/)
COPY ./cmd ./cmd

# Build the Go binary
RUN go build -o /app/main ./cmd/main.go

# Second stage: Create a smaller final image
FROM debian:bullseye-slim

# Set the working directory inside the final image
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Expose any ports (if needed), for example:
# EXPOSE 8080

# Command to run the binary
CMD ["./main"]

