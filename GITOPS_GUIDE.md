# GitOps Guide for Tyk CLI

A comprehensive guide for managing Tyk APIs using GitOps principles with the Tyk CLI.

## Table of Contents

- [GitOps Principles](#gitops-principles)
- [CLI Command Semantics](#cli-command-semantics)
- [File Formats](#file-formats)
- [Basic Workflows](#basic-workflows)
- [Advanced GitOps Patterns](#advanced-gitops-patterns)
- [CI/CD Integration](#cicd-integration)
- [Troubleshooting](#troubleshooting)

## GitOps Principles

The Tyk CLI follows GitOps best practices:

1. **Declarative Configuration** - APIs are described in files, not CLI commands
2. **Version Controlled** - All API definitions are stored in Git
3. **Automated Deployment** - Changes are deployed via CI/CD pipelines  
4. **Observable** - Changes are auditable and rollback-friendly
5. **Convergence** - Actual state converges to desired state

## CLI Command Semantics

### `tyk api create` - Explicit Creation
Always creates a new API, regardless of file content.

```bash
# Works with plain OpenAPI specs
tyk api create -f my-api.yaml

# Works with Tyk-enhanced specs  
tyk api create -f tyk-api.yaml
```

**Behavior:**
- Plain OAS → Auto-generates `x-tyk-api-gateway` extensions
- Tyk OAS with ID → Strips existing ID, creates new API
- Always returns new API ID

### `tyk api apply` - Declarative GitOps
Declarative upsert based on file content and flags.

```bash
# Requires explicit intent for plain OAS
tyk api apply -f my-api.yaml --create

# Updates existing API (requires ID in file)
tyk api apply -f tracked-api.yaml

# Creates if missing, updates if exists
tyk api apply -f api.yaml --create
```

**Behavior:**
- Plain OAS without `--create` → Error with guidance
- Plain OAS with `--create` → Auto-generate extensions and create
- Tyk OAS with ID → Update existing API
- Tyk OAS without ID → Error (needs `--create` or use `create`)

### `tyk api update` - Explicit Update
Always updates an existing API.

```bash
# Update with explicit ID
tyk api update --api-id abc123 -f updated-api.yaml

# Update using ID from file
tyk api update -f api-with-id.yaml
```

### `tyk api get` - Retrieve API
Get API for inspection or GitOps tracking.

```bash
# Get API and save to file for tracking
tyk api get abc123 > apis/my-api.yaml

# Get specific version
tyk api get abc123 --version-name v2
```

## File Formats

### Plain OpenAPI Specification

Standard OpenAPI 3.0 specification without Tyk extensions:

```yaml
# apis/user-service.yaml
openapi: 3.0.0
info:
  title: User Management API
  version: 1.0.0
  description: Manages user accounts and profiles
servers:
  - url: https://api.example.com/users
    description: Production server
paths:
  /users:
    get:
      summary: List users
      responses:
        '200':
          description: Users retrieved successfully
```

**When used with `tyk api create` or `tyk api apply --create`, automatically generates:**

```yaml
x-tyk-api-gateway:
  info:
    name: User Management API
    state:
      active: true
  upstream:
    url: https://api.example.com/users
  server:
    listenPath:
      value: /user-management-api/
      strip: true
```

### Tyk-Enhanced OpenAPI Specification

OpenAPI with Tyk-specific extensions for GitOps tracking:

```yaml
# apis/user-service-tracked.yaml
openapi: 3.0.0
info:
  title: User Management API
  version: 1.0.0
servers:
  - url: https://api.example.com/users

# Tyk-specific configuration
x-tyk-api-gateway:
  info:
    id: "user-api-abc123"  # Required for GitOps tracking
    name: User Management API
    state:
      active: true
  upstream:
    url: https://api.example.com/users
  server:
    listenPath:
      value: /users/
      strip: true
  middleware:
    global:
      auth:
        authHeaderName: Authorization
      cors:
        enabled: true
        allowedOrigins:
          - "*"
      rateLimit:
        per: 100
        rate: 3600

paths:
  /users:
    get:
      summary: List users
      # Tyk-specific path overrides
      x-tyk-api-gateway:
        middleware:
          rateLimit:
            per: 10  # Override global rate limit for this path
            rate: 60
```

## Basic Workflows

### 1. Import Existing OpenAPI Spec

Starting with a standard OpenAPI specification:

```bash
# Step 1: Import plain OAS as new API
tyk api create -f openapi-spec.yaml

# Output: API created with ID abc123
# Step 2: Get the enhanced version for tracking
tyk api get abc123 > apis/my-api-tracked.yaml

# Step 3: Commit to Git
git add apis/my-api-tracked.yaml
git commit -m "feat: add My API to Tyk gateway"
```

### 2. Update API Configuration

Modify the tracked API file and apply changes:

```bash
# Edit apis/my-api-tracked.yaml
# - Change rate limit
# - Add new endpoints
# - Update upstream URL

# Apply changes
tyk api apply -f apis/my-api-tracked.yaml

# Commit changes
git add apis/my-api-tracked.yaml  
git commit -m "feat: increase rate limit and add /health endpoint"
```

### 3. Create New API from Scratch

```bash
# Step 1: Create Tyk-enhanced OAS file
cat > apis/new-service.yaml << 'EOF'
openapi: 3.0.0
info:
  title: New Service API
  version: 1.0.0
servers:
  - url: https://new-service.example.com

x-tyk-api-gateway:
  info:
    name: New Service API
    state:
      active: true
  upstream:
    url: https://new-service.example.com
  server:
    listenPath:
      value: /new-service/
      strip: true

paths:
  /status:
    get:
      summary: Service status
      responses:
        '200':
          description: Service is healthy
EOF

# Step 2: Create the API (will auto-generate ID)
tyk api apply -f apis/new-service.yaml --create

# Step 3: Update file with generated ID
tyk api get <returned-id> > apis/new-service.yaml

# Step 4: Commit
git add apis/new-service.yaml
git commit -m "feat: add New Service API"
```

## Advanced GitOps Patterns

### Environment-Based Directory Structure

Organize APIs by environment for multi-stage deployments:

```
apis/
├── dev/
│   ├── user-service.yaml
│   ├── order-service.yaml
│   └── payment-service.yaml
├── staging/
│   ├── user-service.yaml
│   ├── order-service.yaml
│   └── payment-service.yaml
└── prod/
    ├── user-service.yaml
    ├── order-service.yaml
    └── payment-service.yaml
```

**Deployment script:**
```bash
#!/bin/bash
# deploy-apis.sh

ENVIRONMENT=${1:-dev}
API_DIR="apis/${ENVIRONMENT}"

echo "Deploying APIs to ${ENVIRONMENT}..."

# Set environment-specific configuration
export TYK_DASH_URL="https://dashboard-${ENVIRONMENT}.example.com"
export TYK_AUTH_TOKEN="${TYK_AUTH_TOKEN_ENV}"
export TYK_ORG_ID="${TYK_ORG_ID_ENV}"

# Deploy all APIs in environment directory
for api_file in "${API_DIR}"/*.yaml; do
  echo "Applying $(basename "${api_file}")..."
  tyk api apply -f "${api_file}"
done

echo "Deployment to ${ENVIRONMENT} complete!"
```

### Service-Based Organization

Group APIs by service/team ownership:

```
services/
├── user-management/
│   ├── user-api.yaml
│   ├── profile-api.yaml
│   └── auth-api.yaml
├── e-commerce/
│   ├── catalog-api.yaml
│   ├── cart-api.yaml
│   └── order-api.yaml
└── shared/
    ├── logging-api.yaml
    └── metrics-api.yaml
```

### Configuration Templating

Use environment variables for environment-specific values:

```yaml
# apis/user-service.yaml
openapi: 3.0.0
info:
  title: User Service API
  version: 1.0.0
servers:
  - url: "${USER_SERVICE_URL}"  # Environment-specific

x-tyk-api-gateway:
  info:
    id: "${USER_API_ID}"  # Set per environment
    name: User Service API
    state:
      active: true
  upstream:
    url: "${USER_SERVICE_URL}"
  server:
    listenPath:
      value: /users/
      strip: true
  middleware:
    global:
      rateLimit:
        per: "${USER_API_RATE_LIMIT_PER}"
        rate: "${USER_API_RATE_LIMIT_RATE}"
```

**Environment files:**
```bash
# .env.dev
USER_SERVICE_URL=http://user-service.dev.svc.cluster.local:8080
USER_API_ID=user-api-dev-123
USER_API_RATE_LIMIT_PER=1000
USER_API_RATE_LIMIT_RATE=3600

# .env.prod  
USER_SERVICE_URL=https://user-service.prod.example.com
USER_API_ID=user-api-prod-456
USER_API_RATE_LIMIT_PER=100
USER_API_RATE_LIMIT_RATE=3600
```

## CI/CD Integration

### GitHub Actions Workflow

```yaml
# .github/workflows/deploy-apis.yml
name: Deploy Tyk APIs

on:
  push:
    branches: [main]
    paths: ['apis/**/*.yaml']
  
  pull_request:
    paths: ['apis/**/*.yaml']

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Install Tyk CLI
        run: |
          curl -L "https://github.com/sedkis/tyk-cli/releases/latest/download/tyk-cli_$(uname -s)_$(uname -m).tar.gz" | tar xz
          sudo mv tyk /usr/local/bin/
          
      - name: Validate API specifications
        run: |
          for api_file in apis/**/*.yaml; do
            echo "Validating $api_file"
            # Add validation logic here
            # e.g., OpenAPI spec validation, Tyk extension validation
          done

  deploy-staging:
    needs: validate
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    environment: staging
    steps:
      - uses: actions/checkout@v4
      
      - name: Install Tyk CLI
        run: |
          curl -L "https://github.com/sedkis/tyk-cli/releases/latest/download/tyk-cli_$(uname -s)_$(uname -m).tar.gz" | tar xz
          sudo mv tyk /usr/local/bin/
          
      - name: Deploy to Staging
        env:
          TYK_DASH_URL: ${{ secrets.TYK_STAGING_DASH_URL }}
          TYK_AUTH_TOKEN: ${{ secrets.TYK_STAGING_AUTH_TOKEN }}
          TYK_ORG_ID: ${{ secrets.TYK_STAGING_ORG_ID }}
        run: |
          for api_file in apis/staging/*.yaml; do
            echo "Deploying $(basename "$api_file") to staging..."
            tyk api apply -f "$api_file"
          done

  deploy-production:
    needs: deploy-staging
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/checkout@v4
      
      - name: Install Tyk CLI
        run: |
          curl -L "https://github.com/sedkis/tyk-cli/releases/latest/download/tyk-cli_$(uname -s)_$(uname -m).tar.gz" | tar xz
          sudo mv tyk /usr/local/bin/
          
      - name: Deploy to Production
        env:
          TYK_DASH_URL: ${{ secrets.TYK_PROD_DASH_URL }}
          TYK_AUTH_TOKEN: ${{ secrets.TYK_PROD_AUTH_TOKEN }}
          TYK_ORG_ID: ${{ secrets.TYK_PROD_ORG_ID }}
        run: |
          for api_file in apis/prod/*.yaml; do
            echo "Deploying $(basename "$api_file") to production..."
            tyk api apply -f "$api_file"
          done
```

### GitLab CI/CD Pipeline

```yaml
# .gitlab-ci.yml
stages:
  - validate
  - deploy-staging
  - deploy-production

variables:
  TYK_CLI_VERSION: "latest"

.install-tyk-cli: &install-tyk-cli
  - curl -L "https://github.com/sedkis/tyk-cli/releases/latest/download/tyk-cli_Linux_$(uname -m).tar.gz" | tar xz
  - mv tyk /usr/local/bin/

validate-apis:
  stage: validate
  image: ubuntu:latest
  before_script:
    - apt-get update && apt-get install -y curl
    - *install-tyk-cli
  script:
    - echo "Validating API specifications..."
    - find apis/ -name "*.yaml" -exec echo "Validating {}" \;
    # Add validation logic here
  rules:
    - changes:
      - apis/**/*.yaml

deploy-staging:
  stage: deploy-staging
  image: ubuntu:latest
  before_script:
    - apt-get update && apt-get install -y curl
    - *install-tyk-cli
  script:
    - echo "Deploying to staging environment..."
    - |
      for api_file in apis/staging/*.yaml; do
        echo "Deploying $(basename "$api_file")..."
        tyk api apply -f "$api_file"
      done
  environment:
    name: staging
    url: https://dashboard-staging.example.com
  rules:
    - if: $CI_COMMIT_BRANCH == "main"
      changes:
      - apis/**/*.yaml

deploy-production:
  stage: deploy-production
  image: ubuntu:latest
  before_script:
    - apt-get update && apt-get install -y curl
    - *install-tyk-cli
  script:
    - echo "Deploying to production environment..."
    - |
      for api_file in apis/prod/*.yaml; do
        echo "Deploying $(basename "$api_file")..."
        tyk api apply -f "$api_file"
      done
  environment:
    name: production
    url: https://dashboard.example.com
  when: manual
  rules:
    - if: $CI_COMMIT_BRANCH == "main"
      changes:
      - apis/**/*.yaml
```

### Jenkins Pipeline

```groovy
// Jenkinsfile
pipeline {
    agent any
    
    environment {
        TYK_CLI_VERSION = 'latest'
    }
    
    stages {
        stage('Setup') {
            steps {
                sh '''
                    curl -L "https://github.com/sedkis/tyk-cli/releases/latest/download/tyk-cli_Linux_$(uname -m).tar.gz" | tar xz
                    sudo mv tyk /usr/local/bin/
                    tyk --version
                '''
            }
        }
        
        stage('Validate APIs') {
            steps {
                sh '''
                    echo "Validating API specifications..."
                    find apis/ -name "*.yaml" -type f | while read api_file; do
                        echo "Validating $api_file"
                        # Add validation logic here
                    done
                '''
            }
        }
        
        stage('Deploy to Staging') {
            when {
                branch 'main'
            }
            environment {
                TYK_DASH_URL = credentials('tyk-staging-dash-url')
                TYK_AUTH_TOKEN = credentials('tyk-staging-auth-token')
                TYK_ORG_ID = credentials('tyk-staging-org-id')
            }
            steps {
                sh '''
                    echo "Deploying to staging..."
                    for api_file in apis/staging/*.yaml; do
                        echo "Deploying $(basename "$api_file")..."
                        tyk api apply -f "$api_file"
                    done
                '''
            }
        }
        
        stage('Deploy to Production') {
            when {
                branch 'main'
            }
            environment {
                TYK_DASH_URL = credentials('tyk-prod-dash-url')
                TYK_AUTH_TOKEN = credentials('tyk-prod-auth-token')
                TYK_ORG_ID = credentials('tyk-prod-org-id')
            }
            input {
                message "Deploy to production?"
                ok "Deploy"
            }
            steps {
                sh '''
                    echo "Deploying to production..."
                    for api_file in apis/prod/*.yaml; do
                        echo "Deploying $(basename "$api_file")..."
                        tyk api apply -f "$api_file"
                    done
                '''
            }
        }
    }
    
    post {
        success {
            echo 'Deployment completed successfully!'
        }
        failure {
            echo 'Deployment failed!'
        }
    }
}
```

## Troubleshooting

### Common Issues

#### "Plain OAS document detected" Error

```bash
Error: Plain OAS document detected (missing x-tyk-api-gateway extensions). 
Use 'tyk api create' for plain OAS files, or add --create flag to apply
```

**Solution:** Use `tyk api create` or add `--create` flag:
```bash
# Option 1: Use create command (always creates new API)
tyk api create -f my-api.yaml

# Option 2: Use apply with --create (creates if needed)
tyk api apply -f my-api.yaml --create
```

#### "API ID not found" Error

```bash
Error: API ID not found in x-tyk-api-gateway.info.id. 
Use 'tyk api create' or add --create flag to apply
```

**Solution:** The file has Tyk extensions but no API ID. Either:
1. Add the API ID to track existing API
2. Use `--create` to create new API
3. Use `tyk api create` to explicitly create new API

#### Rate Limit or Authentication Errors

```bash
Error: 401 Unauthorized
Error: 429 Too Many Requests
```

**Solution:** Check your configuration:
```bash
# Verify current configuration
tyk config current

# Test connectivity
tyk api get <any-existing-api-id>

# Update credentials if needed
tyk config set auth-token <new-token>
```

### Best Practices

1. **Always use version control** for API definitions
2. **Test in staging** before deploying to production  
3. **Use meaningful commit messages** for API changes
4. **Keep environment-specific values** in separate files or environment variables
5. **Validate API specifications** before deployment
6. **Use pull requests** for code review of API changes
7. **Monitor API performance** after deployments
8. **Keep API definitions close to service code** when possible

### Debugging Commands

```bash
# Get current environment configuration
tyk config current

# List all environments  
tyk config list

# Test API connectivity
tyk api get <api-id>

# View API details in JSON format
tyk api get <api-id> --json

# Check specific API version
tyk api get <api-id> --version-name v2
```

## Summary

The Tyk CLI provides a robust foundation for GitOps-based API management:

- **Flexible Import** - Works with plain OpenAPI specs or Tyk-enhanced files
- **Declarative Operations** - `apply` command for GitOps workflows  
- **Environment Management** - Multi-environment support with configuration switching
- **CI/CD Ready** - Easy integration with popular CI/CD platforms
- **Version Controlled** - Full audit trail and rollback capabilities

This approach enables teams to manage APIs as code, with all the benefits of modern DevOps practices applied to API gateway management.