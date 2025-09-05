---
layout: default
title: Multi-Environment Setup
parent: Examples
nav_order: 2
description: "Manage APIs across multiple environments with Tyk CLI"
---

# Multi-Environment Setup
{: .no_toc }

Guide for managing APIs across multiple environments (dev, staging, production) with the Tyk CLI.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

## Environment Setup

### Initial Configuration

```bash
# Interactive setup with multiple environments
tyk init

# Or manual setup
tyk config add dev --dashboard-url http://localhost:3000 --auth-token dev-token --org-id dev-org
tyk config add staging --dashboard-url https://staging.api.company.com --auth-token staging-token --org-id staging-org
tyk config add production --dashboard-url https://api.company.com --auth-token prod-token --org-id prod-org
```

### Environment Overview

```bash
# List all environments
tyk config list

# Output:
# Available environments:
# * dev        http://localhost:3000
#   staging    https://staging.api.company.com
#   production https://api.company.com
# 
# Current environment: dev
```

## Project Structure

Organize your API definitions for multi-environment deployment:

```
project/
├── apis/
│   ├── base/
│   │   └── user-api.yaml          # Base API definition
│   ├── environments/
│   │   ├── dev/
│   │   │   └── user-api.yaml      # Dev-specific overrides
│   │   ├── staging/
│   │   │   └── user-api.yaml      # Staging-specific overrides
│   │   └── production/
│   │       └── user-api.yaml      # Prod-specific overrides
│   └── shared/
│       └── common-config.yaml     # Shared configuration
├── scripts/
│   ├── deploy.sh                  # Deployment scripts
│   └── validate.sh                # Validation scripts
└── docs/
    └── api-docs.md
```

## Environment-Specific API Definitions

### Base API Definition

```yaml
# apis/base/user-api.yaml
openapi: 3.0.3
info:
  title: User Management API
  version: 1.0.0

x-tyk-api-gateway:
  info:
    name: user-api
    id: user-api-123
    state:
      active: true
  upstream:
    # This will be overridden per environment
    url: https://backend.example.com
  server:
    listenPath:
      value: /users/
      strip: true

paths:
  /users:
    get:
      summary: List users
      responses:
        '200':
          description: Success
```

### Development Environment

```yaml
# apis/environments/dev/user-api.yaml
openapi: 3.0.3
info:
  title: User Management API (Dev)
  version: 1.0.0-dev

x-tyk-api-gateway:
  info:
    name: user-api-dev
    id: user-api-dev-123
    state:
      active: true
  upstream:
    url: http://localhost:8080  # Local backend
  server:
    listenPath:
      value: /users/
      strip: true
  middleware:
    operations:
      # More permissive in dev
      /users:
        get:
          rateLimit:
            rate: 1000
            per: 60

paths:
  /users:
    get:
      summary: List users
      responses:
        '200':
          description: Success
```

### Staging Environment

```yaml
# apis/environments/staging/user-api.yaml
openapi: 3.0.3
info:
  title: User Management API (Staging)
  version: 1.0.0-rc

x-tyk-api-gateway:
  info:
    name: user-api-staging
    id: user-api-staging-123
    state:
      active: true
  upstream:
    url: https://staging-backend.company.com
  server:
    listenPath:
      value: /users/
      strip: true
  middleware:
    operations:
      /users:
        get:
          rateLimit:
            rate: 100
            per: 60
          authentication:
            enabled: true

paths:
  /users:
    get:
      summary: List users
      security:
        - ApiKeyAuth: []
      responses:
        '200':
          description: Success
```

### Production Environment

```yaml
# apis/environments/production/user-api.yaml
openapi: 3.0.3
info:
  title: User Management API
  version: 1.0.0

x-tyk-api-gateway:
  info:
    name: user-api-prod
    id: user-api-prod-123
    state:
      active: true
  upstream:
    url: https://backend.company.com
  server:
    listenPath:
      value: /users/
      strip: true
  middleware:
    operations:
      /users:
        get:
          rateLimit:
            rate: 50
            per: 60
          authentication:
            enabled: true
          # Additional security in production
          ipWhitelist:
            enabled: true
            allowedIPs: ["10.0.0.0/8", "192.168.0.0/16"]

paths:
  /users:
    get:
      summary: List users
      security:
        - ApiKeyAuth: []
      responses:
        '200':
          description: Success

components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: X-API-Key
```

## Deployment Workflows

### Manual Deployment

```bash
# Deploy to development
tyk config use dev
tyk api apply --file apis/environments/dev/user-api.yaml

# Deploy to staging  
tyk config use staging
tyk api apply --file apis/environments/staging/user-api.yaml

# Deploy to production
tyk config use production
tyk api apply --file apis/environments/production/user-api.yaml
```

### Scripted Deployment

```bash
#!/bin/bash
# scripts/deploy.sh

set -e

ENVIRONMENT=$1
API_FILE=$2

if [[ -z "$ENVIRONMENT" || -z "$API_FILE" ]]; then
    echo "Usage: $0 <environment> <api-file>"
    echo "Example: $0 staging apis/environments/staging/user-api.yaml"
    exit 1
fi

echo "🚀 Deploying $API_FILE to $ENVIRONMENT environment..."

# Switch to target environment
tyk config use "$ENVIRONMENT"

# Validate current environment
CURRENT_ENV=$(tyk config current | grep "Current environment:" | cut -d: -f2 | xargs)
if [[ "$CURRENT_ENV" != "$ENVIRONMENT" ]]; then
    echo "❌ Failed to switch to $ENVIRONMENT environment"
    exit 1
fi

echo "✅ Switched to $ENVIRONMENT environment"

# Deploy API
if tyk api apply --file "$API_FILE"; then
    echo "✅ Successfully deployed to $ENVIRONMENT"
else
    echo "❌ Deployment to $ENVIRONMENT failed"
    exit 1
fi
```

Usage:
```bash
# Deploy to specific environments
./scripts/deploy.sh dev apis/environments/dev/user-api.yaml
./scripts/deploy.sh staging apis/environments/staging/user-api.yaml
./scripts/deploy.sh production apis/environments/production/user-api.yaml
```

### Batch Deployment

```bash
#!/bin/bash
# scripts/deploy-all.sh

set -e

ENVIRONMENT=$1

if [[ -z "$ENVIRONMENT" ]]; then
    echo "Usage: $0 <environment>"
    echo "Available environments: dev, staging, production"
    exit 1
fi

ENV_DIR="apis/environments/$ENVIRONMENT"

if [[ ! -d "$ENV_DIR" ]]; then
    echo "❌ Environment directory $ENV_DIR not found"
    exit 1
fi

echo "🚀 Deploying all APIs to $ENVIRONMENT environment..."

# Switch environment
tyk config use "$ENVIRONMENT"

# Deploy all APIs in environment directory
for api_file in "$ENV_DIR"/*.yaml; do
    if [[ -f "$api_file" ]]; then
        echo "Deploying $(basename "$api_file")..."
        if tyk api apply --file "$api_file"; then
            echo "✅ $(basename "$api_file") deployed successfully"
        else
            echo "❌ Failed to deploy $(basename "$api_file")"
            exit 1
        fi
    fi
done

echo "🎉 All APIs deployed successfully to $ENVIRONMENT"
```

## Environment Promotion

### Promoting from Staging to Production

```bash
#!/bin/bash
# scripts/promote.sh

set -e

SOURCE_ENV="staging"
TARGET_ENV="production"
API_NAME="user-api"

echo "🔄 Promoting $API_NAME from $SOURCE_ENV to $TARGET_ENV..."

# Backup production API
tyk config use "$TARGET_ENV"
tyk api get "$API_NAME" --json > "backup-$API_NAME-$(date +%Y%m%d-%H%M%S).json"
echo "✅ Backed up production API"

# Get staging API
tyk config use "$SOURCE_ENV"
tyk api get "$API_NAME" --json > "temp-$API_NAME.json"
echo "✅ Retrieved staging API"

# Deploy to production
tyk config use "$TARGET_ENV"
if tyk api apply --file "temp-$API_NAME.json"; then
    echo "✅ Successfully promoted to production"
    rm "temp-$API_NAME.json"
else
    echo "❌ Promotion failed, backup available"
    exit 1
fi
```

## Configuration Management

### Environment Variables

Set up different configurations per environment:

```bash
# .env.dev
TYK_DASH_URL=http://localhost:3000
TYK_AUTH_TOKEN=dev-token-here
TYK_ORG_ID=dev-org-id
API_BACKEND_URL=http://localhost:8080
RATE_LIMIT=1000

# .env.staging  
TYK_DASH_URL=https://staging.api.company.com
TYK_AUTH_TOKEN=staging-token-here
TYK_ORG_ID=staging-org-id
API_BACKEND_URL=https://staging-backend.company.com
RATE_LIMIT=100

# .env.production
TYK_DASH_URL=https://api.company.com
TYK_AUTH_TOKEN=prod-token-here
TYK_ORG_ID=prod-org-id
API_BACKEND_URL=https://backend.company.com
RATE_LIMIT=50
```

### Template-Based Configuration

Use templates with environment-specific values:

```yaml
# apis/templates/user-api.template.yaml
openapi: 3.0.3
info:
  title: User Management API ({{.ENVIRONMENT}})
  version: {{.VERSION}}

x-tyk-api-gateway:
  info:
    name: user-api-{{.ENVIRONMENT}}
    id: user-api-{{.ENVIRONMENT}}-123
  upstream:
    url: {{.API_BACKEND_URL}}
  middleware:
    operations:
      /users:
        get:
          rateLimit:
            rate: {{.RATE_LIMIT}}
            per: 60
```

## Testing and Validation

### Environment Health Checks

```bash
#!/bin/bash
# scripts/health-check.sh

ENVIRONMENT=$1

if [[ -z "$ENVIRONMENT" ]]; then
    echo "Usage: $0 <environment>"
    exit 1
fi

echo "🔍 Running health check for $ENVIRONMENT environment..."

# Switch to environment
tyk config use "$ENVIRONMENT"

# Check connection
if ! tyk config current >/dev/null; then
    echo "❌ Cannot connect to $ENVIRONMENT"
    exit 1
fi

echo "✅ Connection to $ENVIRONMENT successful"

# List APIs to verify deployment
echo "📋 APIs in $ENVIRONMENT:"
tyk api list
```

### Smoke Tests

```bash
#!/bin/bash
# scripts/smoke-test.sh

ENVIRONMENT=$1

case $ENVIRONMENT in
    "dev")
        BASE_URL="http://localhost:3000"
        ;;
    "staging")
        BASE_URL="https://staging.api.company.com"
        ;;
    "production")
        BASE_URL="https://api.company.com"
        ;;
    *)
        echo "Unknown environment: $ENVIRONMENT"
        exit 1
        ;;
esac

echo "🧪 Running smoke tests for $ENVIRONMENT..."

# Test API endpoints
if curl -f "$BASE_URL/users/health" >/dev/null 2>&1; then
    echo "✅ Health endpoint working"
else
    echo "❌ Health endpoint failed"
    exit 1
fi

echo "🎉 Smoke tests passed for $ENVIRONMENT"
```

## Best Practices

### 1. Environment Naming
- Use consistent naming: `dev`, `staging`, `production`
- Include environment in API names and IDs
- Use environment-specific tags

### 2. Configuration Management
- Keep environment-specific configs separate
- Use templates for common configurations
- Store secrets securely (not in repos)

### 3. Deployment Safety
- Always backup before production deployments
- Use gradual rollouts for production
- Implement automatic rollback on failure

### 4. Monitoring
- Set up environment-specific monitoring
- Use different alert thresholds per environment
- Track deployment metrics

### 5. Access Control
- Use separate credentials per environment
- Implement least-privilege access
- Rotate credentials regularly