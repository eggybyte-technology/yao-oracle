.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
DOCKER_REGISTRY ?= docker.io/eggybyte
GO_FILES := $(shell find . -name '*.go' -not -path "./pb/*")

# Development
.PHONY: dev-setup
dev-setup: ## Install development dependencies (Go + Flutter)
	@echo "Installing Go development dependencies..."
	@go install github.com/bufbuild/buf/cmd/buf@latest
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Installing Flutter/Dart dependencies..."
	@flutter pub global activate protoc_plugin 21.1.2
	@cd frontend/dashboard && flutter pub get
	@echo "✅ Development setup complete"
	@echo ""
	@echo "Note: Make sure the following are in your PATH:"
	@echo "  - $$HOME/.pub-cache/bin (for protoc-gen-dart)"
	@echo "  - $$GOPATH/bin (for Go tools)"

.PHONY: tidy
tidy: ## Tidy Go modules
	@echo "Tidying Go modules..."
	@go mod tidy
	@echo "✅ Go modules tidied"

# Proto generation
.PHONY: proto-generate
proto-generate: ## Generate Go code from proto files
	@bash scripts/proto-generate.sh

.PHONY: proto-lint
proto-lint: ## Lint proto files
	@buf lint api

.PHONY: proto-format
proto-format: ## Format proto files
	@buf format -w api

.PHONY: dart-proto-generate
dart-proto-generate: ## Generate Dart gRPC code for Flutter dashboard (protobuf 4.x compatible)
	@bash scripts/generate_dart_grpc.sh

.PHONY: proto-dart
proto-dart: dart-proto-generate ## Alias for dart-proto-generate

# Build
.PHONY: build
build: ## Build all services for multiple architectures (linux/amd64,arm64)
	@bash scripts/build.sh

.PHONY: build-local
build-local: proto-generate ## Build all services for local architecture only (quick dev build)
	@echo "Building all services for local development..."
	@mkdir -p bin
	@go build -o bin/proxy ./cmd/proxy
	@go build -o bin/node ./cmd/node
	@go build -o bin/dashboard ./cmd/dashboard
	@go build -o bin/mock-admin ./cmd/mock-admin
	@echo "✅ All services built"

.PHONY: build-proxy
build-proxy: ## Build proxy service for multiple architectures
	@bash scripts/build.sh --service proxy

.PHONY: build-node
build-node: ## Build node service for multiple architectures
	@bash scripts/build.sh --service node

.PHONY: build-dashboard
build-dashboard: ## Build dashboard service for multiple architectures
	@bash scripts/build.sh --service dashboard

.PHONY: build-mock-admin
build-mock-admin: proto-generate ## Build mock-admin service for local testing
	@echo "Building mock-admin for local development..."
	@mkdir -p bin
	@go build -o bin/mock-admin ./cmd/mock-admin
	@echo "✅ mock-admin built"

.PHONY: build-amd64
build-amd64: ## Build all services for linux/amd64 only
	@bash scripts/build.sh --arch amd64

.PHONY: build-arm64
build-arm64: ## Build all services for linux/arm64 only
	@bash scripts/build.sh --arch arm64

# Run
.PHONY: run-node
run-node: build-node ## Run cache node locally
	@./bin/node --port 8081

.PHONY: run-proxy
run-proxy: build-proxy ## Run proxy locally
	@./bin/proxy --port 8080 --config config.json

.PHONY: run-dashboard
run-dashboard: build-dashboard ## Run dashboard in test mode with mock data
	@TEST_MODE=true DASHBOARD_PASSWORD=admin123 HTTP_PORT=8081 ./bin/dashboard

.PHONY: run-mock-admin
run-mock-admin: build-mock-admin ## Run mock-admin service for testing (default port 9090)
	@./bin/mock-admin --grpc-port=9090 --password=admin123 --refresh-interval=5

.PHONY: run-dashboard-dev
run-dashboard-dev: ## Start mock-admin and Flutter dashboard for development
	@bash scripts/run-dashboard-dev.sh

# Docker
.PHONY: docker-build
docker-build: ## Build Docker images for all services (builds binaries + images)
	@bash scripts/docker-build.sh

.PHONY: docker-build-push
docker-build-push: ## Build and push Docker images to registry
	@bash scripts/docker-build.sh --push --version $(VERSION)

.PHONY: docker-build-proxy
docker-build-proxy: ## Build only proxy Docker image
	@bash scripts/docker-build.sh --service proxy

.PHONY: docker-build-node
docker-build-node: ## Build only node Docker image
	@bash scripts/docker-build.sh --service node

.PHONY: docker-build-dashboard
docker-build-dashboard: ## Build only dashboard Docker image
	@bash scripts/docker-build.sh --service dashboard

.PHONY: docker-build-skip
docker-build-skip: ## Build Docker images without rebuilding binaries
	@bash scripts/docker-build.sh --skip-build

# Test
.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -cover ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

.PHONY: test-dashboard
test-dashboard: ## Start mock admin and dashboard dev server for testing
	@bash scripts/test-dashboard.sh

# Lint
.PHONY: lint
lint: ## Run Go linters
	@echo "Running linters..."
	@go fmt ./...
	@go vet ./...
	@echo "✅ Linting complete"

# Clean
.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf pb/
	@rm -f coverage.out coverage.html
	@rm -rf frontend/dashboard/build/
	@echo "✅ Clean complete"

.PHONY: clean-dart
clean-dart: ## Clean Dart/Flutter generated code and build cache
	@echo "Cleaning Dart generated code..."
	@rm -rf frontend/dashboard/lib/generated/yao
	@cd frontend/dashboard && flutter clean
	@echo "✅ Dart clean complete"

.PHONY: clean-all
clean-all: clean clean-dart ## Clean all build artifacts (Go + Dart)

# Helm
.PHONY: helm-lint
helm-lint: ## Lint Helm chart
	@helm lint ./helm/yao-oracle

.PHONY: helm-template
helm-template: ## Render Helm templates
	@helm template yao-oracle ./helm/yao-oracle

.PHONY: helm-install
helm-install: ## Install yao-oracle Helm chart
	@bash scripts/helm-install.sh

.PHONY: helm-upgrade
helm-upgrade: ## Upgrade yao-oracle Helm chart
	@bash scripts/helm-upgrade.sh

.PHONY: helm-uninstall
helm-uninstall: ## Uninstall yao-oracle Helm chart
	@bash scripts/helm-uninstall.sh

.PHONY: helm-install-dev
helm-install-dev: ## Install Helm chart with dev values
	@bash scripts/helm-install.sh --values ./helm/yao-oracle/values-dev.yaml

.PHONY: helm-install-prod
helm-install-prod: ## Install Helm chart with prod values
	@bash scripts/helm-install.sh --values ./helm/yao-oracle/values-prod.yaml

.PHONY: helm-dry-run
helm-dry-run: ## Dry-run Helm installation
	@bash scripts/helm-install.sh --dry-run

# All-in-one targets
.PHONY: all
all: clean proto-generate dart-proto-generate build test ## Clean, generate (Go + Dart), build and test

.PHONY: ci
ci: proto-generate lint test ## Run CI checks (generate, lint, test)

.PHONY: proto-all
proto-all: proto-generate dart-proto-generate ## Generate all proto code (Go + Dart)

