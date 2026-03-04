.DEFAULT_GOAL := help

override CACHE_DIR := $(CURDIR)/.cache
GO_BUILD_CACHE := $(CACHE_DIR)/go-build
GO_MOD_CACHE := $(CACHE_DIR)/go/pkg/mod
GO_TMP_DIR := $(CACHE_DIR)/go/tmp

GOENV := env GOCACHE=$(GO_BUILD_CACHE) GOMODCACHE=$(GO_MOD_CACHE) GOTMPDIR=$(GO_TMP_DIR)

VERSION ?= $(strip $(shell cat VERSION 2>/dev/null))
LDFLAGS := -ldflags "-X wt/internal/buildinfo.Version=$(VERSION)"

.PHONY: help
help: ## Show help
	@awk 'BEGIN {FS = ":.*##"; print "Targets:"} /^[a-zA-Z0-9_.-]+:.*##/ {printf "  %-14s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: init
init: ## Create local cache dirs
	@mkdir -p $(GO_BUILD_CACHE) $(GO_MOD_CACHE) $(GO_TMP_DIR)

.PHONY: fmt
fmt: ## Run gofmt (writes files)
	@files="$$(find . -name '*.go' -not -path './.cache/*')"; \
	if [ -n "$$files" ]; then $(GOENV) gofmt -w $$files; fi

.PHONY: fmt-check
fmt-check: ## Verify gofmt is clean
	@files="$$(find . -name '*.go' -not -path './.cache/*')"; \
	if [ -z "$$files" ]; then exit 0; fi; \
	out="$$(gofmt -l $$files)"; \
	if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi

.PHONY: fix
fix: init ## Run go fix (writes files)
	@$(GOENV) go fix ./...

.PHONY: fix-diff
fix-diff: init ## Show go fix patch (no writes)
	@$(GOENV) go fix -diff ./...

.PHONY: fix-check
fix-check: init ## Verify go fix is clean
	@out="$$( $(GOENV) go fix -diff ./... 2>&1 )"; if [ -n "$$out" ]; then echo "$$out"; echo; echo "go fix needed: run 'make fix'"; exit 1; fi

.PHONY: check
check: init fmt-check fix-check ## Run required checks

.PHONY: test
test: check ## Run tests (requires fmt/fix clean)
	@$(GOENV) go test ./...

.PHONY: build
build: check ## Build (requires fmt/fix clean)
	@$(GOENV) go build $(LDFLAGS) ./cmd/wt

.PHONY: run
run: init ## Run (ARGS="--help")
	@$(GOENV) go run $(LDFLAGS) ./cmd/wt $(ARGS)

.PHONY: clean
clean: ## Remove local caches
	@case "$(CACHE_DIR)" in \
		"$(CURDIR)/.cache") ;; \
		*) echo "refusing to clean unexpected CACHE_DIR=$(CACHE_DIR)"; exit 1 ;; \
	esac
	@rm -rf -- "$(CACHE_DIR)"

.PHONY: premerge
premerge: check test ## Pre-merge gate (version/release notes + tests)
	@./scripts/premerge_verify.sh
