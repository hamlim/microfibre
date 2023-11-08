# Use the offical Golang image to build the app: https://hub.docker.com/_/golang
FROM golang:1.21.4 as builder

# Copy code to the image
WORKDIR /go/src/github.com/hamlim/microfibre
COPY . .

# Build the app
RUN CGO_ENABLED=0 GOOS=linux go build -v -o app

# Start a new image for production without build dependencies
FROM alpine
RUN apk add --no-cache ca-certificates

# Copy the app binary from the builder to the production image
COPY --from=builder /go/src/github.com/hamlim/microfibre/app /app

# Run the app when the vm starts
CMD ["/app"]
