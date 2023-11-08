# Use the offical Golang image to build the app: https://hub.docker.com/_/golang
FROM golang:1.21.4 as builder

# Copy code to the image
WORKDIR /go/src/github.com/hamlim/microfibre
COPY . .

ENV DB_FILE_PATH=/litefs/microfibre.db
ENV GIN_MODE=release

# Build the app
RUN CGO_ENABLED=0 GOOS=linux go build -v -o app

# Start a new image for production without build dependencies
FROM alpine
RUN apk add --no-cache ca-certificates fuse3 sqlite

# Copy the app binary from the builder to the production image
COPY --from=builder /go/src/github.com/hamlim/microfibre/app /app
# Copy litefs binary
COPY --from=flyio/litefs:0.5 /usr/local/bin/litefs /usr/local/bin/litefs

# Run the app when the vm starts
CMD ["litefs mount"]
