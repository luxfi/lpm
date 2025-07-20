# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Lux Plugin Manager (LPM) is a command-line tool for managing virtual machine binaries for the Lux Network. It allows users to build custom repositories to provide VM and subnet definitions outside of the core plugins repository.

## Architecture

### Core Components
- **LPM Core** (`lpm/lpm.go`): Main application logic, handles commands and orchestrates workflows
- **Commands** (`cmd/`): Cobra-based CLI commands (install, uninstall, update, upgrade, add/remove repository, join subnet)
- **Workflows** (`workflow/`): Business logic for each command operation
- **State Management** (`state/`): Manages plugin installation registry and repository metadata
- **Admin Client** (`admin/`): Communicates with Lux node admin API
- **Git Integration** (`git/`): Handles repository cloning and updates

### Key Interfaces
- `workflow.Executor`: Executes workflow operations
- `workflow.Installer`: Handles VM binary installation
- `state.Repository`: Manages plugin repository interactions
- `admin.Client`: Lux node admin API client

### Data Flow
1. Commands parse user input and initialize LPM instance
2. LPM delegates to appropriate workflow
3. Workflows interact with state files, repositories, and node admin API
4. State is persisted in `~/.lpm/` directory

## Development Commands

### Build
```bash
# Build the lpm binary
./scripts/build.sh
# Output: ./build/lpm
```

### Run Tests
```bash
# Run all tests
go test ./...

# Run tests for specific package
go test ./workflow/...

# Run with coverage
go test -cover ./...
```

### Lint
```bash
# Run all linters (golangci-lint and license header check)
./scripts/lint.sh

# Run only golangci-lint
TESTS='golangci_lint' ./scripts/lint.sh

# Fix missing license headers (remove --check flag)
TESTS='license_header' ADDLICENSE_FLAGS="-v" ./scripts/lint.sh
```

### Install from Source
```bash
# Build and install
./scripts/build.sh
export PATH=$PWD/build:$PATH
```

## Testing Patterns

- Uses standard Go testing with `testing` package
- Mock generation via `github.com/golang/mock` (see `mock_*.go` files)
- Test assertions with `github.com/stretchr/testify`
- File system operations mocked with `github.com/spf13/afero`

## Configuration

LPM uses Viper for configuration management with these key paths:
- **LPM Directory**: `~/.lpm/` (state files, repositories, temp files)
- **Plugin Directory**: `$GOPATH/src/github.com/luxfi/node/build/plugins/`
- **Config File**: Optional via `--config-file` flag
- **Credentials**: Optional via `--credentials-file` for private repositories

## Important Files

- `constant/constants.go`: Core constants including default repository info
- `state/state.go`: State file structure and management
- `lpm/lpm.go`: Main application entry point and command orchestration
- `.golangci.yml`: Linter configuration

## Dependencies

- Local replacements in `go.mod`:
  - `github.com/luxfi/node => ../node`
  - `github.com/luxfi/geth => ../geth`

## Working with Private Repositories

Create a credentials file:
```yaml
username: <github-username>
password: <personal-access-token>
```

Use with commands:
```bash
lpm join-subnet --subnet=foobar --credentials-file=/path/to/token
```