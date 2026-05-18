.PHONY: build test test-race vet run-local run-local-wakatime

GH_BIN ?= gh
WAKATIME_KEYCHAIN_SERVICE ?= wakatime-api-key

build:
	go build -o github-stats ./cmd

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

run-local:
	@command -v "$(GH_BIN)" >/dev/null || { echo "GitHub CLI not found. Install gh and run: gh auth login"; exit 1; }
	@cd cmd && GITHUB_TOKEN="$$($(GH_BIN) auth token)" go run .

run-local-wakatime:
	@command -v "$(GH_BIN)" >/dev/null || { echo "GitHub CLI not found. Install gh and run: gh auth login"; exit 1; }
	@cd cmd && GITHUB_TOKEN="$$($(GH_BIN) auth token)" WAKATIME_API_KEY="$$(security find-generic-password -s "$(WAKATIME_KEYCHAIN_SERVICE)" -w)" go run .
