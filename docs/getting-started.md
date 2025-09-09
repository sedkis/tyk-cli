---
layout: default
title: Getting Started
nav_order: 2
description: "Install and configure Tyk CLI to start managing your APIs"
---

# Getting Started with Tyk CLI
{: .no_toc }

This guide will help you install and configure the Tyk CLI to start managing your APIs.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

## Prerequisites

Before you begin, you'll need:

- Access to a Tyk Dashboard (local, cloud, or self-hosted)
- Your Dashboard URL, Auth Token, and Organization ID
- Go 1.21+ (if building from source)

## Installation

### Option 1: Homebrew (Recommended)

```bash
# Add the Tyk tap
brew tap sedkis/tyk

# Install the CLI
brew install tyk

# Verify installation
tyk --version
```

### Option 2: Direct Download

```bash
# Download and install the latest release
curl -L "https://github.com/sedkis/tyk-cli/releases/latest/download/tyk-cli_$(uname -s)_$(uname -m).tar.gz" | tar xz
sudo mv tyk /usr/local/bin/

# Verify installation
tyk --version
```

### Option 3: From Source

```bash
git clone https://github.com/sedkis/tyk-cli.git
cd tyk-cli
go build -o tyk .
sudo mv tyk /usr/local/bin/
```

## Initial Configuration

### Interactive Setup Wizard

The easiest way to get started is with the interactive setup wizard:

```bash
tyk init
```

This will guide you through:
1. Setting up your Dashboard URL
2. Configuring your Auth Token
3. Setting your Organization ID
4. Testing the connection
5. Creating additional environments (optional)

### Quick Setup

For a single environment setup:

```bash
tyk init --quick
```

### Offline Setup

If you want to configure without testing the connection:

```bash
tyk init --skip-test
```

## Finding Your Credentials

### Dashboard URL

- **Local Development**: `http://localhost:3000` (default)
- **Tyk Cloud**: `https://admin.cloud.tyk.io`
- **Self-hosted**: Your custom domain

### Auth Token & Organization ID

1. Log into your Tyk Dashboard
2. Go to **Users** → Your User Profile
3. Find **API Access Credentials**
4. Copy the **Auth Token** and **Organization ID**

## Verify Your Setup

Test your configuration with:

```bash
# Check current environment
tyk config current

# Test connection (should list your APIs)
tyk api list
```

## Basic Usage

### Environment Management

```bash
# List all environments
tyk config list

# Switch environments
tyk config use production

# View current environment details
tyk config current
```

### API Operations

```bash
# List all APIs
tyk api list

# Get specific API details
tyk api get <api-id>

# Create an API from OpenAPI spec
tyk api create --file my-api.yaml

# Update an existing API
tyk api apply --file my-api.yaml
```

## What's Next?

- Learn about [Configuration](configuration.md) options
- Check out the complete [API Reference](api-reference.md)
- Browse [Examples](examples/) for common workflows
- Read about [Contributing](../CONTRIBUTING.md) to the project

## Troubleshooting

### Common Issues

**"Connection refused" errors:**
- Check your Dashboard URL is correct
- Ensure the Tyk Dashboard is running
- Verify network connectivity

**"Unauthorized" errors:**
- Verify your Auth Token is valid
- Check your Organization ID is correct
- Ensure your user has proper permissions

**"Command not found" errors:**
- Verify `tyk` is in your PATH
- Try running `which tyk` to locate the binary

### Getting Help

- Run `tyk --help` for command help
- Use `tyk <command> --help` for specific command help
- Check [GitHub Issues](https://github.com/sedkis/tyk-cli/issues) for known problems
- Ask questions on the [Tyk Community Forum](https://community.tyk.io/)