# Use the official Golang image as a build stage
FROM golang:1.24 as builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/main.go

# Use a minimal base image for the final image, but add ffmpeg via a temporary stage
FROM debian:12-slim as ffmpeg-installer
RUN apt-get update && apt-get install -y ffmpeg && rm -rf /var/lib/apt/lists/*

FROM gcr.io/distroless/base-debian12
WORKDIR /app

# Copy ffmpeg binary from installer stage
COPY --from=ffmpeg-installer /usr/bin/ffmpeg /usr/bin/ffmpeg
COPY --from=ffmpeg-installer /usr/bin/ffprobe /usr/bin/ffprobe

# Copy the built binary from the builder
COPY --from=builder /app/server ./server

# Expose the port (should match your config)
EXPOSE 8080

# Run the binary
CMD ["./server"]
