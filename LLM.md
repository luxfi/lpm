# lpm

## Overview

**Note: This code is currently in Alpha. Proceed at your own risk.** `lpm` is a command-line tool to manage virtual machines binaries for 

## Package Information

- **Type**: go
- **Module**: github.com/luxfi/lpm
- **Repository**: github.com/luxfi/lpm

## Directory Structure

```
.
admin
cmd
config
docs
docs/.next
docs/.source
docs/components
docs/out
internal
internal/vmid
main             # main package — entry point (NOT cmd/lpm/)
state
types
util
workflow
```

## Key Files

- go.mod
- main/main.go     # binary entry point: `go build -o bin/lpm ./main`

## On-disk Manifest Format

Plugin definitions are YAML files under a repository checkout:

- `vms/<name>.yaml`   — wrapped under top-level `vm:` key
- `chains/<name>.yaml` — wrapped under top-level `subnet:` key (legacy
  wrapper key kept for backwards-compat with shipped manifests; the
  on-disk *directory* moved from `subnets/` to `chains/` in
  plugins-core v0.1.3)

`state/repository.go` enforces the wrapper. Unwrapped manifests are
rejected (was: silently zero-valued T returned, leading to install-vm
crashes — see CI-PREMORTEM.md §12a in luxcpp/cevm).

## Development

### Prerequisites

- Go 1.21+

### Build

```bash
go build -o bin/lpm ./main
```

### Test

```bash
go test -v ./...
```

## Integration with Lux Ecosystem

This package is part of the Lux blockchain ecosystem. See the main documentation at:
- GitHub: https://github.com/luxfi
- Docs: https://docs.lux.network

---

*Auto-generated for AI assistants. Last updated: 2025-12-24*
