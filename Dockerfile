# Use the official Go image as the base image
FROM golang:1.25.1-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Install git and other necessary tools
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o git-backup .

# Use a minimal Alpine image for the final stage
FROM alpine:latest

# Install git and ca-certificates
RUN apk --no-cache add git ca-certificates

# Create a non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set the working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/git-backup .

# Create the repos directory and set ownership
RUN mkdir -p repos && chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Set environment variables with default values
ENV GITHUB_TOKENS=""
ENV GITLAB_TOKENS=""
ENV REPOS_DIR="/repos"
ENV REPEAT_INTERVAL=-1

# Set the default command
CMD ["./git-backup"]
