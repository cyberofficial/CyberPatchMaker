# Development Setup

## Prerequisites

- Go 1.24.0 or later
- Git

## Getting the Code

```bash
git clone https://github.com/cyberofficial/CyberPatchMaker.git
cd CyberPatchMaker
go mod download
```

## Building

```bash
go build -o patch-gen.exe ./cmd/generator
go build -o patch-apply.exe ./cmd/applier

# Cross-compile
GOOS=windows GOARCH=amd64 go build -o patch-gen.exe ./cmd/generator
GOOS=linux GOARCH=amd64 go build -o patch-gen ./cmd/generator

# Optimized (smaller binary)
go build -ldflags="-s -w" -o patch-gen.exe ./cmd/generator
```

## Versioned Build

```powershell
.\build.ps1             # build to dist/<version>/
.\build.ps1 -i          # bump patch version then build
.\build.ps1 -ii         # bump minor version then build
.\build.ps1 -iii        # bump major version then build
.\build.ps1 -Clean      # clean dist before build
```

Version constants are in `internal/core/version/version.go`.

## Running Tests

```powershell
.\advanced-test.ps1                          # 59 standard tests
.\advanced-test.ps1 -run1gbtest             # + large patch test
.\advanced-test.ps1 -runlargefile           # + chunked processing test
.\advanced-test.ps1 -run1gbtest -runlargefile  # all tests
```

Test data is auto-generated on first run. On subsequent runs, previous test data is cleaned up automatically.

## Code Organization

```
cmd/generator/main.go    -- CLI patch generator
cmd/applier/main.go      -- CLI patch applier
internal/core/           -- business logic (patcher, scanner, manifest, version, cache, differ, config)
pkg/utils/               -- shared types and utilities (types, checksum, fileops, compress, patch_io)
```

## Adding Features

- **User-facing**: Add to `cmd/` layer
- **Business logic**: Add to appropriate `internal/core/` module
- **Shared utility**: Add to `pkg/utils/`
- **Core types**: Update `pkg/utils/types.go`

## Contributing

1. Create a feature branch: `git checkout -b feature/description`
2. Make changes, run `go fmt ./...`, run `.\advanced-test.ps1`
3. Use conventional commits: `feat:`, `fix:`, `docs:`, `test:`, `refactor:`, `perf:`, `chore:`
4. Submit PR against `main`

All PRs must pass the test suite. Documentation should be updated for any user-facing changes.
