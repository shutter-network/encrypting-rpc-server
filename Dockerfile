FROM golang:1.21-alpine3.17 as goDependencies

COPY ./src/go.mod /deps/go.mod
COPY ./src/go.sum /deps/go.sum
WORKDIR /deps
RUN go mod download

FROM ghcr.io/foundry-rs/foundry:latest as contracts
COPY gnosh-contracts /deps/gnosh-contracts
COPY .git /deps/.git
COPY src/Makefile /deps/src/Makefile
WORKDIR /deps/src
RUN apk add --no-cache make
RUN make compile-contracts

FROM golang:1.21-alpine3.17 as appBuilder
COPY src /src
COPY --from=contracts /deps/gnosh-contracts/out /gnosh-contracts/out
COPY --from=goDependencies /go /go
RUN apk add --no-cache make
WORKDIR /src
RUN mkdir /abis
RUN make build

FROM golang:1.21-alpine3.17
COPY --from=appBuilder /src/bin/encrypting-rpc-server /bin/encrypting-rpc-server
ENTRYPOINT ["encrypting-rpc-server", "start"]