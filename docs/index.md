# Tyk CLI Documentation

A powerful command-line interface for managing Tyk APIs and configurations. Built to streamline API lifecycle management with OpenAPI Specification (OAS) support.

## Quick Navigation

- **[Getting Started](getting-started.md)** - Installation and initial setup
- **[Configuration](configuration.md)** - Environment and config management
- **[API Reference](api-reference.md)** - Complete command reference
- **[Examples](examples/)** - Real-world usage examples
- **[Contributing](../CONTRIBUTING.md)** - How to contribute to the project

## Features

- 🚀 **Interactive Setup Wizard** - Get started quickly with guided configuration
- 🌍 **Unified Config/Environment System** - Named environments with seamless switching
- 📝 **OpenAPI First** - Native support for OAS 3.0 specifications  
- 🔧 **Flexible Configuration** - Environment variables, unified config file, or CLI flags
- 🎨 **Beautiful CLI** - Colorful, intuitive command-line experience
- ✅ **Comprehensive Testing** - >80% test coverage with live environment validation

## Quick Start

### Installation

#### Homebrew (Recommended)

```bash
# Add the Tyk tap
brew tap sedkis/tyk

# Install the CLI
brew install tyk
```

### Initialize Configuration

```bash
# Run the interactive setup wizard
tyk init

# Quick setup for single environment
tyk init --quick
```

## Support

- 🐛 **Issues**: [GitHub Issues](https://github.com/tyktech/tyk-cli/issues)
- 💬 **Community**: [Tyk Community Forum](https://community.tyk.io/)
- 📖 **Main Repository**: [tyk-cli on GitHub](https://github.com/tyktech/tyk-cli)