# AGENTS.md - Geo API Development Guide

## Project Overview

REST API in Go for querying countries, states, and cities (~150k cities). Go 1.22 with standard `net/http` (no external frameworks).

## Build/Lint/Test Commands

```bash
# Run locally (requires data/countries+states+cities.json)
go run ./cmd/api

# Build binary
go build -o geo-api ./cmd/api

# Download dependencies
go mod download

# Code quality (required before committing)
go fmt ./...
go vet ./...
go mod tidy
```

### Testing

```bash
go test ./...                    # Run all tests
go test -v ./...                 # Verbose output
go test -v -run TestGetCountry ./internal/handler  # Specific test
go test -cover ./...             # With coverage
```

### Docker

```bash
docker build -t geo-api .
docker run -p 8082:8082 geo-api
```

## Environment Variables

| Variable         | Default                             | Description                    |
| ---------------- | ----------------------------------- | ------------------------------ |
| `PORT`           | `8082`                              | Server port                    |
| `GEO_DATA_PATH`  | `data/countries+states+cities.json` | Path to geo dataset            |
| `METADATA_PATH`  | `data/cities_metadata.json`        | Path to city metadata          |
| `CORS_ORIGIN`    | `*`                                 | Allowed CORS origin (use specific origin for credentials) |
| `ADMIN_USER`     | -                                   | Admin API username             |
| `ADMIN_PASSWORD` | -                                   | Admin API password             |

## Project Structure

```
geo-api/
├── cmd/api/main.go           # Entry point
├── internal/
│   ├── handler/              # HTTP handlers (handler.go, auth.go)
│   ├── model/geo.go          # Data structures and DTOs
│   └── repository/geo.go     # Data access and in-memory indexes
├── data/                     # Dataset files (not in repo)
├── public/                   # Static files
└── openapi.yaml             # OpenAPI spec
```

## Code Style Guidelines

### Imports

Group stdlib first, then third-party. Example:

```go
import (
    "encoding/json"
    "net/http"
    "strconv"
    "github.com/joho/godotenv"
    "github.com/youruser/geo-api/internal/model"
)
```

### Naming Conventions

- **Packages:** lowercase (`handler`, `model`, `repository`)
- **Types:** PascalCase (`GeoRepository`, `CountrySummary`)
- **Functions/Methods:** PascalCase (`GetCountries`, `GetStats`)
- **Variables/Fields:** camelCase (`dataPath`, `countryByISO2`)
- **Acronyms:** Keep original casing (`ISO2`, not `Iso2`)

### Error Handling

- Use `fmt.Errorf("action: %w", err)` for wrapped errors
- Log errors at boundary (handlers), not deep layers:

```go
data, err := os.ReadFile(dataPath)
if err != nil {
    return nil, fmt.Errorf("reading geo data: %w", err)
}
```

### HTTP Handlers

- Use `http.HandlerFunc` with receiver methods
- Status codes: `200 OK`, `400 Bad Request`, `404 Not Found`, `500 Internal Server Error`
- Use helper functions for consistent JSON responses

### Concurrency

- Use `sync.RWMutex` for read-heavy workloads
- Lock before writes, RLock before reads, use `defer`:

```go
r.mu.RLock()
defer r.mu.RUnlock()
```

### JSON Serialization

- Use struct tags: `json:"field_name"`
- Use `omitempty` for optional fields
- Create separate DTO types for API responses
- Use `map[string]any` for flexible metadata

### Pagination

- Support `?page=` and `?limit=` query params
- Default limit: 50, max: 500
- Return `PaginatedResponse` struct

## Repository Pattern

Layered architecture: `Handler` -> `Repository` -> Response

## Database

- In-memory (loaded from JSON on startup)
- Indexes built on startup for O(1) lookups by ISO2, ID

## API Conventions

```json
// Paginated response
{ "data": [...], "total": 250, "page": 1, "limit": 50 }

// Error
{ "error": "resource not found" }

// Search: case-insensitive substring matching via ?search=
```

## Adding New Endpoints

1. Add handler method to `internal/handler/handler.go`
2. Register route in `RegisterRoutes()`
3. Add repository method if needed in `internal/repository/geo.go`
4. Add response model if needed in `internal/model/geo.go`
5. Update `openapi.yaml`
6. Test with `go run ./cmd/api`

## Common Tasks

### Add city metadata

Admin endpoint exists: `PUT /admin/cities/{id}/metadata`  
Metadata persists to `data/cities_metadata.json`
