# Development Setup

Guide for developers contributing to CyberPatchMaker.

## Overview

This guide helps you set up a development environment to:
- Build CyberPatchMaker from source
- Run and debug the code
- Write and run tests
- Contribute changes back to the project

---

## Prerequisites

### Required

**1. Go 1.21 or later**

Download from [go.dev/dl](https://go.dev/dl/)

**Verify installation:**
```bash
go version
# Should show: go version go1.21.0 or later
```

**2. Git**

Download from [git-scm.com](https://git-scm.com/)

**Verify installation:**
```bash
git --version
# Should show: git version 2.x.x or later
```

---

### Recommended

**IDE/Editor:**
- [VS Code](https://code.visualstudio.com/) with Go extension
- [GoLand](https://www.jetbrains.com/go/) (commercial)
- [Vim](https://www.vim.org/) with vim-go plugin

**Tools:**
- [golangci-lint](https://golangci-lint.run/) - Code linting
- [delve](https://github.com/go-delve/delve) - Go debugger
- [Make](https://www.gnu.org/software/make/) - Build automation (optional)

---

## Getting the Code

### Clone Repository

**HTTPS:**
```bash
git clone https://github.com/cyberofficial/CyberPatchMaker.git
cd CyberPatchMaker
```

**SSH:**
```bash
git clone git@github.com:cyberofficial/CyberPatchMaker.git
cd CyberPatchMaker
```

---

### Fork for Contributing

**External contributors should fork first:**

1. Fork on GitHub: Click "Fork" button
2. Clone your fork:
```bash
git clone https://github.com/YOUR_USERNAME/CyberPatchMaker.git
cd CyberPatchMaker
```

3. Add upstream remote:
```bash
git remote add upstream https://github.com/cyberofficial/CyberPatchMaker.git
```

4. Verify remotes:
```bash
git remote -v
# origin    https://github.com/YOUR_USERNAME/CyberPatchMaker.git (fetch)
# origin    https://github.com/YOUR_USERNAME/CyberPatchMaker.git (push)
# upstream  https://github.com/cyberofficial/CyberPatchMaker.git (fetch)
# upstream  https://github.com/cyberofficial/CyberPatchMaker.git (push)
```

---

## Development Environment

### Download Dependencies

```bash
go mod download
```

This downloads all required Go modules listed in `go.mod`.

---

### IDE Setup

#### VS Code

**Install Go extension:**
1. Open VS Code
2. Install "Go" extension by Go Team at Google
3. Open Command Palette (Ctrl+Shift+P / Cmd+Shift+P)
4. Run: "Go: Install/Update Tools"
5. Select all tools and install

**Recommended settings** (`.vscode/settings.json`):
```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "go.formatTool": "gofmt",
  "editor.formatOnSave": true,
  "[go]": {
    "editor.codeActionsOnSave": {
      "source.organizeImports": true
    }
  }
}
```

---

#### GoLand

**Configuration:**
1. Open project directory
2. GoLand auto-detects Go SDK
3. Enable "Go modules" in Preferences → Go → Go Modules
4. Configure code style: Preferences → Editor → Code Style → Go

**Recommended:**
- Enable "Reformat on save"
- Enable "Optimize imports on save"
- Use GoLand's built-in test runner
- Use GoLand's debugger (excellent Go support)

---

#### Vim with vim-go

**Install vim-go:**
```vim
" Add to .vimrc
Plug 'fatih/vim-go', { 'do': ':GoUpdateBinaries' }
```

**Basic configuration:**
```vim
" .vimrc
let g:go_fmt_command = "gofmt"
let g:go_auto_type_info = 1
let g:go_highlight_functions = 1
let g:go_highlight_methods = 1
```

---

## Building from Source

### Build Generator

**Windows (PowerShell):**
```powershell
go build -o generator.exe ./cmd/generator
```

**Linux/macOS:**
```bash
go build -o generator ./cmd/generator
```

---

### Build Applier

**Windows (PowerShell):**
```powershell
go build -o applier.exe ./cmd/applier
```

**Linux/macOS:**
```bash
go build -o applier ./cmd/applier
```

---

### Build All

**Build both tools:**
```bash
go build ./cmd/...
```

---

### Cross-Compilation

**Build for Windows (from any OS):**
```bash
GOOS=windows GOARCH=amd64 go build -o generator.exe ./cmd/generator
GOOS=windows GOARCH=amd64 go build -o applier.exe ./cmd/applier
```

**Build for Linux (from any OS):**
```bash
GOOS=linux GOARCH=amd64 go build -o generator ./cmd/generator
GOOS=linux GOARCH=amd64 go build -o applier ./cmd/applier
```

**Build for macOS (from any OS):**
```bash
GOOS=darwin GOARCH=amd64 go build -o generator ./cmd/generator
GOOS=darwin GOARCH=amd64 go build -o applier ./cmd/applier
```

**All platforms at once:**
```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o build/windows/generator.exe ./cmd/generator
GOOS=windows GOARCH=amd64 go build -o build/windows/applier.exe ./cmd/applier

# Linux
GOOS=linux GOARCH=amd64 go build -o build/linux/generator ./cmd/generator
GOOS=linux GOARCH=amd64 go build -o build/linux/applier ./cmd/applier

# macOS
GOOS=darwin GOARCH=amd64 go build -o build/macos/generator ./cmd/generator
GOOS=darwin GOARCH=amd64 go build -o build/macos/applier ./cmd/applier
```

---

### Build Flags

**Optimized build (smaller, faster):**
```bash
go build -ldflags="-s -w" ./cmd/generator
# -s: strip symbol table
# -w: strip DWARF debug info
```

**Debug build (with symbols):**
```bash
go build -gcflags="all=-N -l" ./cmd/generator
# -N: disable optimizations
# -l: disable inlining
# Better for debugging
```

---

## Running Tests

### Full Test Suite

**Windows (PowerShell):**
```powershell
.\advanced-test.ps1
```

Test data is automatically generated on first run.

**Expected output (first run):**
```
Checking for test versions...
Version 1.0.0 not found, creating...
Version 1.0.0 created (3 files, 2 directories)
Version 1.0.1 not found, creating...
Version 1.0.1 created (4 files, 2 directories)
Version 1.0.2 not found, creating...
Version 1.0.2 created (11 files, 6 directories, 3 levels deep)
Created 3 test version(s)

=== Test Results ===
All 20 tests passed!
```

---

### Unit Tests

**Run all tests:**
```bash
go test ./...
```

**Run specific package:**
```bash
go test ./pkg/patcher
go test ./pkg/manifest
go test ./pkg/backup
```

**Verbose output:**
```bash
go test -v ./...
```

**Run specific test:**
```bash
go test -run TestApplyPatch ./pkg/patcher
go test -run TestGenerateManifest ./pkg/manifest
```

---

### Coverage

**Generate coverage:**
```bash
go test -cover ./...
```

**Coverage report:**
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Coverage by package:**
```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

---

### Benchmarks

**Run benchmarks:**
```bash
go test -bench=. ./...
```

**Benchmark specific function:**
```bash
go test -bench=BenchmarkDiff ./pkg/differ
```

**Benchmark with memory stats:**
```bash
go test -bench=. -benchmem ./...
```

---

## Code Structure

### Directory Layout

```
CyberPatchMaker/
├── cmd/
│   ├── generator/        # Generator CLI tool
│   │   └── main.go       # Entry point
│   └── applier/          # Applier CLI tool
│       └── main.go       # Entry point
│
├── pkg/
│   ├── patcher/          # Core patching logic
│   │   ├── applier.go    # Apply patches
│   │   ├── generator.go  # Generate patches
│   │   └── operations.go # Patch operations
│   │
│   ├── manifest/         # Manifest handling
│   │   ├── manifest.go   # Manifest structure
│   │   ├── generator.go  # Generate manifests
│   │   └── loader.go     # Load manifests
│   │
│   ├── diff/             # Binary diff algorithms
│   │   ├── bsdiff.go     # bsdiff implementation
│   │   └── differ.go     # Diff interface
│   │
│   ├── backup/           # Backup system
│   │   ├── backup.go     # Create backups
│   │   ├── restore.go    # Restore from backup
│   │   └── lifecycle.go  # Backup lifecycle
│   │
│   └── verify/           # Verification logic
│       ├── hash.go       # SHA-256 hashing
│       └── verify.go     # Pre/post verification
│
├── testdata/             # Test version files
│   ├── versions/         # Test versions
│   │   ├── 1.0.0/        # Version 1.0.0
│   │   ├── 1.0.1/        # Version 1.0.1
│   │   └── 1.0.2/        # Version 1.0.2
│   └── patches/          # Generated test patches
│
├── docs/                 # Documentation
│   ├── README.md         # Documentation hub
│   ├── quick-start.md    # Getting started
│   └── ...               # Other docs
│
├── go.mod                # Go module definition
├── go.sum                # Dependency checksums
├── advanced-test.ps1     # Comprehensive test suite (20 tests)
├── LICENSE               # Apache 2.0 license
└── README.md             # Project README
```

---

### Package Responsibilities

**cmd/generator:**
- Parse command-line arguments
- Coordinate patch generation workflow
- Handle user input/output

**cmd/applier:**
- Parse command-line arguments
- Coordinate patch application workflow
- Handle user input/output

**pkg/patcher:**
- Core patch generation logic
- Core patch application logic
- Patch operations (add, modify, delete)

**pkg/manifest:**
- Generate manifests for versions
- Load and parse manifests
- Manifest structure and validation

**pkg/diff:**
- Binary diff generation (bsdiff)
- Binary patch application (bspatch)
- Diff algorithm interface

**pkg/backup:**
- Create backups before patching
- Restore from backups on failure
- Manage backup lifecycle (3-operation system)

**pkg/verify:**
- Calculate SHA-256 checksums
- Pre-patch verification
- Post-patch verification

---

## Adding New Features

### Architecture Patterns

**Follow these patterns** (see [architecture.md](architecture.md)):

1. **Package-based organization** - Group related functionality
2. **Interface-based design** - Define interfaces for testability
3. **Error handling** - Always return errors, never panic
4. **Testing** - Write tests for all new code
5. **Documentation** - Document all exported functions

---

### Where to Add Code

**Adding new compression algorithm:**
- Add to `pkg/compress/` (new package)
- Implement `Compressor` interface
- Register in `cmd/generator/main.go`
- Update `docs/compression-guide.md`

**Adding new verification method:**
- Add to `pkg/verify/verify.go`
- Implement `Verifier` interface
- Call from `pkg/patcher/applier.go`
- Write tests in `pkg/verify/verify_test.go`

**Adding new command-line flag:**
- Add to `cmd/generator/main.go` or `cmd/applier/main.go`
- Parse in `parseFlags()` function
- Pass to relevant package
- Update help text and documentation

**Adding new patch operation:**
- Add to `pkg/patcher/operations.go`
- Implement `Operation` interface
- Handle in `pkg/patcher/applier.go`
- Write tests in `pkg/patcher/operations_test.go`

---

### Testing Requirements

**All new features must have:**

1. **Unit tests** - Test individual functions
2. **Integration tests** - Test end-to-end workflows
3. **Edge case tests** - Test boundary conditions
4. **Error tests** - Test error handling

**Example:**
```go
func TestNewFeature(t *testing.T) {
    // Setup
    input := "test input"
    expected := "expected output"
    
    // Execute
    result, err := NewFeature(input)
    
    // Verify
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result != expected {
        t.Errorf("expected %q, got %q", expected, result)
    }
}
```

---

### Documentation Requirements

**All new features must have:**

1. **Code comments** - Explain what and why
2. **Package documentation** - Overview in package comment
3. **Function documentation** - Describe parameters and return values
4. **User documentation** - Update relevant docs/ files

**Example:**
```go
// Package compress provides compression algorithms for patch files.
// It supports multiple algorithms including zstd, gzip, and brotli.
package compress

// Compressor is the interface for compression algorithms.
type Compressor interface {
    // Compress compresses the input data using the configured algorithm.
    // Returns the compressed data or an error if compression fails.
    Compress(data []byte) ([]byte, error)
    
    // Decompress decompresses the input data.
    // Returns the decompressed data or an error if decompression fails.
    Decompress(data []byte) ([]byte, error)
}
```

---

## Submitting Changes

### Fork → Branch → Commit → Push → PR

**1. Create feature branch:**
```bash
git checkout -b feature/add-brotli-compression
```

**Branch naming conventions:**
- `feature/description` - New features
- `fix/description` - Bug fixes
- `docs/description` - Documentation
- `test/description` - Test improvements
- `refactor/description` - Code refactoring

---

**2. Make changes:**
```bash
# Edit files
vim pkg/compress/brotli.go

# Run tests
go test ./...

# Format code
go fmt ./...
```

---

**3. Commit changes:**
```bash
git add pkg/compress/brotli.go
git commit -m "feat: add brotli compression support"
```

**Commit message format (conventional commits):**
```
<type>: <description>

[optional body]

[optional footer]
```

**Types:**
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation
- `test` - Tests
- `refactor` - Code refactoring
- `perf` - Performance improvement
- `chore` - Maintenance

**Example:**
```
feat: add brotli compression support

- Implement Brotli compressor
- Add compression level support (1-11)
- Add tests for compression and decompression
- Update documentation

Closes #123
```

---

**4. Push to your fork:**
```bash
git push origin feature/add-brotli-compression
```

---

**5. Create Pull Request:**

1. Go to GitHub repository
2. Click "New Pull Request"
3. Select your fork and branch
4. Fill out PR template:
   - Description of changes
   - Testing performed
   - Related issues
5. Submit PR

---

### Pull Request Template

```markdown
## Description
Brief description of what this PR does.

## Changes
- List of specific changes
- One change per bullet

## Testing
How was this tested?
- [ ] Unit tests added/updated
- [ ] Integration tests pass
- [ ] Manual testing performed

## Checklist
- [ ] Code follows project style
- [ ] Tests pass locally
- [ ] Documentation updated
- [ ] Commit messages follow convention

## Related Issues
Closes #123
```

---

### Code Review Process

1. **Automated checks run:**
   - Tests must pass
   - Linting must pass
   - Build must succeed

2. **Maintainer reviews code:**
   - Code quality
   - Test coverage
   - Documentation
   - Backward compatibility

3. **Address feedback:**
   - Make requested changes
   - Push updates to same branch
   - PR updates automatically

4. **Merge:**
   - Maintainer merges when approved
   - Branch is deleted
   - Changes appear in main branch

---

## Coding Standards

### Go Best Practices

**Follow [Effective Go](https://go.dev/doc/effective_go):**
- Use `gofmt` for formatting
- Write clear, idiomatic Go code
- Use short, descriptive names
- Handle all errors
- Write documentation comments

---

### Formatting

**Always run before committing:**
```bash
go fmt ./...
```

**Configure editor to format on save** (see IDE Setup above).

---

### Linting

**Install golangci-lint:**
```bash
# macOS/Linux
brew install golangci-lint

# Windows
scoop install golangci-lint

# Or download from https://golangci-lint.run/
```

**Run linter:**
```bash
golangci-lint run
```

**Fix auto-fixable issues:**
```bash
golangci-lint run --fix
```

---

### Error Handling

**Always handle errors:**
```go
// ✓ Good
file, err := os.Open("file.txt")
if err != nil {
    return fmt.Errorf("failed to open file: %w", err)
}
defer file.Close()

// ✗ Bad
file, _ := os.Open("file.txt")  // Ignoring error!
```

**Wrap errors with context:**
```go
if err := doSomething(); err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}
```

---

### Documentation

**All exported functions must have comments:**
```go
// GeneratePatch creates a patch file from source to target version.
// It compares the two versions and generates binary diffs for changed files.
// Returns an error if versions don't exist or generation fails.
func GeneratePatch(sourceVer, targetVer string) error {
    // Implementation
}
```

**Package comments:**
```go
// Package patcher provides core patching functionality.
// It handles patch generation, application, and verification.
package patcher
```

---

### Naming Conventions

**Follow Go standards:**
- `MixedCaps` for exported names
- `mixedCaps` for unexported names
- Short names for local variables (`i`, `err`, `buf`)
- Descriptive names for package-level variables

**Examples:**
```go
// ✓ Good
func GeneratePatch(src, dst string) error
var ErrInvalidVersion = errors.New("invalid version")
type PatchGenerator struct { ... }

// ✗ Bad
func generate_patch(src, dst string) error  // snake_case
var INVALID_VERSION = errors.New(...)       // SCREAMING_SNAKE_CASE
type patchGenerator struct { ... }          // unexported type
```

---

## Debugging

### Delve Debugger

**Install delve:**
```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

**Debug generator:**
```bash
dlv debug ./cmd/generator -- --versions-dir ./versions --new-version 1.0.1
```

**Debug tests:**
```bash
dlv test ./pkg/patcher
```

**Set breakpoints:**
```bash
(dlv) break patcher.go:42
(dlv) continue
(dlv) print localVar
(dlv) next
(dlv) step
(dlv) quit
```

---

### VS Code Debugging

**Configuration** (`.vscode/launch.json`):
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Generator",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/generator",
      "args": [
        "--versions-dir", "./testdata/versions",
        "--new-version", "1.0.1"
      ]
    }
  ]
}
```

**Usage:**
1. Set breakpoints (click left of line numbers)
2. Press F5 to start debugging
3. Use debug toolbar to step through code

---

### Profiling

**CPU profiling:**
```bash
go test -cpuprofile=cpu.prof ./pkg/patcher
go tool pprof cpu.prof
```

**Memory profiling:**
```bash
go test -memprofile=mem.prof ./pkg/patcher
go tool pprof mem.prof
```

**Analyze profile:**
```bash
(pprof) top        # Show top functions
(pprof) list func  # Show source for function
(pprof) web        # Open web visualization
```

---

## Common Development Tasks

### Task 1: Add New Compression Algorithm

**Steps:**

1. **Create package:**
```bash
mkdir -p pkg/compress
touch pkg/compress/brotli.go
```

2. **Implement interface:**
```go
package compress

type Brotli struct {
    level int
}

func (b *Brotli) Compress(data []byte) ([]byte, error) {
    // Implementation
}

func (b *Brotli) Decompress(data []byte) ([]byte, error) {
    // Implementation
}
```

3. **Add tests:**
```go
func TestBrotliCompress(t *testing.T) {
    // Test implementation
}
```

4. **Register in generator:**
```go
// cmd/generator/main.go
switch compressionType {
case "brotli":
    compressor = &compress.Brotli{Level: level}
}
```

5. **Update documentation:**
- docs/compression-guide.md
- Add usage examples

---

### Task 2: Add New Command-Line Flag

**Steps:**

1. **Define flag:**
```go
// cmd/generator/main.go
var parallel = flag.Bool("parallel", false, "Enable parallel processing")
```

2. **Parse flag:**
```go
flag.Parse()
```

3. **Use flag:**
```go
if *parallel {
    // Enable parallel mode
}
```

4. **Update help:**
```go
flag.Usage = func() {
    fmt.Fprintf(os.Stderr, "Usage: generator [options]\n")
    fmt.Fprintf(os.Stderr, "Options:\n")
    flag.PrintDefaults()
}
```

5. **Update documentation:**
- docs/generator-guide.md
- docs/cli-reference.md

---

### Task 3: Improve Performance

**Steps:**

1. **Benchmark current performance:**
```bash
go test -bench=. -benchmem ./pkg/patcher > before.txt
```

2. **Profile the code:**
```bash
go test -cpuprofile=cpu.prof ./pkg/patcher
go tool pprof -http=:8080 cpu.prof
```

3. **Identify bottlenecks:**
- Look for hot functions in profile
- Check for unnecessary allocations

4. **Optimize:**
```go
// Example: Reduce allocations
// Before:
for i := 0; i < n; i++ {
    data := make([]byte, size)  // Allocates every iteration
}

// After:
data := make([]byte, size)
for i := 0; i < n; i++ {
    // Reuse buffer
}
```

5. **Benchmark again:**
```bash
go test -bench=. -benchmem ./pkg/patcher > after.txt
benchcmp before.txt after.txt
```

6. **Verify correctness:**
```powershell
go test ./...
.\advanced-test.ps1
```

---

## Contributing Guidelines

### Before You Start

**Check existing issues:**
- Search for similar issues/PRs
- Comment on issue to claim it
- Discuss approach with maintainers

**For large changes:**
- Open issue first to discuss
- Get feedback on approach
- Avoid surprises in PR review

---

### Small, Focused PRs

**Prefer small PRs that:**
- Address one issue/feature
- Are easy to review
- Have clear purpose

**Split large changes into:**
- Multiple smaller PRs
- Incremental improvements
- Logical chunks

---

### Tests Required

**All PRs must:**
- Include tests for new code
- Pass all existing tests
- Maintain or improve coverage

**Test checklist:**
- [ ] Unit tests added
- [ ] Integration tests added (if needed)
- [ ] Edge cases tested
- [ ] Error cases tested
- [ ] All tests pass locally

---

### Documentation Required

**All PRs must:**
- Update code comments
- Update relevant docs/ files
- Update CLI help text (if needed)
- Include usage examples

**Documentation checklist:**
- [ ] Code comments added
- [ ] Package documentation updated
- [ ] User documentation updated
- [ ] Examples added

---

### Backward Compatibility

**Maintain backward compatibility:**
- Don't break existing APIs
- Don't break patch file format
- Don't break command-line interface
- Document breaking changes if unavoidable

---

### Security Considerations

**Think about security:**
- Validate all user input
- Check file paths (no path traversal)
- Verify checksums
- Handle errors safely

---

## Getting Help

### Resources

**Documentation:**
- This guide (development-setup.md)
- [Architecture](architecture.md)
- [Contributing](../CONTRIBUTING.md) (if exists)

**Community:**
- GitHub Issues - Report bugs, request features
- GitHub Discussions - Ask questions
- Pull Requests - Code review and feedback

**Go Resources:**
- [Go Documentation](https://go.dev/doc/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go by Example](https://gobyexample.com/)

---

## Related Documentation

- [Architecture](architecture.md) - System design
- [Testing Guide](testing-guide.md) - Test suite details
- [Generator Guide](generator-guide.md) - Generator usage
- [Applier Guide](applier-guide.md) - Applier usage
- [Quick Start](quick-start.md) - Getting started
