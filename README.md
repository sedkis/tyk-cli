# Tyk CLI

A powerful command-line interface for managing Tyk APIs and configurations. Built to streamline API lifecycle management with OpenAPI Specification (OAS) support.

[![Go Version](https://img.shields.io/github/go-mod/go-version/tyktech/tyk-cli)](https://golang.org/)
[![License](https://img.shields.io/github/license/tyktech/tyk-cli)](LICENSE)
[![Build Status](https://img.shields.io/github/workflow/status/tyktech/tyk-cli/CI)](https://github.com/tyktech/tyk-cli/actions)

## ✨ Features

- 🚀 **Interactive Setup Wizard** - Get started quickly with guided configuration
- 🌍 **Unified Config/Environment System** - Named environments with seamless switching
- 📝 **OpenAPI First** - Native support for OAS 3.0 specifications  
- 🔧 **Flexible Configuration** - Environment variables, unified config file, or CLI flags
- 🎨 **Beautiful CLI** - Colorful, intuitive command-line experience
- ✅ **Comprehensive Testing** - >80% test coverage with live environment validation

## 🚀 Quick Start

### Installation

#### Homebrew (Recommended)

```bash
# Add the Tyk tap
brew tap tyktech/tyk

# Install the CLI
brew install tyk
```

#### Direct Download

```bash
# Download and install the latest release
curl -L "https://github.com/tyktech/tyk-cli/releases/latest/download/tyk-cli_$(uname -s)_$(uname -m).tar.gz" | tar xz
sudo mv tyk /usr/local/bin/

# Verify installation
tyk --version
```

#### From Source

```bash
git clone https://github.com/tyktech/tyk-cli.git
cd tyk-cli
go build -o tyk .
sudo mv tyk /usr/local/bin/
```

### Initialize Configuration

```bash
# Run the interactive setup wizard
tyk init

# Quick setup for single environment
tyk init --quick

# Skip connection testing (for offline setup)
tyk init --skip-test
```

## 📋 Commands

### Configuration/Environment Management
```bash
# Interactive setup wizard
tyk init                    # Full setup wizard with multiple environments
tyk init --quick           # Quick single-environment setup  

# Environment management (unified with configuration)
tyk config list            # List all environments
tyk config use staging     # Switch to staging environment
tyk config current         # Show current environment
tyk config set dashboard-url https://api.tyk.io  # Update current environment
```

### API Management
```bash
# GitOps-style declarative management
tyk api apply --file api.yaml      # Upsert API (create/update based on ID)
tyk api apply --file new-api.yaml --create  # Create new API via apply

# Explicit CRUD operations
tyk api get <api-id>                # Get API details
tyk api get <api-id> --version-name v2  # Get specific version
tyk api create --file api.yaml     # Always create new API
tyk api update --api-id <id> --file updated.yaml  # Always update existing
tyk api delete <api-id>             # Delete API (with confirmation)
tyk api delete <api-id> --yes       # Delete without confirmation

# Utilities (Phase 3)
tyk api convert --file api.yaml --format apidef  # Convert OAS to Tyk format
```

## ⚙️ Configuration

The Tyk CLI uses a unified environment/configuration system with the following precedence (highest to lowest):

1. **Command-line flags** (`--dash-url`, `--auth-token`, `--org-id`)
2. **Environment variables** (`TYK_DASH_URL`, `TYK_AUTH_TOKEN`, `TYK_ORG_ID`)
3. **Named environments in config file** (`~/.config/tyk/cli.toml`)

Each "environment" is simply a named set of configuration values.

### Environment Variables

```bash
export TYK_DASH_URL=http://localhost:3000
export TYK_AUTH_TOKEN=your-api-token
export TYK_ORG_ID=your-org-id
```

### Config File (Unified Environment System)

Configuration is automatically saved to `~/.config/tyk/cli.toml`:

```toml
default_environment = "dev"

[environments.dev]
dashboard_url = "http://localhost:3000"
auth_token = "dev-api-token" 
org_id = "dev-org-id"

[environments.staging]
dashboard_url = "https://staging.tyk.io"
auth_token = "staging-token"
org_id = "staging-org-id"

[environments.production]
dashboard_url = "https://api.yourcompany.com"
auth_token = "prod-token"
org_id = "prod-org-id"
```

**Note**: Environments ARE the configuration system - no redundancy between "config" and "environments".

## 🔍 Finding Your Credentials

### Dashboard URL
- **Local Development**: `http://localhost:3000` (default)
- **Tyk Cloud**: `https://admin.cloud.tyk.io`
- **Self-hosted**: Your custom domain

### Auth Token
1. Log into your Tyk Dashboard
2. Go to **Users** → Your User Profile
3. Find **API Access Credentials**
4. Copy the **Auth Token**

### Organization ID
- Found in your Dashboard URL: `/a/{org-id}/`
- Or in Dashboard: **System Management** → **Organizations**

## 🛠️ Development

### Prerequisites

- Go 1.21+
- Make (optional)

### Setup

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

# Build with make (if available)
make build
make test
```

### Project Structure

```
tyk-cli/
├── cmd/           # CLI commands and subcommands
├── internal/      # Internal packages
│   ├── config/    # Configuration management
│   ├── client/    # HTTP client for Tyk Dashboard API
│   └── util/      # Utilities and helpers
├── pkg/           # Public packages (if any)
├── test/          # Integration tests
└── docs/          # Documentation
```

## 🗺️ Roadmap

### 🏗️ Foundation
- Multi-environment configuration system
- HTTP client with Tyk Dashboard integration
- Interactive CLI framework with colors
- Comprehensive test suite (>80% coverage)

### 🎯 Core API Management
- `tyk api get` - Retrieve API details with optional version selection
- `tyk api create` - Create new APIs from OAS files (explicit creation)
- `tyk api apply` - Declarative upsert based on API ID presence
- `tyk api update` - Update existing APIs explicitly
- `tyk api delete` - Delete APIs with confirmation prompts
- Full JSON output support with `--json` flag
- Proper exit code system (0=success, 2=bad args, 3=not found, 4=conflict)

### 🔧 Enhanced Features
- `tyk api convert` - Convert between OAS and Tyk API definition formats
- Enhanced error handling and user experience improvements
- Advanced JSON output formatting

### 🚀 Future Enhancements
- API versioning commands (`versions list/create/switch-default`)
- API validation and linting
- GitOps diff functionality

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Common Development Tasks

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Lint code (if golangci-lint is installed)
golangci-lint run

# Format code
go fmt ./...
```

## 📄 License

This project is licensed under the [MIT License](LICENSE).

## 🆘 Support

- 📖 **Documentation**: Check the [Getting Started Guide](GETTING_STARTED.md)
- 🐛 **Issues**: [GitHub Issues](https://github.com/tyktech/tyk-cli/issues)
- 💬 **Community**: [Tyk Community Forum](https://community.tyk.io/)

## 🙏 Acknowledgments

Built with:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management
- [Survey](https://github.com/AlecAivazis/survey) - Interactive prompts
- [Color](https://github.com/fatih/color) - Terminal colors
