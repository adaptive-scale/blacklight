# Binary name
BINARY_NAME=blacklight
ifeq ($(OS),Windows_NT)
    BINARY_EXTENSION=.exe
else
    BINARY_EXTENSION=
endif

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-s -w"

# Determine the operating system
ifeq ($(OS),Windows_NT)
    DETECTED_OS=windows
    INSTALL_DIR=$(USERPROFILE)\bin
    SEP=\\
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Darwin)
        DETECTED_OS=darwin
    else
        DETECTED_OS=linux
    endif
    INSTALL_DIR=/usr/local/bin
    SEP=/
endif

# Determine the architecture
ifeq ($(OS),Windows_NT)
    ifeq ($(PROCESSOR_ARCHITECTURE),AMD64)
        ARCH=amd64
    else ifeq ($(PROCESSOR_ARCHITECTURE),ARM64)
        ARCH=arm64
    else
        ARCH=386
    endif
else
    UNAME_M := $(shell uname -m)
    ifeq ($(UNAME_M),x86_64)
        ARCH=amd64
    else ifeq ($(UNAME_M),arm64)
        ARCH=arm64
    else
        ARCH=386
    endif
endif

# Output directory
OUTDIR=dist
BINARY_PATH=$(OUTDIR)$(SEP)$(BINARY_NAME)$(BINARY_EXTENSION)

# Color codes for help output
YELLOW=\033[1;33m
NC=\033[0m
GREEN=\033[1;32m
BLUE=\033[1;34m

.PHONY: all build clean test deps tidy install help

# Help target
help: ## Display this help message
	@printf "$(YELLOW)Blacklight Secret Scanner - Available Commands:$(NC)\n"
	@printf "\n"
	@printf "$(GREEN)Usage:$(NC)\n"
	@printf "  make $(BLUE)<target>$(NC)\n"
	@printf "\n"
	@printf "$(GREEN)Targets:$(NC)\n"
	@awk '{ \
		if ($$0 ~ /^[a-zA-Z0-9_-]+:.*?## .*$$/) { \
			printf "  $(BLUE)%-15s$(NC) %s\n", substr($$1, 1, length($$1)-1), substr($$0, index($$0,"##")+3) \
		} \
	}' $(MAKEFILE_LIST)
	@printf "\n"
	@printf "$(GREEN)Examples:$(NC)\n"
	@printf "  make build              # Build for current platform\n"
	@printf "  make build-all          # Build for all platforms\n"
	@printf "  make install            # Install to system\n"
	@printf "\n"
	@printf "$(GREEN)Current Configuration:$(NC)\n"
	@printf "  OS: $(DETECTED_OS)\n"
	@printf "  Architecture: $(ARCH)\n"
	@printf "  Output Directory: $(OUTDIR)\n"
	@printf "  Install Directory: $(INSTALL_DIR)\n"

all: clean build ## Clean and build the project

build: ## Build binary for current platform
	@printf "Building for $(DETECTED_OS)/$(ARCH)...\n"
	@mkdir -p $(OUTDIR)
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_PATH)
	@printf "Binary built at $(BINARY_PATH)\n"

build-all: clean ## Build binaries for all supported platforms
	@printf "Building for multiple platforms...\n"
	@mkdir -p $(OUTDIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(OUTDIR)/$(BINARY_NAME)-linux-amd64
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(OUTDIR)/$(BINARY_NAME)-linux-arm64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(OUTDIR)/$(BINARY_NAME)-darwin-amd64
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(OUTDIR)/$(BINARY_NAME)-darwin-arm64
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(OUTDIR)/$(BINARY_NAME)-windows-amd64.exe
	@printf "All binaries built in $(OUTDIR)/\n"

clean: ## Clean build artifacts and temporary files
	@printf "Cleaning...\n"
	@rm -rf $(OUTDIR)
	$(GOCLEAN)

test: ## Run all tests
	@printf "Running tests...\n"
	$(GOTEST) -v ./...

deps: ## Download and verify dependencies
	@printf "Downloading dependencies...\n"
	$(GOGET) ./...

tidy: ## Tidy and verify dependencies
	@printf "Tidying up modules...\n"
	$(GOMOD) tidy

install: build ## Install binary to system
	@printf "Installing binary...\n"
ifeq ($(OS),Windows_NT)
	@if not exist "$(INSTALL_DIR)" mkdir "$(INSTALL_DIR)"
	@copy /y "$(BINARY_PATH)" "$(INSTALL_DIR)"
	@printf "Installed at $(INSTALL_DIR)$(SEP)$(BINARY_NAME)$(BINARY_EXTENSION)\n"
	@printf "Please add $(INSTALL_DIR) to your PATH if not already added\n"
else
	@cp "$(BINARY_PATH)" "$(INSTALL_DIR)"
	@printf "Installed at $(INSTALL_DIR)$(SEP)$(BINARY_NAME)$(BINARY_EXTENSION)\n"
endif 