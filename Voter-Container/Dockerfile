FROM golang:1.21 AS build

# Set the Current Working Directory inside the container
WORKDIR /build_dir

# Copy the voter-api folder to inside the container
COPY /voter-api .

# Download all dependencies.
RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /voter-api

# Start a new stage from alpine
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /

# Copy the Pre-built binary file from the previous stage
COPY --from=build /voter-api /voter-api

# Expose port 8080 to the outside world
EXPOSE 8080

# Set the environment variable (We can also set host and port for the voter-api, but no need as the default, localhost:8080, is what we want)
ENV REDIS_URL=host.docker.internal:6379

# Run the voter-api binary
CMD ["/voter-api"]