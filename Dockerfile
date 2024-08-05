# Use the official Golang image as the base image
FROM golang:1.22.5-alpine

# Install necessary C libraries for SQLite3
RUN apk add --no-cache gcc musl-dev

# Set the Current Working Directory inside the container
WORKDIR /app

# Enable CGO
ENV CGO_ENABLED=1

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o forum ./cmd/web/

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./forum"]
