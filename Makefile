.PHONY: all build clean test fmt vet lint install uninstall run help

# Binary name
BINARY_NAME=aws-mfa-sso
OUTPUT_NAME=aws-mfa-sso

# Main package path
MAIN_PATH=./cmd/aws-mfa-sso

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Build parameters
LDFLAGS=-ldflags "-s -w"

all: test build ## Run tests and build the binary

build: ## Build the binary
	$(GOBUILD) $(LDFLAGS) -o $(OUTPUT_NAME) $(MAIN_PATH)
	@echo "Build complete: $(OUTPUT_NAME)"

clean: ## Remove build artifacts
	$(GOCLEAN)
	rm -f $(OUTPUT_NAME)
	rm -f -ws-mfa-sso
	@echo "Clean complete"

test: ## Run tests
	$(GOTEST) -v ./...

fmt: ## Format Go code
	$(GOFMT) ./...

vet: ## Run go vet
	$(GOVET) ./...

lint: fmt vet ## Run formatters and linters

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

install: build ## Install the binary to GOPATH/bin
	$(GOCMD) install $(MAIN_PATH)

uninstall: ## Remove the binary from GOPATH/bin
	rm -f $(GOPATH)/bin/$(BINARY_NAME)

run: build ## Build and run the binary
	./$(OUTPUT_NAME)

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
