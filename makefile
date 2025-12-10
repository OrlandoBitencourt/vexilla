.PHONY: test test-v test-cover test-race test-bench clean help

# Run all tests
test:
	@echo "🧪 Running tests..."
	go test -v

# Run tests with verbose output
test-v:
	@echo "🧪 Running tests (verbose)..."
	go test -v -cover

# Run tests with coverage
test-cover:
	@echo "📊 Running tests with coverage..."
	go test -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -html=coverage.out -o coverage.txt
	@echo "✅ Coverage report: coverage.html"

# Run tests with race detector
test-race:
	@echo "🏁 Running tests with race detector..."
	go test -race

# Run benchmarks
test-bench:
	@echo "⚡ Running benchmarks..."
	go test -bench=. -benchmem

# Clean test artifacts
clean:
	@echo "🧹 Cleaning..."
	rm -f coverage.out coverage.html
	rm -rf /tmp/vexilla-cache*
	go clean -testcache

# Show help
help:
	@echo "Vexilla Test Commands"
	@echo "====================="
	@echo ""
	@echo "make test        - Run all tests"
	@echo "make test-v      - Run tests with verbose output"
	@echo "make test-cover  - Generate coverage report"
	@echo "make test-race   - Run with race detector"
	@echo "make test-bench  - Run benchmarks"
	@echo "make clean       - Clean test artifacts"