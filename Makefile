.PHONY: build run clean

CLI_BINARY_NAME=tds
SVR_BINARY_NAME=tdss
BUILD_DIR=./bin
CMD_DIR=./cmd/
CLI_PATH=$(CMD_DIR)/cli/
SVR_PATH=$(CMD_DIR)/cli/

build_cli:
	@echo "Building $(CLI_BINARY_NAME)..."
	@rm -f $(BUILD_DIR)/$(CLI_BINARY_NAME)
	@go build -o $(BUILD_DIR)/$(CLI_BINARY_NAME) $(CLI_PATH)
	@echo "✅ Build complete: $(CLI_BINARY_NAME)"

build_svr:
	@echo "Building $(SVR_BINARY_NAME)..."
	@rm -f $(BUILD_DIR)/$(SVR_BINARY_NAME)
	@go build -o $(BUILD_DIR)/$(SVR_BINARY_NAME) $(SVR_PATH)
	@echo "✅ Build complete: $(SVR_BINARY_NAME)"

build:
	$(MAKE) build_cli 
	$(MAKE) build_svr

run: build_svr
	@echo "Running $(SVR_BINARY_NAME)..."
	@$(BUILD_DIR)/$(SVR_BINARY_NAME)

run-cli: build_cli
	@echo "Running $(CLI_BINARY_NAME)..."
	@$(BUILD_DIR)/$(CLI_BINARY_NAME)

test:
	@echo "Running tests..."
	@go test -cover ./...
	@echo "✅ Tests complete"

clean:
	@echo "Cleaning up..."
	@rm -f $(BUILD_DIR)/$(CLI_BINARY_NAME)
	@rm -f $(BUILD_DIR)/$(SVR_BINARY_NAME)
	@rm -f code-scan-results_*.md
	@go clean -testcache
	@echo "✅ Clean complete"

clean-all:
	$(MAKE) clean
	@echo "Cleaning go cache"
	@ go clean -cache -modcache -testcache

install: build
	@echo "Installing CLI tool $(CLI_BINARY_NAME)..."
	@go install $(CLI_PATH)
	@echo "Installing server $(SVR_BINARY_NAME)..."
	@go install $(SVR_PATH)
	@echo "✅ Installed to GOPATH/bin"
