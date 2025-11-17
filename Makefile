# Makefile for gofi - Application Launcher

# Project metadata
BINARY_NAME=gofi
VERSION=0.1.0
BUILD_DIR=build
INSTALL_PREFIX=/usr/local
CONFIG_DIR=$(HOME)/.config/gofi
ASSETS_DIR=assets

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -s -w"
BUILDFLAGS=-trimpath

# Detect OS for installation paths
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
    DESKTOP_DIR=$(HOME)/.local/share/applications
    BIN_DIR=$(HOME)/.local/bin
else
    DESKTOP_DIR=$(INSTALL_PREFIX)/share/applications
    BIN_DIR=$(INSTALL_PREFIX)/bin
endif

# Default target
.DEFAULT_GOAL := help

##@ General

.PHONY: help
help: ## Display this help message
	@printf "\033[1mgofi - Application Launcher\033[0m\n"
	@printf "Version: $(VERSION)\n\n"
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[34m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[34m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: build
build: ## Build the binary
	@printf "\033[32mBuilding $(BINARY_NAME)...\033[0m\n"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILDFLAGS) $(LDFLAGS) -v -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/gofi/main.go
	@printf "\033[32m✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME)\033[0m\n"

.PHONY: build-release
build-release: clean ## Build optimized release binary
	@printf "\033[32mBuilding release version $(VERSION)...\033[0m\n"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILDFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/gofi/main.go
	@printf "\033[32m✓ Release build complete\033[0m\n"

.PHONY: run
run: build ## Build and run the application
	@printf "\033[34mRunning $(BINARY_NAME)...\033[0m\n"
	@$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: dev
dev: ## Run in development mode (with debug output)
	@printf "\033[34mRunning in development mode...\033[0m\n"
	DEBUG=1 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) . && DEBUG=1 $(BUILD_DIR)/$(BINARY_NAME)

##@ Code Quality

.PHONY: fmt
fmt: ## Format code
	@printf "\033[34mFormatting code...\033[0m\n"
	$(GOFMT) ./...
	@printf "\033[32m✓ Code formatted\033[0m\n"

.PHONY: vet
vet: ## Run go vet
	@printf "\033[34mRunning go vet...\033[0m\n"
	$(GOCMD) vet ./...
	@printf "\033[32m✓ Vet complete\033[0m\n"

.PHONY: lint
lint: ## Run golangci-lint (requires golangci-lint installed)
	@printf "\033[34mRunning linter...\033[0m\n"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
		printf "\033[32m✓ Lint complete\033[0m\n"; \
	else \
		printf "\033[33m⚠ golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest\033[0m\n"; \
	fi

.PHONY: check
check: fmt vet test ## Run all checks (format, vet, test)
	@printf "\033[32m✓ All checks passed\033[0m\n"

##@ Dependencies

.PHONY: deps
deps: ## Download dependencies
	@printf "\033[34mDownloading dependencies...\033[0m\n"
	$(GOMOD) download
	@printf "\033[32m✓ Dependencies downloaded\033[0m\n"

.PHONY: deps-update
deps-update: ## Update dependencies
	@printf "\033[34mUpdating dependencies...\033[0m\n"
	$(GOGET) -u ./...
	$(GOMOD) tidy
	@printf "\033[32m✓ Dependencies updated\033[0m\n"

.PHONY: deps-tidy
deps-tidy: ## Tidy go.mod and go.sum
	@printf "\033[34mTidying dependencies...\033[0m\n"
	$(GOMOD) tidy
	@printf "\033[32m✓ Dependencies tidied\033[0m\n"

.PHONY: deps-verify
deps-verify: ## Verify dependencies
	@printf "\033[34mVerifying dependencies...\033[0m\n"
	$(GOMOD) verify
	@printf "\033[32m✓ Dependencies verified\033[0m\n"

##@ Cleanup

.PHONY: clean
clean: ## Remove build artifacts
	@printf "\033[33mCleaning build artifacts...\033[0m\n"
	@rm -rf $(BUILD_DIR)
	$(GOCLEAN)
	@printf "\033[32m✓ Clean complete\033[0m\n"

.PHONY: clean-cache
clean-cache: ## Clean Go build cache
	@printf "\033[33mCleaning Go cache...\033[0m\n"
	$(GOCLEAN) -cache
	@printf "\033[32m✓ Cache cleaned\033[0m\n"

.PHONY: clean-all
clean-all: clean clean-cache ## Clean everything (build + cache)
	@printf "\033[32m✓ Full clean complete\033[0m\n"

##@ Distribution

.PHONY: package
package: clean build-release ## Create distribution package
	@printf "\033[34mCreating package...\033[0m\n"
	@mkdir -p $(BUILD_DIR)/package/$(BINARY_NAME)-$(VERSION)
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(BUILD_DIR)/package/$(BINARY_NAME)-$(VERSION)/
	@cp README.md $(BUILD_DIR)/package/$(BINARY_NAME)-$(VERSION)/
	@cp README_CONFIG.md $(BUILD_DIR)/package/$(BINARY_NAME)-$(VERSION)/
	@cp -r $(ASSETS_DIR) $(BUILD_DIR)/package/$(BINARY_NAME)-$(VERSION)/
	@cd $(BUILD_DIR)/package && tar -czf $(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-$(VERSION)
	@printf "\033[32m✓ Package created: $(BUILD_DIR)/package/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz\033[0m\n"

.PHONY: dist
dist: package ## Alias for package

##@ Nix (if using Nix)

.PHONY: nix-build
nix-build: ## Build using Nix
	@printf "\033[34mBuilding with Nix...\033[0m\n"
	nix build
	@printf "\033[32m✓ Nix build complete\033[0m\n"

.PHONY: nix-run
nix-run: ## Run using Nix
	@printf "\033[34mRunning with Nix...\033[0m\n"
	nix run

.PHONY: nix-shell
nix-shell: ## Enter Nix development shell
	@printf "\033[34mEntering Nix shell...\033[0m\n"
	nix develop

##@ Documentation

.PHONY: docs
docs: ## Generate documentation
	@printf "\033[34mGenerating documentation...\033[0m\n"
	@mkdir -p $(BUILD_DIR)/docs
	@$(GOCMD) doc -all > $(BUILD_DIR)/docs/godoc.txt
	@printf "\033[32m✓ Documentation generated: $(BUILD_DIR)/docs/\033[0m\n"

##@ Information

.PHONY: version
version: ## Show version
	@printf "$(BINARY_NAME) version $(VERSION)\n"

.PHONY: info
info: ## Show project information
	@printf "\033[1mProject Information\033[0m\n"
	@printf "  Name:           $(BINARY_NAME)\n"
	@printf "  Version:        $(VERSION)\n"
	@printf "  Build Dir:      $(BUILD_DIR)\n"
	@printf "  Install Prefix: $(INSTALL_PREFIX)\n"
	@printf "  Config Dir:     $(CONFIG_DIR)\n"
	@printf "  Binary Dir:     $(BIN_DIR)\n"
	@printf "  Desktop Dir:    $(DESKTOP_DIR)\n"
	@printf "\n"
	@printf "\033[1mGo Environment\033[0m\n"
	@printf "  Go Version:     $$(go version)\n"
	@printf "  GOPATH:         $$(go env GOPATH)\n"
	@printf "  GOOS:           $$(go env GOOS)\n"
	@printf "  GOARCH:         $$(go env GOARCH)\n"

##@ Quick Commands

.PHONY: all
all: clean build test ## Build and test everything

.PHONY: rebuild
rebuild: clean build ## Clean and rebuild

.PHONY: quick
quick: ## Quick build (no clean)
	@$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@printf "\033[32m✓ Quick build complete\033[0m\n"