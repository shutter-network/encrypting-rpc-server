# Stage 1: Build Go Dependencies
FROM golang:1.21-alpine3.17 as goDependencies

COPY ./src/go.mod /app/go.mod
COPY ./src/go.sum /app/go.sum
WORKDIR /app
RUN go mod download

# Stage 2: Build the Go Application
FROM golang:1.21-alpine3.17 as appBuilder

RUN apk add --no-cache gcc g++ musl-dev
COPY ./src /app
COPY --from=goDependencies /go /go
WORKDIR /app
RUN mkdir /abis
RUN CGO_ENABLED=1 GOOS=linux go build -o /bin/encrypting-rpc-server

# Stage 3: Final Image
FROM alpine:3.17

# Copy the Pre-built binary file from the previous stage
COPY --from=appBuilder /bin/encrypting-rpc-server /bin/encrypting-rpc-server

# Command to run the executable
ENTRYPOINT ["/bin/encrypting-rpc-server", "start"]