# Contributing to CyberPatchMaker

Thank you for your interest in contributing to CyberPatchMaker!

## Getting Started

### Prerequisites

- Go 1.24.0 or later
- Git
- Make (optional, for building)

### Setting Up Development Environment

```bash
# Clone the repository
git clone https://github.com/cyberofficial/CyberPatchMaker.git
cd CyberPatchMaker

# Build the tools
go build -o patch-gen ./cmd/generator
go build -o patch-apply ./cmd/applier

# Run tests
.\advanced-test.ps1  # Windows PowerShell
# or
bash test.sh         # Linux/macOS (if available)
```

## Development Workflow

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

### 2. Make Your Changes

- Follow the existing code style
- Add comments for complex logic
- Update documentation if needed

### 3. Test Your Changes

```bash
# Run the test suite
.\advanced-test.ps1

# Test manually with your scenario
patch-gen --from-dir ./test-v1 --to-dir ./test-v2 --output ./test-patches
patch-apply --patch ./test-patches/*.patch --current-dir ./test-v1 --dry-run
```

### 4. Update Documentation

If your change affects:
- User-facing features → Update relevant guide in `docs/`
- API changes → Update `docs/architecture.md` or `docs/data-structures.md`
- CLI changes → Update `docs/cli-reference.md`

### 5. Submit a Pull Request

```bash
git push origin feature/your-feature-name
# Then create a PR on GitHub
```

## Code Style

### Formatting

- Use standard Go formatting: `go fmt ./...`
- Use meaningful variable names
- Keep lines under 120 characters when practical

### Comments

- Exported functions should have comments
- Complex logic should be explained
- Use `//` for single-line comments
- Use `/* */` for multi-line comments only when necessary

### Example

```go
// NewApplier creates a new patch applier instance
func NewApplier() *Applier {
    return &Applier{
        differ: differ.NewDiffer(),
    }
}
```

## Project Structure

```
CyberPatchMaker/
├── cmd/              # CLI tools (entry points)
├── internal/core/    # Business logic
│   ├── cache/        # Scan caching
│   ├── config/       # Configuration
│   ├── differ/       # Binary diffing
│   ├── manifest/     # Manifest operations
│   ├── patcher/      # Patch generation/application
│   ├── scanner/      # Directory scanning
│   └── version/      # Version management
├── pkg/utils/        # Shared utilities
└── docs/             # Documentation
```

## Adding Features

### Before You Start

1. Check existing issues and PRs
2. Discuss major changes in an issue first
3. Consider backwards compatibility

### Feature Checklist

- [ ] Implementation complete
- [ ] Tests pass
- [ ] Documentation updated
- [ ] Examples provided (if applicable)
- [ ] Backwards compatible (or breaking changes documented)

## Testing

### Running Tests

```bash
# Windows
.\advanced-test.ps1

# The test suite validates:
# - Build success
# - Patch generation
# - Patch application
# - Verification
# - Error handling
# - Backup/restore
```

### Writing Tests

When adding new functionality:
1. Add test cases to `advanced-test.ps1`
2. Test both success and failure cases
3. Test edge cases (empty files, large files, etc.)

## Documentation

### User Documentation

Located in `docs/` directory:
- Feature guides (e.g., `generator-guide.md`)
- Reference documentation (e.g., `cli-reference.md`)
- Technical guides (e.g., `architecture.md`)

### Updating Documentation

When changing behavior:
1. Update the relevant guide
2. Update examples if needed
3. Update `README.md` for major features

### Documentation Style

- Use clear, simple language
- Provide examples for complex features
- Include troubleshooting tips
- Cross-reference related topics

## Commit Messages

### Format

```
type(scope): brief description

Detailed description of changes.

Closes #issue-number
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Test changes
- `chore`: Maintenance tasks

### Examples

```
feat(patcher): add support for chunk sidecar files

Implement sidecar JSON files to track chunk metadata for
very large patches that exceed 4GB per part.

Closes #123

fix(scanner): handle symbolic links correctly

Prevent infinite loops when scanning directories with
symbolic links. Now skips symlinks during traversal.

docs(update): clarify large file handling behavior

Update large-file-handling.md to accurately reflect
full file replacement strategy for files >1GB.
```

## Pull Request Guidelines

### Title

Use the same format as commit messages:
```
feat(patcher): add support for chunk sidecar files
```

### Description

Include:
- What changes were made
- Why the changes were needed
- How users are affected
- Testing performed
- Breaking changes (if any)

### Review Process

1. Automated checks must pass
2. At least one maintainer approval required
3. Address review feedback
4. Squash commits if requested

## Release Process

Releases are managed by maintainers:
1. Update version in `internal/core/version/version.go`
2. Update CHANGELOG (if present)
3. Create git tag
4. Build release binaries
5. Create GitHub release

## Community Guidelines

### Code of Conduct

- Be respectful and constructive
- Welcome new contributors
- Focus on what is best for the community
- Show empathy toward other community members

### Getting Help

- Open an issue for bugs
- Start a discussion for questions
- Check existing documentation first

## Recognition

Contributors will be:
- Listed in CONTRIBUTORS file (if present)
- Mentioned in release notes (for significant contributions)
- Credited in commit history

Thank you for contributing to CyberPatchMaker!
