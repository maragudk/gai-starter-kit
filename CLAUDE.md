# CLAUDE.md - Guide for AI Tools

## Build/Test/Lint Commands
- Build: `make start` (builds CSS and runs app)
- Test all: `make test` (runs tests with coverage)
- Test single: `go test -v -tags sqlite_fts5 ./path/to/package -run TestName`
- Coverage: `make cover` (opens HTML coverage report)
- Lint: `make lint` (runs golangci-lint)
- Build Docker: `make build-docker` (builds CSS and Docker images)

## Code Style Guidelines
- **Imports**: Group by standard lib, third-party, internal; alphabetical within groups
- **Formatting**: Follow Go standard formatting (`go fmt`)
- **Types**: Use clear, documented custom types; proper struct field alignment
- **Naming**: PascalCase for exported types/functions, camelCase for variables
- **Error Handling**: Check and propagate errors with context (via maragu.dev/errors)
- **Testing**: Table-driven tests, helper functions with t.Helper(), use maragu.dev/is for assertions, prefer integration tests over mocks
- **Project Structure**: Clean separation between packages (cmd/app, http, sql, model, ai)

## Testing

- **Prefer integration tests over mocks**: External dependencies can generally be used in tests directly, so there
  is little need for mocking. See the `sqltest` package for getting a database connection for testing,
  and the `aitest` package for getting a client for both chat completion and embedding generation.
