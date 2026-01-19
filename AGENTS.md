# tap - Agent Operations Guide

> Keep this file under 60 lines. Operational commands only.

## Build

```bash
go build -o tap ./cmd/tap
```

## Test

```bash
# All tests
go test ./... -v

# Specific package
go test ./internal/core/... -v

# With coverage
go test ./... -coverprofile=coverage.out
```

## Run

```bash
# Development
go run ./cmd/tap

# Built binary
./tap
```

## Lint

```bash
# Install if needed: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run
```

## Dependencies

```bash
# Add dependency
go get github.com/package/name

# Tidy
go mod tidy
```

## Key Dependencies

| Package | Purpose |
|---------|---------|
| github.com/charmbracelet/bubbletea | TUI framework |
| github.com/charmbracelet/bubbles | TUI components (list, textinput) |
| github.com/charmbracelet/huh | Form library |
| github.com/charmbracelet/lipgloss | Styling |
| github.com/spf13/cobra | CLI framework |
| github.com/spf13/viper | Config management |
| github.com/adrg/xdg | XDG paths |
| gopkg.in/yaml.v3 | YAML parsing |

## Project Structure

```
cmd/tap/main.go          # Entry point
internal/core/           # Business logic (scanner, parser, executor)
internal/tui/            # Bubble Tea components
internal/config/         # Configuration management
internal/cli/            # Cobra commands
```

## Commit Convention

```
feat(core): add metadata parser
fix(tui): handle window resize
test(scanner): add directory scanning tests
chore: update dependencies
```
