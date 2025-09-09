---
layout: default
title: API Reference
nav_order: 4
description: "Complete reference for all Tyk CLI commands and options"
---

# API Reference
{: .no_toc }

Complete reference for all Tyk CLI commands and options.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

## Global Flags

These flags are available for all commands:

| Flag | Environment Variable | Description |
|------|---------------------|-------------|
| `--auth-token` | `TYK_AUTH_TOKEN` | Dashboard API auth token |
| `--dash-url` | `TYK_DASH_URL` | Tyk Dashboard URL |
| `--org-id` | `TYK_ORG_ID` | Organization ID |
| `--json` | - | Output in JSON format |
| `-h, --help` | - | Help for command |
| `-v, --version` | - | Show version |

## Commands

### `tyk init`

Interactive setup wizard to configure Tyk CLI quickly.

```bash
tyk init [flags]
```

**Flags:**
- `--quick` - Quick setup (single environment)
- `--skip-test` - Skip connection testing

**Examples:**
```bash
# Full interactive setup
tyk init

# Quick single-environment setup
tyk init --quick

# Setup without connection testing
tyk init --skip-test
```

---

### `tyk config`

Unified environment and configuration management.

#### `tyk config list`

List all configured environments.

```bash
tyk config list
```

#### `tyk config current`

Show current active environment details.

```bash
tyk config current
```

#### `tyk config use`

Switch to a different environment.

```bash
tyk config use <environment-name>
```

**Examples:**
```bash
tyk config use staging
tyk config use production
```

#### `tyk config add`

Add a new environment.

```bash
tyk config add <name> --dashboard-url <url> --auth-token <token> --org-id <id>
```

**Examples:**
```bash
tyk config add dev --dashboard-url http://localhost:3000 --auth-token dev-token --org-id dev-org
tyk config add staging --dashboard-url https://staging.tyk.io --auth-token staging-token --org-id staging-org
```

#### `tyk config set`

Update current environment configuration.

```bash
tyk config set <key> <value>
```

**Examples:**
```bash
tyk config set dashboard-url https://api.tyk.io
tyk config set auth-token new-token
tyk config set org-id new-org-id
```

#### `tyk config remove`

Remove an environment.

```bash
tyk config remove <environment-name>
```

---

### `tyk api`

Commands for managing OAS-native APIs in Tyk Dashboard.

#### `tyk api get`

Get an API by ID.

```bash
tyk api get <api-id> [flags]
```

**Flags:**
- `--version-name` - Get specific API version
- `--oas-only` - Return only the OpenAPI specification without Tyk extensions

**Examples:**
```bash
# Get API details
tyk api get my-api-123

# Get specific version
tyk api get my-api-123 --version-name v2

# Get clean OpenAPI spec only
tyk api get my-api-123 --oas-only

# Get API in JSON format
tyk api get my-api-123 --json
```

#### `tyk api import-oas`

Import clean OpenAPI spec to create new API.

```bash
tyk api import-oas [flags]
```

**Flags:**
- `--file` - Path to OpenAPI specification file
- `--url` - URL to OpenAPI specification

**Examples:**
```bash
# Import from local file
tyk api import-oas --file api.yaml

# Import from URL
tyk api import-oas --url https://api.example.com/openapi.json

# Import and output JSON
tyk api import-oas --file api.yaml --json
```

#### `tyk api update-oas`

Update existing API's OpenAPI spec only.

```bash
tyk api update-oas <api-id> [flags]
```

**Flags:**
- `--file` - Path to OpenAPI specification file
- `--url` - URL to OpenAPI specification

**Examples:**
```bash
# Update from local file
tyk api update-oas my-api-123 --file updated-api.yaml

# Update from URL
tyk api update-oas my-api-123 --url https://api.example.com/openapi.json
```

#### `tyk api apply`

Apply API configuration (declarative upsert - creates or updates based on API ID).

```bash
tyk api apply --file <path> [flags]
```

**Flags:**
- `--file` - Path to OAS specification file
- `--create` - Force creation of new API via apply

**Examples:**
```bash
# Apply API (create or update based on ID)
tyk api apply --file api.yaml

# Force create new API via apply
tyk api apply --file new-api.yaml --create
```

#### `tyk api delete`

Delete an API by ID.

```bash
tyk api delete <api-id> [flags]
```

**Flags:**
- `--yes` - Skip confirmation prompt

**Examples:**
```bash
# Delete with confirmation
tyk api delete my-api-123

# Delete without confirmation
tyk api delete my-api-123 --yes
```

#### `tyk api convert`

Convert OAS to Tyk API definition format.

```bash
tyk api convert --file <path> --format <format>
```

**Flags:**
- `--file` - Path to OAS specification file
- `--format` - Output format (apidef)

**Examples:**
```bash
# Convert OAS to Tyk API definition
tyk api convert --file api.yaml --format apidef
```

---

### `tyk completion`

Generate autocompletion script for your shell.

```bash
tyk completion <shell>
```

**Supported shells:** bash, zsh, fish, powershell

**Examples:**
```bash
# Bash completion
tyk completion bash > /etc/bash_completion.d/tyk

# Zsh completion
tyk completion zsh > "${fpath[1]}/_tyk"

# Fish completion
tyk completion fish > ~/.config/fish/completions/tyk.fish
```

---

### `tyk help`

Get help about any command.

```bash
tyk help [command]
```

**Examples:**
```bash
tyk help
tyk help api
tyk help api create
```

## Configuration Precedence

The CLI uses the following configuration precedence (highest to lowest):

1. **Command-line flags** (`--dash-url`, `--auth-token`, `--org-id`)
2. **Environment variables** (`TYK_DASH_URL`, `TYK_AUTH_TOKEN`, `TYK_ORG_ID`)
3. **Named environments in config file** (`~/.config/tyk/cli.toml`)

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | General error |
| 2 | Configuration error |
| 3 | API error |
| 4 | File error |

## Output Formats

Most commands support both human-readable and JSON output:

```bash
# Human-readable (default)
tyk api get my-api

# JSON output
tyk api get my-api --json
```