# Dockerfile

# Use an official lightweight Go image as the build environment
FROM golang:1.19-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files to cache dependencies
COPY src/go.mod src/go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN go build -o websocket-server main.go

# Use a minimal base image to run the application
FROM alpine:latest

# Set the working directory
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/websocket-server .

# Expose the port that the server listens on
EXPOSE 6969

# Run the WebSocket server
CMD ["./websocket-server"]
