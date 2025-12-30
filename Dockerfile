# Build stage
FROM golang:1.21-alpine AS builder

# Install git for go modules
RUN apk add --no-cache git

# Set the working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o mmfm-playback-go ./cmd/mmfm-playback

FROM scratch AS export
COPY --from=builder /app/mmfm-playback-go /mmfm-playback-go-armv7

# Final stage
FROM alpine:latest

# Install ffmpeg for audio playback
RUN apk --no-cache add ffmpeg mplayer

# Create a non-root user
RUN adduser -D -s /bin/sh appuser

# Create cache directory
RUN mkdir -p /home/appuser/cache
RUN chown -R appuser:appuser /home/appuser

# Copy the binary from builder stage
COPY --from=builder /app/mmfm-playback-go /usr/local/bin/mmfm-playback-go

# Copy the config file
COPY configs/config.json /app/config.json

ENV FFPLAY_PATH=/usr/bin/ffplay \
  FFMPEG_PATH=/usr/bin/ffmpeg \
  MPLAYER_PATH=/usr/bin/mplayer \
  CACHE_PATH=/tmp \
  WEBSOCKET_API=ws://localhost:8080/api/v1/ws

# Switch to non-root user
USER appuser

# Expose any necessary ports (if needed for API)
EXPOSE 8080

# Run the application
CMD ["mmfm-playback-go", "-c", "/app/config.json"]