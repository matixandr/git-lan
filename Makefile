BINARY  := git-lan
PKG     := github.com/matixandr/git-lan
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X $(PKG)/cmd.Version=$(VERSION)

# Cross-compile matrix: GOOS/GOARCH pairs shipped as release artifacts.
PLATFORMS := linux/amd64 darwin/amd64 darwin/arm64 windows/amd64

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

.PHONY: dist
dist: $(PLATFORMS)

# Pattern rule: `make linux/amd64` builds dist/git-lan-linux-amd64[.exe].
.PHONY: $(PLATFORMS)
$(PLATFORMS):
	@os=$(word 1,$(subst /, ,$@)); arch=$(word 2,$(subst /, ,$@)); \
	ext=""; [ "$$os" = "windows" ] && ext=".exe"; \
	out="dist/$(BINARY)-$$os-$$arch$$ext"; \
	echo "building $$out"; \
	GOOS=$$os GOARCH=$$arch go build -ldflags "$(LDFLAGS)" -o $$out .

.PHONY: clean
clean:
	rm -rf bin dist

.DEFAULT_GOAL := build
