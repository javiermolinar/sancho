# AGENTS.md - AI Coding Agent Guidelines

This document provides instructions for AI coding agents working in the sancho codebase.

## Project Overview

**sancho** is a TUI application implementing deep work methodology for scheduling and productivity tracking. It uses Cobra for CLI, SQLite for storage (pure Go driver, no CGO), and integrates with GitHub Copilot for LLM-based planning.


## Working worlflow

- ALWAYS create tests and run maka build, make test and linting. 
- Don't do anything else besides the created plan, don't go rogue
- Review always the changes, propose refactors or improvements
- Commit 

## Build, Test, and Lint Commands

### Quick Reference

| Command | Description |
|---------|-------------|
| `make build` | Build optimized binary for darwin/arm64 |
| `make build-dev` | Build without optimizations (faster iteration) |
| `make test` | Run all tests with race detector and coverage |
| `make lint` | Run golangci-lint |
| `make check` | Run fmt, vet, lint, and test |

### Running Tests

```bash
# Run all tests with race detector
make test

# Run tests without race detector (faster)
make test-short

# Run a single test file
go test -v ./internal/task/...

# Run a specific test function
go test -v ./internal/task/... -run TestNew

# Run a specific subtest
go test -v ./internal/task/... -run TestNew_Errors/empty_description

# Run tests with coverage report
make test-coverage

# Run integration tests
make test-integration
```

### Linting and Formatting

```bash
# Run linter
make lint

# Run linter with auto-fix
make lint-fix

# Format code
make fmt

# Run go vet
make vet

# Tidy modules
make mod
```


## Code Style Guidelines

### Import Organization

Organize imports in three groups separated by blank lines:
1. Standard library
2. External third-party packages
3. Internal project packages

```go
import (
    "context"
    "errors"
    "fmt"

    "github.com/spf13/cobra"
)
```

### Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| Files | lowercase, underscores | `sqlite_test.go`, `dateutil.go` |
| Exported functions | PascalCase | `NewApp`, `CreateTask` |
| Unexported functions | camelCase | `parseCategory`, `validateTime` |
| Interfaces | PascalCase, descriptive | `Repository`, `Client` |
| Struct types | PascalCase | `Task`, `Config`, `SQLite` |
| Constants (exported) | PascalCase with prefix | `StatusScheduled`, `CategoryDeep` |
| Constants (unexported) | camelCase | `copilotBaseURL` |
| Variables | camelCase | `scheduledDate`, `httpClient` |



### Architecture

- Favor simplicity
- Composition over complex abstractions
- Unit tests and integration tests when possible. No mocking
- Small iterations step by step, no big changes
- Use table driven tests as much as possible

### Error Handling

Define sentinel errors at package level:

```go
var (
    ErrEmptyDescription = errors.New("description cannot be empty")
    ErrInvalidCategory  = errors.New("category must be 'deep' or 'shallow'")
)
```

Wrap errors with context using `%w`:

```go
return fmt.Errorf("loading config: %w", err)
return fmt.Errorf("start time: %w", err)
```

Use early returns for guard clauses:

```go
if err := validate(input); err != nil {
    return err
}
```

Return `nil, nil` for "not found" scenarios (not an error):

```go
if err == sql.ErrNoRows {
    return nil, nil
}
```

### Function Patterns

Use `context.Context` as the first parameter for I/O operations:

```go
func (s *SQLite) CreateTask(ctx context.Context, t *task.Task) error
```

Name constructors `New` or `NewXxx`:

```go
func New(path string) (*SQLite, error)
func NewApp(repo task.Repository, cfg *config.Config) *App
```

Use deferred cleanup with explicit error ignoring:

```go
defer func() { _ = rows.Close() }()
defer func() { _ = tx.Rollback() }()
```

### Type Patterns

Use interfaces for dependency injection:

```go
type Repository interface {
    CreateTask(ctx context.Context, task *Task) error
    GetTask(ctx context.Context, id int64) (*Task, error)
}
```

Use pointers for optional/nullable fields:

```go
Outcome       *Outcome  // nil means not set
PostponedFrom *int64    // FK, nil if not postponed
```

### Documentation

Every package needs a doc comment:

```go
// Package task defines the core domain types for sancho.
package task
```

Document exported functions starting with the function name:

```go
// New creates a new Task with validation.
// date can be empty (defaults to today) or in YYYY-MM-DD format.
func New(description, category, date, start, end string) (*Task, error)
```

After any change update the ./agents-notes/PLAN.md

### Testing Patterns

Use table-driven tests with subtests:

```go
tests := []struct {
    name    string
    input   string
    wantErr error
}{
    {name: "valid input", input: "test", wantErr: nil},
    {name: "empty input", input: "", wantErr: ErrEmpty},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        _, err := Process(tt.input)
        if !errors.Is(err, tt.wantErr) {
            t.Errorf("got %v, want %v", err, tt.wantErr)
        }
    })
}
```

Use `t.Fatalf` for setup failures, `t.Errorf` for assertion failures.

### SQL Queries

Use multi-line backtick strings for queries:

```go
query := `
    SELECT id, description, category
    FROM tasks
    WHERE scheduled_date = ?
    ORDER BY scheduled_start
`
```

### File Permissions

Use octal notation with `0o` prefix:

```go
os.MkdirAll(dir, 0o755)
os.WriteFile(path, data, 0o644)
```

## Linter Configuration

The project uses golangci-lint v2 with these enabled linters:
- bodyclose, errcheck, govet, ineffassign, staticcheck, unused
- gocritic (diagnostic, style, performance tags)
- misspell (US locale), unconvert, unparam

Formatters: gofmt, goimports

Run `make lint` before committing. Use `make lint-fix` for auto-fixes.

## Key Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `modernc.org/sqlite` | Pure Go SQLite (no CGO) |
| `github.com/openai/openai-go` | OpenAI SDK for Copilot |
| `github.com/pelletier/go-toml/v2` | TOML config parsing |
| `github.com/fatih/color` | Terminal colors |

## Common Patterns

### Repository Pattern

Domain types in `internal/task/`, interface in `internal/task/repository.go`, implementation in `internal/db/sqlite.go`.

### Configuration

Layered: defaults -> file (`~/.config/sancho/config.toml`) -> environment variables.
