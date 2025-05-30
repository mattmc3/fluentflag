.PHONY: build test clean all fmt lint vet help

GO=go
GOFLAGS=-trimpath
PUBLISH_DIR=publish
PACKAGE=github.com/mattmc3/fluentgo
MAIN_PACKAGE=$(PACKAGE)
BIN_NAME=foogo

help: ## Show this help
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

all: clean build test ## Clean, build, and test

build: ## Build the project
	@echo "Building $(BIN_NAME)..."
	@mkdir -p $(PUBLISH_DIR)
	$(GO) build $(GOFLAGS) -o $(PUBLISH_DIR)/$(BIN_NAME) $(MAIN_PACKAGE)

test: ## Run tests
	@echo "Running tests..."
	$(GO) test -v ./... | ./bin/colorize

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(PUBLISH_DIR)
	find . -type f -name "*.test" -delete
	find . -type f -name "*.out" -delete

# Code quality
fmt: ## Format the code
	$(GO) fmt ./...

vet: ## Run go vet
	$(GO) vet ./...

lint: ## Run golint (requires golint to be installed)
	golint -set_exit_status ./...

# Run the application
run: build ## Run the application
	./$(PUBLISH_DIR)/$(BIN_NAME)
