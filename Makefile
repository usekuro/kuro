APP_NAME=usekuro
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT ?= $(shell git rev-parse --short HEAD)
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS = -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)

.PHONY: help build test clean install lint fmt vet security bench coverage docker

# 📖 Show available commands
help:
	@echo "UseKuro Development Commands"
	@echo ""
	@echo "Building:"
	@echo "  build          Build the binary"
	@echo "  build-all      Build for all platforms"
	@echo "  install        Install to \$$GOPATH/bin"
	@echo "  clean          Clean build artifacts"
	@echo ""
	@echo "Development:"
	@echo "  run            Run HTTP mock example"
	@echo "  debug          Run with debug logging"
	@echo "  web            Start web interface"
	@echo "  validate       Validate example files"
	@echo ""
	@echo "Testing:"
	@echo "  test           Run all tests"
	@echo "  test-race      Run tests with race detection"
	@echo "  test-coverage  Generate coverage report"
	@echo "  bench          Run benchmarks"
	@echo "  test-integration  Run integration tests"
	@echo ""
	@echo "Quality:"
	@echo "  lint           Run linter"
	@echo "  fmt            Format code"
	@echo "  vet            Run go vet"
	@echo "  security       Run security scanner"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build   Build Docker image"
	@echo "  docker-run     Run Docker container"
	@echo ""
	@echo "Utilities:"
	@echo "  deps           Install dependencies"
	@echo "  gen-keys       Generate test SSH keys"
	@echo "  ports          Show occupied ports"

# 🛠️ Build the binary
build:
	@echo "🛠️ Building $(APP_NAME)..."
	@mkdir -p bin
	go build -ldflags="$(LDFLAGS)" -o bin/$(APP_NAME) ./cmd/$(APP_NAME)
	@echo "✅ Binary built: bin/$(APP_NAME)"

# 🌍 Build for all platforms
build-all: clean
	@echo "🌍 Building for all platforms..."
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/$(APP_NAME)-linux-amd64 ./cmd/$(APP_NAME)
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o bin/$(APP_NAME)-linux-arm64 ./cmd/$(APP_NAME)
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/$(APP_NAME)-darwin-amd64 ./cmd/$(APP_NAME)
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o bin/$(APP_NAME)-darwin-arm64 ./cmd/$(APP_NAME)
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/$(APP_NAME)-windows-amd64.exe ./cmd/$(APP_NAME)
	@echo "✅ All platforms built in bin/"

# 📦 Install to $GOPATH/bin
install:
	@echo "📦 Installing $(APP_NAME)..."
	go install -ldflags="$(LDFLAGS)" ./cmd/$(APP_NAME)
	@echo "✅ Installed $(APP_NAME)"

# 🧪 Run all tests
test:
	@echo "🧪 Running tests..."
	go test -v ./...

# 🏃 Run tests with race detection
test-race:
	@echo "🏃 Running tests with race detection..."
	go test -race -v ./...

# 📊 Generate test coverage report
test-coverage:
	@echo "📊 Generating coverage report..."
	go test -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

# 🚀 Run benchmarks
bench:
	@echo "🚀 Running benchmarks..."
	go test -bench=. -benchmem ./...

# 🔗 Run integration tests
test-integration: build
	@echo "🔗 Running integration tests..."
	go test -v -tags=integration ./tests/integration/

# 🧹 Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	rm -rf bin dist coverage.out coverage.html
	rm -f tests/testdata/id_rsa tests/testdata/id_rsa.pub
	go clean -cache -testcache

# 🎨 Format code
fmt:
	@echo "🎨 Formatting code..."
	go fmt ./...
	@echo "✅ Code formatted"

# 🔍 Run go vet
vet:
	@echo "🔍 Running go vet..."
	go vet ./...

# 🔒 Run security scanner
security:
	@echo "🔒 Running security scanner..."
	@which gosec > /dev/null || (echo "Installing gosec..." && go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest)
	gosec ./...

# 📝 Run linter
lint:
	@echo "📝 Running linter..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin)
	golangci-lint run

# 🚀 Run an HTTP mock example
run:
	@echo "🚀 Running HTTP mock example..."
	go run ./cmd/$(APP_NAME) run mocks/http_simple.kuro

# 🔁 Validate example files
validate:
	@echo "🔁 Validating example files..."
	@for file in examples/*.kuro mocks/*.kuro; do \
		echo "Validating $$file..."; \
		go run ./cmd/$(APP_NAME) validate "$$file" || exit 1; \
	done
	@echo "✅ All files validated"

# 🧵 Run all mocks in a folder (batch mode)
boot:
	go run ./cmd/$(APP_NAME) boot mocks/

# 📦 Export a kuro collection
export:
	go run ./cmd/$(APP_NAME) export --collection mocks/collection.kuroc --output dist/full_demo.zip

# 📥 Install dependencies
deps:
	@echo "📥 Installing dependencies..."
	go mod download
	go mod tidy

# 🌐 Run web server (interface)
web:
	@echo "🌐 Starting web interface at http://localhost:8798"
	go run ./cmd/$(APP_NAME) web

# 🏥 Health check of web server
health:
	@echo "🏥 Checking server health..."
	@curl -s http://localhost:8798/health | jq . || echo "❌ Server not responding or jq not installed"

# 🐛 Run mock with debug logging
debug:
	@echo "🐛 Running with debug logging..."
	LOG_LEVEL=debug go run ./cmd/$(APP_NAME) run mocks/http_simple.kuro

# 🔍 Test HTTP mock health (port 8080 by default)
test-mock-health:
	@echo "🔍 Testing mock health endpoint..."
	@curl -s http://localhost:8080/health | jq . || echo "❌ Mock not responding on port 8080"

# 📊 Show occupied ports
ports:
	@echo "📊 Checking occupied ports..."
	@lsof -i :8080 -i :8798 -i :2022 || echo "No processes found on common ports"

# 🧪 Complete mock test
test-mock:
	@echo "🧪 Running comprehensive mock test..."
	@echo "Starting mock in background..."
	@LOG_LEVEL=info go run ./cmd/$(APP_NAME) run mocks/http_simple.kuro &
	@sleep 2
	@echo "Testing health endpoint..."
	@curl -s http://localhost:8080/health || echo "Health check failed"
	@echo "Testing custom endpoints..."
	@curl -s http://localhost:8080/hello || echo "Hello endpoint failed"
	@curl -s http://localhost:8080/status || echo "Status endpoint failed"
	@echo "Stopping background process..."
	@pkill -f "usekuro run" || echo "No process to kill"

# 🔑 Generate test SSH keys for development
gen-keys:
	@echo "🔑 Generating test SSH keys..."
	@mkdir -p tests/testdata
	@rm -f tests/testdata/id_rsa tests/testdata/id_rsa.pub
	@ssh-keygen -t rsa -b 4096 -f tests/testdata/id_rsa -N "" -q
	@chmod 600 tests/testdata/id_rsa
	@echo "✅ Test keys generated in tests/testdata/"

# 🐳 Build Docker image
docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t usekuro/kuro:latest -t usekuro/kuro:$(VERSION) .
	@echo "✅ Docker image built: usekuro/kuro:$(VERSION)"

# 🐳 Run Docker container
docker-run:
	@echo "🐳 Running Docker container..."
	docker run --rm -p 8798:8798 usekuro/kuro:latest

# 📋 Development setup
dev-setup:
	@echo "📋 Setting up development environment..."
	@make deps
	@make gen-keys
	@echo "✅ Development environment ready"

# 🚢 Release preparation
pre-release: clean lint test security test-coverage validate build-all
	@echo "🚢 Pre-release checks completed successfully"

# Default target
default: help
