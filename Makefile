# iptv — BusinessRepo Makefile. CI/CD and humans both call these targets.
BINARY      := iptv
PKG         := github.com/kinncj/iptv
BUILD_DIR   := .
GO          ?= go
GOFLAGS     ?=
LDFLAGS     ?= -s -w

.DEFAULT_GOAL := help

## help: list available targets
.PHONY: help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | awk -F: '{printf "  \033[36m%-16s\033[0m%s\n", $$1, $$2}'

## deps: download and tidy modules
.PHONY: deps
deps:
	$(GO) mod tidy

## build-tui: build the TUI binary into ./iptv
.PHONY: build-tui
build-tui:
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) ./app

## build: alias for build-tui
.PHONY: build
build: build-tui

## run-tui: build and run the TUI
.PHONY: run-tui
run-tui: build-tui
	./$(BINARY)

## run: alias for run-tui
.PHONY: run
run: run-tui

## refresh: run the TUI forcing a playlist re-download
.PHONY: refresh
refresh: build-tui
	./$(BINARY) -refresh

## rebuild: regenerate ./playlists (all.m3u + countries/) from the upstream repos
.PHONY: rebuild
rebuild: build-tui
	./$(BINARY) -export playlists

## test: run unit tests
.PHONY: test
test:
	$(GO) test ./... -count=1

## test-race: run unit tests with the race detector
.PHONY: test-race
test-race:
	$(GO) test ./... -race -count=1

## test-acceptance: run acceptance/feature tests
.PHONY: test-acceptance
test-acceptance:
	$(GO) test ./tests/... -count=1

## lint: vet and format-check the codebase
.PHONY: lint
lint:
	$(GO) vet ./...
	@test -z "$$(gofmt -l app common tests)" || (echo "gofmt needed:"; gofmt -l app common tests; exit 1)

## ci: the full gate CI runs — lint, race tests, and build
.PHONY: ci
ci: lint test-race build-tui

## fmt: format the codebase
.PHONY: fmt
fmt:
	gofmt -w app common tests

## install: install the binary into GOBIN/PATH
.PHONY: install
install:
	$(GO) install $(GOFLAGS) -ldflags "$(LDFLAGS)" ./app

## release: cross-compile release binaries + man page into dist/ (VERSION=x.y.z)
.PHONY: release
release:
	./packaging/build-release.sh $(VERSION)

## aur: regenerate the -bin PKGBUILD/.SRCINFO from dist/ (VERSION=x.y.z)
.PHONY: aur
aur:
	./packaging/aur/gen-pkgbuild.sh $(VERSION)

## clean: remove build artifacts and cache
.PHONY: clean
clean:
	rm -f $(BINARY)
	rm -rf $(BUILD_DIR)/dist $(BUILD_DIR)/data
