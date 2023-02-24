FROM golang:1.19 AS build

WORKDIR /build

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server -a -ldflags="-s -w" ./cmd/server

FROM scratch

COPY --from=build /build/server /
COPY --from=build /build/config/dockercompose/server.yaml /config/server.yaml

EXPOSE 8099

ENTRYPOINT ["/server"]