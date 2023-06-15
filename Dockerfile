FROM golang:1.13-alpine as builder

# Add Maintainer Info
LABEL maintainer="Sam Zhou <sam@mixmedia.com>"

# Set the Current Working Directory inside the container
WORKDIR /app/mmfm-playback-go

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go version \
 && export GO111MODULE=on \
 && export GOPROXY=https://goproxy.cn,direct \
 && go mod vendor \
 && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mmfm-go

######## Start a new stage from scratch #######
FROM alpine:latest  

RUN apk add --no-cache tzdata dumb-init gettext-envsubst ffmpeg mplayer

WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/mmfm-playback-go/mmfm-go .
COPY ./config.json ./config.json

ENV TZ=Asia/Hong_Kong \
 SERVICE_NAME=mmfm-go \
 FFPLAY_BIN=/usr/bin/ffplay \
 FFPROBE_BIN=/usr/bin/ffprobe \
 MPLAYER_BIN=/usr/bin/mplayer \
 MMFM_HOST=192.168.33.6


ENTRYPOINT ["/usr/bin/dumb-init", "--"]

CMD envsubst < /app/config.json > /app/temp.json \
 && /app/mmfm-go -c /app/temp.json
