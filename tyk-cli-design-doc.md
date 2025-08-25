# Tyk CLI Configuration Management Design Document

## Overview

This document outlines the design for a comprehensive Tyk CLI tool that enables developers to efficiently manage API configurations without the complexity of handling hundreds of individual command-line flags.

## Problem Statement

The [Tyk OAS API specification](https://tyk.io/docs/api-management/gateway-config-tyk-oas/) contains hundreds of configuration options across multiple sections (server, middleware, authentication, upstream, etc.). Creating individual CLI flags for each option would result in:

- **Unmanageable CLI interface** with hundreds of flags
- **Poor developer experience** with complex command discovery
- **Maintenance burden** requiring CLI updates for every new configuration option
- **Inconsistent patterns** across different configuration types

## Design Goals

1. **Scalability**: Handle hundreds of configuration options without CLI bloat
2. **Flexibility**: Support any Tyk OAS configuration without code changes
3. **Developer Experience**: Intuitive, discoverable, and safe operations
4. **Consistency**: Uniform patterns across all configuration types
5. **Integration**: Work seamlessly with CI/CD pipelines and automation

## Architecture

### Core Command Structure

```
tyk api config <command> <api-id> [options]
```

### Primary Commands

| Command | Purpose | Example |
|---------|---------|---------|
| `set` | Set configuration using JSON path notation | `tyk api config set <id> --set "server.gatewayTags.enabled=true"` |
| `apply` | Apply configuration from file | `tyk api config apply --file config.yaml --api-id <id>` |
| `get` | Retrieve current configuration | `tyk api config get <id> --format yaml` |
| `diff` | Show configuration differences | `tyk api config diff <id> --file config.yaml` |
| `validate` | Validate configuration | `tyk api config validate --file config.yaml` |

## Configuration Methods

### 1. JSON Path-Based Configuration

**Syntax**: `--set "path.to.config=value"`

```bash
# Single configuration
tyk api config set b84fe1a04e5648927971c0557971565c \
  --set "server.gatewayTags.enabled=true"

# Multiple configurations
tyk api config set b84fe1a04e5648927971c0557971565c \
  --set "server.gatewayTags.enabled=true" \
  --set "server.gatewayTags.tags=[dev2,production]" \
  --set "server.detailedActivityLogs.enabled=true"
```

**Supported Value Types**:
- Booleans: `true`, `false`
- Strings: `"value"` or `value`
- Numbers: `123`, `45.67`
- Arrays: `[item1,item2,item3]`
- Objects: `{key1:value1,key2:value2}`

### 2. File-Based Configuration

**YAML Configuration**:
```yaml
# config.yaml
server:
  gatewayTags:
    enabled: true
    tags: ["dev2", "production"]
  detailedActivityLogs:
    enabled: true
  detailedTracing:
    enabled: true
middleware:
  global:
    cache:
      enabled: true
      timeout: 300
upstream:
  url: "https://api.example.com"
```

**Usage**:
```bash
tyk api config apply --file config.yaml --api-id b84fe1a04e5648927971c0557971565c
```

### 3. Template-Based Configuration

**Template Definition**:
```yaml
# templates/production-api.yaml
metadata:
  name: "production-api"
  description: "Standard production API configuration"
  version: "1.0"
variables:
  - name: "environment"
    type: "string"
    required: true
  - name: "region"
    type: "string"
    default: "us-east"
config:
  server:
    gatewayTags:
      enabled: true
      tags: ["{{.environment}}", "{{.region}}", "production"]
    detailedActivityLogs:
      enabled: true
  middleware:
    global:
      cache:
        enabled: true
        timeout: 300
```

**Usage**:
```bash
tyk api config apply-template production-api \
  --api-id b84fe1a04e5648927971c0557971565c \
  --vars "environment=prod,region=eu-west"
```

### 4. Interactive Configuration

```bash
tyk api config edit b84fe1a04e5648927971c0557971565c --interactive
```

**Interactive Flow**:
```
? Select configuration section:
  > Server Configuration
    Middleware Configuration
    Authentication Configuration
    Upstream Configuration

? Server Configuration options:
  > Gateway Tags
    Listen Path
    Detailed Logging
    Circuit Breaker

? Gateway Tags configuration:
  Enable gateway tags? (Y/n): Y
  Enter tags (comma-separated): dev2, production
  ✓ Configuration updated
```

## Command Reference

### Core Commands

#### `tyk api config set`
Set individual configuration values using JSON path notation.

**Syntax**:
```bash
tyk api config set <api-id> --set "<path>=<value>" [options]
```

**Options**:
- `--set "<path>=<value>"` - Set configuration value (repeatable)
- `--dry-run` - Preview changes without applying
- `--backup-to <path>` - Backup current configuration before changes
- `--validate` - Validate configuration before applying
- `--output <format>` - Output format (json, yaml, table)

**Examples**:
```bash
# Enable gateway tags
tyk api config set b84fe1a04e5648927971c0557971565c \
  --set "server.gatewayTags.enabled=true" \
  --set "server.gatewayTags.tags=[dev2]"

# Complex middleware configuration
tyk api config set b84fe1a04e5648927971c0557971565c \
  --set "middleware.global.cache.enabled=true" \
  --set "middleware.global.cache.timeout=300" \
  --set "middleware.global.trafficLogs.enabled=true"
```

#### `tyk api config apply`
Apply configuration from YAML or JSON files.

**Syntax**:
```bash
tyk api config apply --file <config-file> --api-id <api-id> [options]
```

**Options**:
- `--file <path>` - Configuration file path (required)
- `--api-id <id>` - Target API ID (required)
- `--merge` - Merge with existing configuration (default: replace)
- `--backup-current` - Backup current configuration
- `--validate-before-apply` - Validate before applying changes

#### `tyk api config get`
Retrieve current API configuration.

**Syntax**:
```bash
tyk api config get <api-id> [options]
```

**Options**:
- `--format <format>` - Output format (json, yaml, table)
- `--section <section>` - Get specific configuration section
- `--output-file <path>` - Save to file instead of stdout

**Examples**:
```bash
# Get full configuration as YAML
tyk api config get b84fe1a04e5648927971c0557971565c --format yaml

# Get only server configuration
tyk api config get b84fe1a04e5648927971c0557971565c --section server

# Save configuration to file
tyk api config get b84fe1a04e5648927971c0557971565c \
  --format yaml --output-file current-config.yaml
```

#### `tyk api config diff`
Show differences between current configuration and proposed changes.

**Syntax**:
```bash
tyk api config diff <api-id> --file <config-file> [options]
```

**Options**:
- `--file <path>` - Configuration file to compare
- `--format <format>` - Diff format (unified, side-by-side, json)
- `--context <lines>` - Number of context lines in diff

#### `tyk api config validate`
Validate configuration against Tyk OAS schema.

**Syntax**:
```bash
tyk api config validate --file <config-file> [options]
```

**Options**:
- `--file <path>` - Configuration file to validate
- `--schema-version <version>` - Tyk OAS schema version
- `--strict` - Enable strict validation mode

### Utility Commands

#### `tyk api config list-templates`
List available configuration templates.

```bash
tyk api config list-templates [--category <category>]
```

#### `tyk api config create-template`
Create configuration template from existing API.

```bash
tyk api config create-template --from-api <api-id> \
  --name <template-name> \
  --description <description>
```

#### `tyk api config backup`
Backup API configuration.

```bash
tyk api config backup <api-id> --output-dir <backup-directory>
```

#### `tyk api config restore`
Restore API configuration from backup.

```bash
tyk api config restore <api-id> --from-backup <backup-file>
```

## Configuration Sections

Based on the [Tyk OAS specification](https://tyk.io/docs/api-management/gateway-config-tyk-oas/), the CLI supports all configuration sections:

### Server Configuration
- `server.gatewayTags` - Gateway tags for routing
- `server.listenPath` - API listen path configuration
- `server.detailedActivityLogs` - Detailed logging settings
- `server.detailedTracing` - Tracing configuration

### Middleware Configuration
- `middleware.global` - Global middleware settings
- `middleware.operations` - Per-operation middleware

### Authentication Configuration
- `authentication.enabled` - Authentication enablement
- `authentication.baseIdentityProvider` - Primary auth method
- `authentication.authConfigs` - Authentication configurations

### Upstream Configuration
- `upstream.url` - Upstream service URL
- `upstream.serviceDiscovery` - Service discovery settings

## Safety and Validation

### Pre-Apply Validation
1. **Schema Validation**: Validate against Tyk OAS schema
2. **Dependency Checking**: Verify configuration dependencies
3. **Value Range Validation**: Check numeric ranges and enum values
4. **Required Field Validation**: Ensure required fields are present

### Backup and Recovery
```bash
# Automatic backup before changes
tyk api config set <api-id> --set "config=value" --backup-to "./backups/"

# Manual backup
tyk api config backup <api-id> --output-dir "./backups/$(date +%Y%m%d-%H%M%S)"

# Restore from backup
tyk api config restore <api-id> --from-backup "./backups/20240115-143022/config.yaml"
```

### Dry Run Mode
```bash
# Preview changes without applying
tyk api config set <api-id> --set "server.gatewayTags.enabled=true" --dry-run

# Output shows:
# Configuration changes preview:
# + server.gatewayTags.enabled: true
# 
# No changes applied (dry-run mode)
```

## Integration Patterns

### CI/CD Pipeline Integration

**GitHub Actions Example**:
```yaml
name: Deploy API Configuration
on:
  push:
    branches: [main]
    paths: ['api-configs/**']

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Install Tyk CLI
        run: |
          curl -sSL https://install.tyk.io/cli | sh
          
      - name: Validate Configuration
        run: |
          tyk api config validate --file api-configs/production.yaml
          
      - name: Deploy Configuration
        run: |
          tyk api config apply \
            --file api-configs/production.yaml \
            --api-id ${{ secrets.API_ID }} \
            --backup-current \
            --validate-before-apply
        env:
          TYK_DASHBOARD_URL: ${{ secrets.TYK_DASHBOARD_URL }}
          TYK_DASHBOARD_SECRET: ${{ secrets.TYK_DASHBOARD_SECRET }}
```

### Environment-Specific Deployments

**Development**:
```bash
tyk api config apply-template base-api \
  --api-id b84fe1a04e5648927971c0557971565c \
  --vars "environment=development,debug=true" \
  --set "server.gatewayTags.tags=[dev2,testing]"
```

**Production**:
```bash
tyk api config apply-template base-api \
  --api-id b84fe1a04e5648927971c0557971565c \
  --vars "environment=production,debug=false" \
  --set "server.gatewayTags.tags=[production,eu-west]" \
  --require-approval
```

## Error Handling

### Validation Errors
```bash
$ tyk api config set <api-id> --set "server.gatewayTags.enabled=invalid"

Error: Invalid configuration value
  Path: server.gatewayTags.enabled
  Value: "invalid"
  Expected: boolean (true/false)
  
Suggestion: Use --set "server.gatewayTags.enabled=true"
```

### Network Errors
```bash
$ tyk api config apply --file config.yaml --api-id <api-id>

Error: Failed to connect to Tyk Dashboard
  URL: http://tyk-dashboard.localhost:3000
  Status: Connection refused
  
Suggestion: Check that Tyk Dashboard is running and accessible
```

### Configuration Conflicts
```bash
$ tyk api config set <api-id> --set "authentication.enabled=false" --set "authentication.authConfigs.oauth.enabled=true"

Warning: Configuration conflict detected
  authentication.enabled=false but authentication.authConfigs.oauth.enabled=true
  
Do you want to continue? (y/N): n
Operation cancelled
```

## Configuration Examples

### Basic Gateway Tags Setup
```bash
# Your specific use case
tyk api config set b84fe1a04e5648927971c0557971565c \
  --set "server.gatewayTags.enabled=true" \
  --set "server.gatewayTags.tags=[dev2]"
```

### Production API Configuration
```yaml
# production-config.yaml
server:
  gatewayTags:
    enabled: true
    tags: ["production", "eu-west", "critical"]
  detailedActivityLogs:
    enabled: true
  detailedTracing:
    enabled: true
  listenPath:
    strip: true
    value: "/api/v2/httpbin/"

middleware:
  global:
    cache:
      enabled: true
      timeout: 300
    trafficLogs:
      enabled: true
    contextVariables:
      enabled: true

authentication:
  enabled: true
  stripAuthorizationData: true
  baseIdentityProvider: "authToken"
  authConfigs:
    authToken:
      authHeaderName: "Authorization"

upstream:
  url: "https://production-httpbin.company.com"
  serviceDiscovery:
    enabled: false
```

```bash
tyk api config apply --file production-config.yaml \
  --api-id b84fe1a04e5648927971c0557971565c \
  --backup-current \
  --validate-before-apply
```

### Multi-Environment Template
```yaml
# templates/httpbin-api.yaml
metadata:
  name: "httpbin-api"
  description: "HTTPBin API configuration template"
  version: "1.0"

variables:
  - name: "environment"
    type: "string"
    required: true
    options: ["development", "staging", "production"]
  - name: "region"
    type: "string"
    default: "us-east"
  - name: "enable_caching"
    type: "boolean"
    default: false

config:
  server:
    gatewayTags:
      enabled: true
      tags: ["{{.environment}}", "{{.region}}", "httpbin"]
    detailedActivityLogs:
      enabled: "{{if eq .environment \"production\"}}true{{else}}false{{end}}"
    detailedTracing:
      enabled: true

  middleware:
    global:
      cache:
        enabled: "{{.enable_caching}}"
        timeout: "{{if .enable_caching}}300{{else}}0{{end}}"
      trafficLogs:
        enabled: true

  upstream:
    url: "{{if eq .environment \"production\"}}https://prod-httpbin.company.com{{else}}http://dev-httpbin.localhost{{end}}"
```

```bash
# Development deployment
tyk api config apply-template httpbin-api \
  --api-id b84fe1a04e5648927971c0557971565c \
  --vars "environment=development,region=us-east,enable_caching=false"

# Production deployment
tyk api config apply-template httpbin-api \
  --api-id b84fe1a04e5648927971c0557971565c \
  --vars "environment=production,region=eu-west,enable_caching=true"
```

## Implementation Considerations

### Technical Requirements
1. **JSON Path Library**: Use robust JSON path implementation (e.g., JSONPath, JMESPath)
2. **Schema Validation**: Integrate with Tyk OAS JSON Schema
3. **Template Engine**: Support Go templates or similar for variable substitution
4. **Configuration Merging**: Deep merge capabilities for partial updates
5. **Backup Management**: Versioned backup system with metadata

### Performance Considerations
1. **Lazy Loading**: Load configuration schemas on demand
2. **Caching**: Cache API configurations for repeated operations
3. **Parallel Operations**: Support bulk operations across multiple APIs
4. **Streaming**: Stream large configuration files instead of loading entirely in memory

### Security Considerations
1. **Credential Management**: Secure handling of API keys and secrets
2. **Audit Logging**: Log all configuration changes with user attribution
3. **Access Control**: Integration with Tyk RBAC system
4. **Encryption**: Encrypt sensitive configuration values in transit and at rest

## Future Enhancements

### Phase 2 Features
1. **Configuration Drift Detection**: Compare deployed vs. desired configuration
2. **Multi-API Operations**: Apply configurations across multiple APIs
3. **Configuration Versioning**: Track and manage configuration versions
4. **Rollback Capabilities**: Easy rollback to previous configurations
5. **Configuration Validation Rules**: Custom validation rules for organization policies

### Phase 3 Features
1. **GitOps Integration**: Native Git repository integration
2. **Policy as Code**: Define and enforce configuration policies
3. **Automated Migration**: Migrate configurations between environments
4. **Configuration Analytics**: Insights and recommendations for configurations
5. **Integration with Tyk Operator**: Seamless Kubernetes integration

## Conclusion

This design provides a scalable, flexible, and developer-friendly approach to managing Tyk API configurations. By avoiding the hundreds of individual CLI flags problem and instead using JSON path notation, file-based configurations, and templates, developers can efficiently manage complex API configurations while maintaining safety, validation, and integration capabilities.

The design supports all current Tyk OAS configuration options and provides a foundation for future enhancements without requiring CLI interface changes.

