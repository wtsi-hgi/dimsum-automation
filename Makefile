PKG := github.com/wtsi-hgi/dimsum-automation
VERSION := $(shell git describe --tags --always --long --dirty)
TAG := $(shell git describe --abbrev=0 --tags)
LDFLAGS = -ldflags "-X ${PKG}/cmd.Version=${VERSION}"
export GOPATH := $(shell go env GOPATH)
PATH := ${PATH}:${GOPATH}/bin

default: install

build: export CGO_ENABLED = 0
build:
	go build -tags netgo ${LDFLAGS}

install: export CGO_ENABLED = 0
install:
	@rm -f ${GOPATH}/bin/dimsum-automation
	@go install -tags netgo ${LDFLAGS}
	@echo installed to ${GOPATH}/bin/dimsum-automation

test: export CGO_ENABLED = 1
test:
	@go test -tags netgo --count 1 .
	@go test -tags netgo --count 1 $(shell go list ./... | tail -n+2)

race: export CGO_ENABLED = 1
race:
	@go test -tags netgo -race --count 1 .
	@go test -tags netgo -race --count 1 $(shell go list ./... | tail -n+2)

bench: export CGO_ENABLED = 1
bench:
	go test -tags netgo --count 1 -run Bench -bench=. ./...

# curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.59.1
lint:
	@golangci-lint run

clean:
	@rm -f ./dimsum-automation

.PHONY: test race bench lint build install clean
