# Contributing to Vexilla

Thank you for your interest in contributing to Vexilla! 🎉

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Docker (for running Flagr locally)
- Make (optional, for convenience commands)

### Local Development

1. **Clone the repository**
```bash
git clone https://github.com/OrlandoBitencourt/vexilla.git
cd vexilla
```

2. **Install dependencies**
```bash
go mod download
```

3. **Run tests**
```bash
go test ./...
```

4. **Run tests with coverage**
```bash
go test -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

5. **Run Flagr locally**
```bash
docker run -it -p 18000:18000 ghcr.io/openflagr/flagr
```

## Project Structure

```
vexilla/
├── pkg/
│   ├── vexilla/      # Core types and config
│   ├── evaluator/    # Evaluation logic
│   ├── storage/      # Storage implementations
│   ├── client/       # Flagr HTTP client
│   ├── server/       # Webhook & Admin API
│   ├── telemetry/    # OpenTelemetry
│   └── circuit/      # Circuit breaker
├── examples/         # Usage examples
└── scripts/          # Utility scripts
```

## Testing Guidelines

### Unit Tests

- All packages must have corresponding `_test.go` files
- Aim for >80% code coverage
- Use table-driven tests where appropriate
- Mock external dependencies

Example:
```go
func TestEvaluator_Evaluate(t *testing.T) {
    tests := []struct {
        name     string
        flag     vexilla.Flag
        expected string
    }{
        // test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

### Integration Tests

- Test against real Flagr instance
- Use Docker for isolated testing
- Clean up resources in defer statements

## Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Run `go vet` before committing
- Add godoc comments for exported functions

## Pull Request Process

1. **Create a feature branch**
```bash
git checkout -b feature/your-feature-name
```

2. **Make your changes**
- Write tests first (TDD)
- Implement feature
- Ensure all tests pass

3. **Commit with clear messages**
```bash
git commit -m "feat: add support for X"
```

Follow conventional commits:
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation
- `test:` - Test changes
- `refactor:` - Code refactoring

4. **Push and create PR**
```bash
git push origin feature/your-feature-name
```

5. **PR Requirements**
- All tests must pass
- Code coverage should not decrease
- Update documentation if needed
- Add entry to CHANGELOG.md

## Performance Benchmarks

Run benchmarks with:
```bash
go test -bench=. -benchmem ./pkg/evaluator/
```

## Questions?

- Open an issue for bugs
- Start a discussion for feature requests
- Join our community (link TBD)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
