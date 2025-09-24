# Zylo Compiler Makefile

.PHONY: all build test clean install fmt vet doc

# Default target
all: build test

# Build the compiler
build:
	cd cmd/zylo && go build -o ../../bin/zylo

# Run all tests
test:
	go test ./...

# Run integration tests
test-integration:
	go test ./tests/...

# Run lexer tests
test-lexer:
	go test ./internal/lexer/...

# Run parser tests
test-parser:
	go test ./internal/parser/...

# Run codegen tests
test-codegen:
	go test ./internal/codegen/...

# Run semantic analysis tests
test-sema:
	go test ./internal/sema/...

# Clean build artifacts
clean:
	rm -rf bin/
	go clean ./...

# Install the compiler globally
install: build
	cp bin/zylo /usr/local/bin/zylo

# Format Go code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Generate documentation
doc:
	go doc -all ./...

# Build examples
build-examples:
	./bin/zylo build examples/hello.zylo
	./bin/zylo build examples/fib.zylo

# Run examples
run-examples:
	./bin/zylo run examples/hello.zylo
	./bin/zylo run examples/fib.zylo

# Development setup
dev-setup:
	go mod tidy
	go mod download

# CI pipeline
ci: fmt vet test build

# Create distribution
dist: clean build
	mkdir -p dist/
	cp bin/zylo dist/
	cp -r examples/ dist/
	cp -r docs/ dist/
	cp README.md dist/
	cp LICENSE dist/

# Docker build
docker-build:
	docker build -t zylo-compiler .

# Docker run
docker-run:
	docker run --rm zylo-compiler

# Help
help:
	@echo "Zylo Compiler - Available targets:"
	@echo "  all              - Build and test everything"
	@echo "  build            - Build the compiler"
	@echo "  test             - Run all tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-lexer       - Run lexer tests"
	@echo "  test-parser      - Run parser tests"
	@echo "  test-codegen     - Run codegen tests"
	@echo "  test-sema        - Run semantic analysis tests"
	@echo "  clean            - Clean build artifacts"
	@echo "  install          - Install compiler globally"
	@echo "  fmt              - Format Go code"
	@echo "  vet              - Run go vet"
	@echo "  doc              - Generate documentation"
	@echo "  build-examples   - Build example programs"
	@echo "  run-examples     - Run example programs"
	@echo "  dev-setup        - Setup development environment"
	@echo "  ci               - Run CI pipeline"
	@echo "  dist             - Create distribution package"
	@echo "  docker-build     - Build Docker image"
	@echo "  docker-run       - Run Docker container"
	@echo "  help             - Show this help message"
