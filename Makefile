BINARY  := git-lan
PKG     := github.com/matixandr/git-lan
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X $(PKG)/cmd.Version=$(VERSION)

.PHONY: build
build:
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) .

.PHONY: install
install:
	go install -ldflags "$(LDFLAGS)" .

.PHONY: test
test:
	go test ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: clean
clean:
	rm -rf bin dist

.DEFAULT_GOAL := build
