FROM golang:1.19 AS build

WORKDIR /build

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o client -a -ldflags="-s -w" ./cmd/client

FROM scratch

COPY --from=build /build/client /
COPY --from=build /build/config/dockercompose/client.yaml /config/client.yaml

ENTRYPOINT ["/client"]