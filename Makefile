# IPTV TUI, BusinessRepo Makefile. CI/CD and humans both call these targets.
# Targets are slash-namespaced: tui/build, test/race, aur/release, and so on.
BINARY   := iptv
PKG      := github.com/kinncj/iptv
GO       ?= go
GOFLAGS  ?=
LDFLAGS  ?= -s -w
VERSION  ?= 0.0.0

.DEFAULT_GOAL := help

## help: list available targets
.PHONY: help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | awk -F: '{printf "  \033[36m%-18s\033[0m%s\n", $$1, $$2}'

# ── build / run ──────────────────────────────────────────────────────────────

## deps: download and tidy modules
.PHONY: deps
deps:
	$(GO) mod tidy

## tui/build: build the TUI binary into ./iptv
.PHONY: tui/build
tui/build:
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS) -X main.version=$(VERSION)" -o $(BINARY) ./app

## tui/run: build and run the TUI
.PHONY: tui/run
tui/run: tui/build
	./$(BINARY)

## tui/install: install the binary into GOBIN/PATH
.PHONY: tui/install
tui/install:
	$(GO) install $(GOFLAGS) -ldflags "$(LDFLAGS) -X main.version=$(VERSION)" ./app

# ── playlists ────────────────────────────────────────────────────────────────

## playlists/rebuild: regenerate ./playlists (all.m3u + countries/) from upstream
.PHONY: playlists/rebuild
playlists/rebuild: tui/build
	./$(BINARY) -export playlists

# ── quality ──────────────────────────────────────────────────────────────────

## test: run unit tests
.PHONY: test
test:
	$(GO) test ./... -count=1

## test/race: run tests with the race detector
.PHONY: test/race
test/race:
	$(GO) test ./... -race -count=1

## test/acceptance: run acceptance/feature tests
.PHONY: test/acceptance
test/acceptance:
	$(GO) test ./tests/... -count=1

## lint: vet and gofmt-check the codebase
.PHONY: lint
lint:
	$(GO) vet ./...
	@test -z "$$(gofmt -l app common tests)" || (echo "gofmt needed:"; gofmt -l app common tests; exit 1)

## fmt: format the codebase
.PHONY: fmt
fmt:
	gofmt -w app common tests

## ci: the gate CI runs, lint plus race tests plus build
.PHONY: ci
ci: lint test/race tui/build

# ── release / distribution ───────────────────────────────────────────────────

## release/build: cross-compile binaries + man + LICENSE into dist/ (VERSION=x.y.z)
.PHONY: release/build
release/build: semver-guard
	./packaging/build-release.sh $(VERSION)

## gh/release: create the GitHub release and upload dist/ assets (needs gh + a pushed tag)
.PHONY: gh/release
gh/release: release/build
	gh release create v$(VERSION) \
		dist/iptv-tui_linux_amd64 dist/iptv-tui_linux_arm64 \
		dist/iptv-tui_darwin_amd64 dist/iptv-tui_darwin_arm64 \
		dist/iptv-tui.1 dist/LICENSE dist/SHA256SUMS \
		--title "IPTV TUI v$(VERSION)" --generate-notes

## aur/pkgbuild: regenerate the -bin PKGBUILD/.SRCINFO from dist/ (VERSION=x.y.z)
.PHONY: aur/pkgbuild
aur/pkgbuild: release/build
	./packaging/aur/gen-pkgbuild.sh $(VERSION)

## aur/release: build, regenerate PKGBUILD, and push iptv-tui-bin to the AUR (VERSION=x.y.z)
.PHONY: aur/release
aur/release: aur/pkgbuild
	./packaging/aur/publish.sh $(VERSION)

# ── housekeeping ─────────────────────────────────────────────────────────────

## clean: remove build artifacts and cache
.PHONY: clean
clean:
	rm -f $(BINARY)
	rm -rf dist data

# semver-guard fails unless VERSION is a bare semantic version like 1.3.0.
.PHONY: semver-guard
semver-guard:
	@echo "$(VERSION)" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+$$' \
		|| { echo "VERSION must be semver (e.g. VERSION=1.3.0), got '$(VERSION)'"; exit 1; }
