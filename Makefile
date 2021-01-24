.PHONY: dev build install release test clean

GOOS=linux
CGO_ENABLED=0
GO111MODULE=off
VERSION=$(shell git describe --abbrev=0 --tags 2>/dev/null || echo "$VERSION")
COMMIT=$(shell git rev-parse --short HEAD || echo "$COMMIT")

all: dev

dev: build 
	@./main version

build:
	@go build -tags "cggo static_build" -installsuffix cggo \
		-ldflags "-w \
		-X $(shell go list)/Version=$(VERSION) \
		-X $(shell go list)/Commit=$(COMMIT)" \
		./main.go

install: build
	@go install .

test:
	@go test -v -cover -race ./...

clean:
	@git clean -f -d -X
