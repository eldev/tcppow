CURDIR=$(shell pwd)
BINDIR=${CURDIR}/bin
SERVER_PACKAGE=${CURDIR}/cmd/server
CLIENT_PACKAGE=${CURDIR}/cmd/client

.PHONY: test build-server build-client build-all run-server run-client run-docker-compose clean

test:
	go clean --testcache
	go test ./...

bindir:
	mkdir -p ${BINDIR}

build-server: bindir
	go build -o ${BINDIR}/server ${SERVER_PACKAGE}

build-client: bindir
	go build -o ${BINDIR}/client ${CLIENT_PACKAGE}

build-all: build-server build-client

run-server:
	SERVER_CONFIG_PATH=config/local/server.yaml go run cmd/server/server.go

run-client:
	CLIENT_CONFIG_PATH=config/local/client.yaml go run cmd/client/client.go

run-docker-compose:
	docker compose up --force-recreate --build

clean:
	rm -rf ${BINDIR}
	go clean --testcache