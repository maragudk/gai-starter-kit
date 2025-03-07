# CLAUDE.md - Guide for AI Tools

## Build/Test/Lint Commands
- Build: `make start` (builds CSS and runs app)
- Test all: `make test` (runs tests with coverage)
- Test single: `go test -v ./path/to/package -run TestName`
- Coverage: `make cover` (opens HTML coverage report)
- Lint: `make lint` (runs golangci-lint)
- Build Docker: `make build-docker` (builds CSS and Docker images)

## Code Style Guidelines
- **Imports**: Group by standard lib, third-party, internal; alphabetical within groups
- **Formatting**: Follow Go standard formatting (`go fmt`)
- **Types**: Use clear, documented custom types; proper struct field alignment
- **Naming**: PascalCase for exported types/functions, camelCase for variables
- **Error Handling**: Check and propagate errors with context (via maragu.dev/errors)
- **Testing**: Table-driven tests, helper functions with t.Helper(), use maragu.dev/is for assertions
- **Project Structure**: Clean separation between packages (cmd/app, http, sql, model)