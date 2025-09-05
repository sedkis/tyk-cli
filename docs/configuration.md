---
layout: default
title: Configuration
nav_order: 3
description: "Environment management and configuration options for Tyk CLI"
---

# Configuration Guide
{: .no_toc }

The Tyk CLI uses a unified environment/configuration system that makes it easy to work with multiple Tyk instances (dev, staging, production) and switch between them seamlessly.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

## Overview

In the unified approach, **environments ARE the configuration system**. Each environment contains:
- Dashboard URL
- Auth Token  
- Organization ID

This means you can have multiple named configurations and easily switch between them.

## Configuration Precedence

The CLI resolves configuration in the following order (highest to lowest priority):

1. **Command-line flags** (`--dash-url`, `--auth-token`, `--org-id`)
2. **Environment variables** (`TYK_DASH_URL`, `TYK_AUTH_TOKEN`, `TYK_ORG_ID`)
3. **Named environments in config file** (`~/.config/tyk/cli.toml`)

## Config File Location

The configuration file is automatically created at:
- **macOS/Linux**: `~/.config/tyk/cli.toml`
- **Windows**: `%APPDATA%\tyk\cli.toml`

## Environment Management

### Interactive Setup

The easiest way to configure environments is with the interactive wizard:

```bash
# Full setup with multiple environments
tyk init

# Quick single environment setup
tyk init --quick
```

### Manual Environment Management

#### List Environments

```bash
tyk config list
```

Example output:
```
Available environments:
* dev        http://localhost:3000
  staging    https://staging-api.company.com
  production https://api.company.com

Current environment: dev
```

#### Add New Environment

```bash
tyk config add <name> --dashboard-url <url> --auth-token <token> --org-id <id>
```

Examples:
```bash
# Add development environment
tyk config add dev \
  --dashboard-url http://localhost:3000 \
  --auth-token eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9... \
  --org-id 507f1f77bcf86cd799439011

# Add staging environment
tyk config add staging \
  --dashboard-url https://staging.tyk.company.com \
  --auth-token eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9... \
  --org-id 507f1f77bcf86cd799439012

# Add production environment
tyk config add production \
  --dashboard-url https://api.company.com \
  --auth-token eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9... \
  --org-id 507f1f77bcf86cd799439013
```

#### Switch Environments

```bash
tyk config use <environment-name>
```

Examples:
```bash
tyk config use staging
tyk config use production
tyk config use dev
```

#### View Current Environment

```bash
tyk config current
```

Example output:
```
Current environment: staging
Dashboard URL: https://staging.tyk.company.com
Organization ID: 507f1f77bcf86cd799439012
Auth Token: eyJ0eXA... (truncated)
```

#### Update Environment Settings

```bash
tyk config set <key> <value>
```

Examples:
```bash
# Update dashboard URL for current environment
tyk config set dashboard-url https://new-api.company.com

# Update auth token
tyk config set auth-token eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...

# Update organization ID
tyk config set org-id 507f1f77bcf86cd799439999
```

#### Remove Environment

```bash
tyk config remove <environment-name>
```

## Config File Format

The configuration file uses TOML format:

```toml
default_environment = "dev"

[environments.dev]
dashboard_url = "http://localhost:3000"
auth_token = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9..."
org_id = "507f1f77bcf86cd799439011"

[environments.staging]
dashboard_url = "https://staging.tyk.company.com"
auth_token = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9..."
org_id = "507f1f77bcf86cd799439012"

[environments.production]
dashboard_url = "https://api.company.com"
auth_token = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9..."
org_id = "507f1f77bcf86cd799439013"
```

## Environment Variables

You can also use environment variables for configuration:

```bash
export TYK_DASH_URL=http://localhost:3000
export TYK_AUTH_TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...
export TYK_ORG_ID=507f1f77bcf86cd799439011
```

## Command-Line Flags

Override any configuration with command-line flags:

```bash
tyk api get my-api \
  --dash-url https://api.company.com \
  --auth-token eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9... \
  --org-id 507f1f77bcf86cd799439013
```

## Finding Your Credentials

### Dashboard URL

| Environment Type | Typical URL |
|-----------------|-------------|
| Local Development | `http://localhost:3000` |
| Tyk Cloud | `https://admin.cloud.tyk.io` |
| Self-hosted | Your custom domain |

### Auth Token & Organization ID

1. Log into your Tyk Dashboard
2. Navigate to **Users** → **Your User Profile**
3. Find the **API Access Credentials** section
4. Copy both the **Auth Token** and **Organization ID**

![Tyk Dashboard Credentials](../images/dashboard-credentials.png)

## Security Best Practices

### Protect Your Tokens

- **Never commit** auth tokens to version control
- Use **environment variables** in CI/CD pipelines
- **Rotate tokens** regularly
- Use **separate tokens** for different environments

### Example CI/CD Setup

```yaml
# GitHub Actions example
env:
  TYK_DASH_URL: ${{ secrets.TYK_DASH_URL }}
  TYK_AUTH_TOKEN: ${{ secrets.TYK_AUTH_TOKEN }}
  TYK_ORG_ID: ${{ secrets.TYK_ORG_ID }}

steps:
  - name: Deploy API
    run: tyk api apply --file api.yaml
```

### Example Docker Setup

```bash
# Pass environment variables to container
docker run -e TYK_DASH_URL=$TYK_DASH_URL \
           -e TYK_AUTH_TOKEN=$TYK_AUTH_TOKEN \
           -e TYK_ORG_ID=$TYK_ORG_ID \
           -v $(pwd):/workspace \
           tyk-cli tyk api apply --file /workspace/api.yaml
```

## Common Workflows

### Multi-Environment Deployment

```bash
# Deploy to staging first
tyk config use staging
tyk api apply --file api.yaml

# Then deploy to production
tyk config use production  
tyk api apply --file api.yaml
```

### Environment-Specific Configs

You can maintain separate API definitions per environment:

```
apis/
├── base-api.yaml           # Common configuration
├── dev-api.yaml           # Development overrides
├── staging-api.yaml       # Staging overrides
└── production-api.yaml    # Production overrides
```

```bash
# Deploy environment-specific configs
tyk config use dev && tyk api apply --file apis/dev-api.yaml
tyk config use staging && tyk api apply --file apis/staging-api.yaml
tyk config use production && tyk api apply --file apis/production-api.yaml
```

### Backup and Restore

```bash
# Backup current environment
tyk api get my-api --json > backup-api.json

# Restore to different environment
tyk config use staging
tyk api apply --file backup-api.json
```

## Troubleshooting

### Connection Issues

```bash
# Test current environment connection
tyk config current

# Test with specific credentials
tyk api get --dash-url https://test.tyk.io --auth-token <token> --org-id <org>
```

### Configuration Issues

```bash
# Check which config is being used
tyk config current

# Reset to default
rm ~/.config/tyk/cli.toml
tyk init
```

### Permission Issues

If you get "unauthorized" errors:
1. Verify your auth token is valid
2. Check your organization ID is correct
3. Ensure your user has API management permissions
4. Try regenerating your API token