.PHONY: dev build install test clean

GOOS=linux
GOARCH=amd64
CGO_ENABLED=0
GO111MODULE=off
BUILD_TIME=$(shell date +"%Y-%M-%dT%H:%M:%S" || echo "$BUILD_TIME")
COMMIT=$(shell git rev-parse --short HEAD || echo "$COMMIT")

all: dev

dev: build 
	@./main version

build:
	@go build -tags "cggo static_build" -installsuffix cggo \
		-ldflags "-w \
		-X $(shell GO111MODULE=off go list)/BuildTime=$(BUILD_TIME) \
		-X $(shell GO111MODULE=off go list)/Commit=$(COMMIT)" \
		./main.go

install: build
	@go install .

test:
	@go test -v -cover -race ./...

clean:
	@git clean -f -d -X
