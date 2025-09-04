# Contributing to Tyk CLI

Thank you for your interest in contributing to Tyk CLI! This guide will help you get started with contributing to the project.

## 🚀 Quick Start

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/your-username/tyk-cli.git
   cd tyk-cli
   ```
3. **Install dependencies**:
   ```bash
   go mod download
   ```
4. **Make your changes**
5. **Test your changes**:
   ```bash
   go test ./...
   ```
6. **Submit a pull request**

## 📋 Development Setup

### Prerequisites

- **Go 1.21+** - [Installation guide](https://golang.org/doc/install)
- **Git** - For version control
- **Make** (optional) - For using Makefile commands

### Local Development

```bash
# Clone the repository
git clone https://github.com/tyktech/tyk-cli.git
cd tyk-cli

# Install dependencies
go mod download

# Build the CLI
go build -o tyk .

# Run tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Format code
go fmt ./...

# Run linter (if golangci-lint is installed)
golangci-lint run
```

### Using Make (Optional)

```bash
# Build the project
make build

# Run tests
make test

# Run tests with coverage
make test-coverage

# Format and lint code
make fmt
make lint

# Clean build artifacts
make clean
```

## 🏗️ Project Structure

```
tyk-cli/
├── cmd/                    # CLI commands and subcommands
│   ├── root.go            # Root command and global flags
│   ├── init.go            # Interactive setup wizard
│   ├── api/               # API management commands
│   │   ├── apply.go       # Apply API configurations
│   │   ├── create.go      # Create new APIs
│   │   ├── get.go         # Get API details
│   │   ├── update.go      # Update existing APIs
│   │   └── delete.go      # Delete APIs
│   └── config/            # Configuration commands
│       ├── list.go        # List environments
│       ├── use.go         # Switch environments
│       └── current.go     # Show current environment
├── internal/              # Internal packages
│   ├── config/            # Configuration management
│   │   ├── config.go      # Config file operations
│   │   └── environment.go # Environment management
│   ├── client/            # HTTP client for Tyk Dashboard API
│   │   ├── client.go      # Base client implementation
│   │   └── api.go         # API-specific operations
│   └── util/              # Utilities and helpers
│       ├── output.go      # Output formatting
│       └── validation.go  # Input validation
├── pkg/                   # Public packages (if any)
├── test/                  # Integration tests
│   ├── fixtures/          # Test fixtures and sample data
│   └── integration/       # Integration test suites
├── docs/                  # Documentation
└── scripts/               # Build and deployment scripts
```

## 🎯 Types of Contributions

We welcome various types of contributions:

### 🐛 Bug Reports

When filing a bug report, please include:

- **Clear description** of the issue
- **Steps to reproduce** the problem
- **Expected vs actual behavior**
- **Environment details** (OS, Go version, CLI version)
- **Relevant logs or error messages**

Use the bug report template:

```markdown
**Describe the bug**
A clear description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Run command '...'
2. See error

**Expected behavior**
What you expected to happen.

**Environment:**
- OS: [e.g., macOS 13.0]
- Go version: [e.g., 1.21.0]
- CLI version: [e.g., v1.0.0]

**Additional context**
Any other context about the problem.
```

### ✨ Feature Requests

For new features, please:

- **Check existing issues** to avoid duplicates
- **Describe the use case** and why it's needed
- **Propose an implementation** approach (optional)
- **Consider breaking changes** and compatibility

### 📝 Documentation

Documentation improvements are always welcome:

- Fix typos or unclear explanations
- Add missing documentation
- Improve examples
- Update outdated information

### 💻 Code Contributions

When contributing code:

- **Follow Go conventions** and idioms
- **Write tests** for new functionality
- **Update documentation** if needed
- **Keep changes focused** (one feature/fix per PR)

## 🔧 Development Guidelines

### Code Style

- **Follow Go standards**: Use `go fmt` and `golangci-lint`
- **Use meaningful names**: Variables, functions, and packages should have descriptive names
- **Add comments**: Especially for exported functions and complex logic
- **Handle errors**: Always check and handle errors appropriately

### Testing

- **Write unit tests** for new functions
- **Add integration tests** for new commands
- **Maintain test coverage** above 80%
- **Use table-driven tests** where appropriate

Example test structure:
```go
func TestSomeFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {"valid input", "test", "expected", false},
        {"invalid input", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := SomeFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("SomeFunction() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if result != tt.expected {
                t.Errorf("SomeFunction() = %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### Commit Guidelines

We follow conventional commits:

- **feat**: New features
- **fix**: Bug fixes
- **docs**: Documentation changes
- **style**: Code style changes (no functional changes)
- **refactor**: Code refactoring
- **test**: Adding or fixing tests
- **chore**: Maintenance tasks

Examples:
```
feat: add api versioning support
fix: resolve config file permission issue
docs: update installation instructions
test: add integration tests for api create command
```

## 🔄 Pull Request Process

### Before Submitting

1. **Create a branch** from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** with clear, focused commits

3. **Run tests** and ensure they pass:
   ```bash
   go test ./...
   ```

4. **Run linter** and fix any issues:
   ```bash
   golangci-lint run
   ```

5. **Update documentation** if needed

### Submitting the PR

1. **Push your branch** to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create a pull request** on GitHub

3. **Fill out the PR template** completely:

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Documentation update
- [ ] Refactoring

## Testing
- [ ] Tests pass locally
- [ ] Added new tests if needed
- [ ] Manual testing completed

## Checklist
- [ ] Code follows project guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No breaking changes (or documented)
```

### Review Process

1. **Automated checks** must pass (tests, linting)
2. **Code review** by maintainers
3. **Address feedback** promptly
4. **Maintainer approval** required for merge

## 🧪 Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test ./internal/config

# Run with verbose output
go test -v ./...
```

### Integration Tests

Integration tests require a running Tyk Dashboard:

```bash
# Set up test environment
export TYK_DASH_URL=http://localhost:3000
export TYK_AUTH_TOKEN=your-test-token
export TYK_ORG_ID=your-test-org

# Run integration tests
go test ./test/integration/...
```

### Test Coverage

We aim for >80% test coverage. Check coverage with:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## 📚 Documentation

### Writing Documentation

- Use **clear, concise language**
- Include **practical examples**
- **Test all code examples**
- Follow **existing documentation style**

### Documentation Types

- **API Reference**: Complete command documentation
- **Guides**: Step-by-step tutorials
- **Examples**: Real-world usage scenarios
- **Contributing**: This guide

### Building Documentation

Documentation is built using Jekyll for GitHub Pages:

```bash
# Install Jekyll (one time)
gem install bundler jekyll

# Navigate to docs directory
cd docs

# Install dependencies
bundle install

# Serve locally
bundle exec jekyll serve
```

## 🎉 Recognition

Contributors are recognized in:

- **Release notes** for significant contributions
- **Contributors section** in README
- **GitHub contributors** page

## 📞 Getting Help

If you need help:

- **GitHub Issues**: Ask questions or report problems
- **GitHub Discussions**: Community discussions and ideas
- **Tyk Community Forum**: General Tyk-related questions

## 🏷️ Release Process

For maintainers:

1. **Update version** in relevant files
2. **Create release notes**
3. **Tag release**: `git tag -a v1.2.3 -m "Release v1.2.3"`
4. **Push tag**: `git push origin v1.2.3`
5. **GitHub Actions** handles the rest

## 📜 License

By contributing to Tyk CLI, you agree that your contributions will be licensed under the [MIT License](LICENSE).

---

Thank you for contributing to Tyk CLI! Your efforts help make API management easier for everyone. 🚀