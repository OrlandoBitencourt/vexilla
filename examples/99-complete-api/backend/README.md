# Vexilla Demo - Backend API

Go + Gin + Vexilla backend demonstrating feature flags in production.

## Quick Start

```bash
# Install dependencies
go mod download

# Run server
go run main.go
```

Server runs on `http://localhost:8080`

## Features

- ✅ Deterministic rollout based on CPF hash
- ✅ Kill switch for emergency shutdown
- ✅ Rate limiting middleware (flag-controlled)
- ✅ Admin endpoints for cache management
- ✅ CORS enabled for frontend

## Environment Variables

```bash
# Optional - defaults shown
FLAGR_URL=http://localhost:18000
PORT=8080
```

## API Endpoints

See main [README.md](../README.md#api-documentation) for full API documentation.

## Architecture

### Deterministic Bucket Calculation

```go
func DeterministicBucketFromCPF(cpf string) (int, error) {
    // 1. Clean CPF (remove non-digits)
    // 2. Hash with SHA-256
    // 3. Convert to uint64
    // 4. Modulo 100 → bucket (0-99)
}
```

### Flag Evaluation

```go
ctx := vexilla.NewContext(userID).
    WithAttribute("role", role).
    WithAttribute("cpf", cpf)

enabled := client.Bool(context.Background(), "api.checkout.v2", ctx)
```

## Development

```bash
# Run with auto-reload (install air first)
air

# Run tests
go test ./...

# Build binary
go build -o vexilla-api main.go
```
