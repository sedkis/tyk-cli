# Tyk API Command Architecture Design

## Overview

The Tyk CLI provides three complementary commands for API management:
- **`tyk api create`** - Explicit creation of new APIs
- **`tyk api update`** - Explicit updates to existing APIs  
- **`tyk api apply`** - Declarative upsert based on API ID presence

This design provides both explicit CRUD operations and GitOps-style declarative management.

## Command Architecture

### `tyk api apply` - Declarative Upsert (Primary)
**Usage**: `tyk api apply --file api.yaml [options]`

**Behavior**:
- **File-driven**: Only accepts `--file` input (no CLI flags for API creation)
- **ID Detection**: Checks for `x-tyk-api-gateway.info.id` in OAS document
- **Upsert Logic**:
  - **ID Present**: Updates existing API (calls update internally)
  - **ID Missing**: Rejects with error message directing to `--create` flag or `tyk api create`
- **Safety**: Prevents accidental bulk API creation in CI/CD pipelines

### `tyk api create` - Explicit Creation
**Usage**: `tyk api create --file api.yaml [options]`

**Behavior**:
- Always creates new API (generates new ID)
- Accepts OAS files without existing `x-tyk-api-gateway.info.id`
- Returns new API ID for subsequent operations
- Can accept override flags for upstream URL, listen path, etc.

### `tyk api update` - Explicit Update
**Usage**: `tyk api update --api-id <id> --file api.yaml [options]`

**Behavior**:
- Always updates existing API
- Requires explicit `--api-id` flag or ID in file
- Replaces OAS document for specified API/version
- Returns updated API details

## Flag Design

### `tyk api apply` Flags

**Required:**
- `--file`, `-f`: Path to OAS file with API definition

**Optional:**
- `--create`: Allow creation of new APIs when ID is missing
- `--version-name`: Version name (default: extract from file or "v1")
- `--set-default`: Set as default version (default: true)

### `tyk api create` Flags

**Required:**
- `--file`, `-f`: Path to OAS file

**Optional:**
- `--version-name`: Version name
- `--set-default`: Set as default version
- `--upstream-url`: Override upstream URL from file
- `--listen-path`: Override listen path from file
- `--name`: Override API name from file
- `--custom-domain`: Custom domain for API

### `tyk api update` Flags

**Required:**
- `--api-id`: Target API ID (alternative: ID in file)
- `--file`, `-f`: Path to OAS file

**Optional:**
- `--version-name`: Target version name
- `--set-default`: Set as default version

## Input Validation

### API ID Detection Logic
```go
func extractAPIID(oasDoc map[string]interface{}) (string, bool) {
    // Navigate to x-tyk-api-gateway.info.id
    if xTyk, exists := oasDoc["x-tyk-api-gateway"]; exists {
        if xTykMap, ok := xTyk.(map[string]interface{}); ok {
            if info, exists := xTykMap["info"]; exists {
                if infoMap, ok := info.(map[string]interface{}); ok {
                    if id, exists := infoMap["id"]; exists {
                        if idStr, ok := id.(string); ok && idStr != "" {
                            return idStr, true
                        }
                    }
                }
            }
        }
    }
    return "", false
}

func validateOASFile(filePath string) (map[string]interface{}, error) {
    content := loadFile(filePath)
    
    var doc map[string]interface{}
    if err := parseDocument(content, &doc); err != nil {
        return nil, err
    }
    
    // Validate OpenAPI format
    if openapi, exists := doc["openapi"]; !exists {
        return nil, fmt.Errorf("not a valid OpenAPI specification")
    } else if !strings.HasPrefix(openapi.(string), "3.") {
        return nil, fmt.Errorf("unsupported OpenAPI version, requires 3.x")
    }
    
    return doc, nil
}
```

### Validation Rules

**Apply Command Validation:**
- File must be valid OpenAPI 3.x specification
- If no `--create` flag and no API ID in file, reject with helpful error
- If API ID present, verify API exists on dashboard
- Version name must be valid identifier if specified

**Create Command Validation:**
- File must be valid OpenAPI 3.x specification
- Ignore any existing API ID in file (generates new one)
- Required OAS fields must be present (info.title, paths, etc.)
- Override flags must be valid (URLs, paths, etc.)

**Update Command Validation:**
- Must provide either `--api-id` flag or ID in file
- File must be valid OpenAPI 3.x specification
- Target API must exist on dashboard
- Version must exist if specified

## Apply Command Flow

```
┌─────────────────┐
│ Parse Flags     │
│ & Load File     │
└─────────┬───────┘
          │
          ▼
┌─────────────────┐
│ Validate OAS    │
│ Document        │
└─────────┬───────┘
          │
          ▼
┌─────────────────┐
│ Extract API ID  │
│ from x-tyk ext  │
└─────────┬───────┘
          │
          ▼
    ┌─────────────┐     ┌─────────────────┐
    │ ID Present? │────▶│ Update Existing │
    └─────┬───────┘  Y  │ API (via update)│
          │N            └─────────────────┘
          ▼
    ┌─────────────┐     ┌─────────────────┐
    │--create     │────▶│ Create New API  │
    │flag given?  │  Y  │ (via create)    │
    └─────┬───────┘     └─────────────────┘
          │N
          ▼
    ┌─────────────────┐
    │ Error: Missing  │
    │ ID. Use --create│
    │ or tyk api create│
    └─────────────────┘
```

## Create/Update Command Flow

```
┌─────────────────┐
│ Parse Flags     │
│ & Load File     │
└─────────┬───────┘
          │
          ▼
┌─────────────────┐
│ Validate OAS &  │
│ Required Flags  │
└─────────┬───────┘
          │
          ▼
┌─────────────────┐
│ Apply Flag      │
│ Overrides       │
└─────────┬───────┘
          │
          ▼
┌─────────────────┐
│ Call Dashboard  │
│ API (POST/PUT)  │
└─────────┬───────┘
          │
          ▼
┌─────────────────┐
│ Output Result   │
│ (Human/JSON)    │
└─────────────────┘
```

## Implementation Plan

### Phase 1: Refactor Existing Create Command
1. Simplify create command to explicit creation only
2. Remove complex mode detection (file-only input)
3. Ensure create always generates new API ID

### Phase 2: Implement Update Command
1. Add update command with --api-id requirement
2. Implement PUT operation to Dashboard API
3. Add version-specific update support

### Phase 3: Implement Apply Command
1. Add API ID extraction from x-tyk-api-gateway extension
2. Implement upsert logic (ID present → update, missing → error)
3. Add --create flag for new API creation in apply

### Phase 4: Enhanced Features
1. Add comprehensive validation for all three commands
2. Implement consistent error handling and messaging
3. Add flag override functionality where appropriate

## Error Handling

### Apply Command Errors
- **Missing ID without --create**: "API ID not found in file. Use 'tyk api create' to create a new API, or add --create flag to create via apply"
- **API not found**: "API with ID 'xyz' not found. Verify the API exists or use 'tyk api create'"
- **Invalid OAS**: "Invalid OpenAPI specification in 'path': missing required field 'info.title'"

### Create Command Errors
- **File not found**: "File 'path' does not exist"
- **Parse error**: "Invalid JSON/YAML in 'path': line X, column Y"
- **Invalid OAS**: "Not a valid OpenAPI 3.x specification"
- **Creation conflict**: "API creation failed: name already exists or other conflict"

### Update Command Errors
- **Missing API ID**: "Missing required API ID. Use --api-id flag or ensure x-tyk-api-gateway.info.id is set in file"
- **API not found**: "API with ID 'xyz' not found"
- **Version not found**: "Version 'v2' not found for API 'xyz'"

## Examples

### Apply Command (GitOps Workflow)
```bash
# Upsert existing API (has x-tyk-api-gateway.info.id)
tyk api apply --file api-with-id.yaml

# Create new API with apply
tyk api apply --file new-api.yaml --create

# Error case - missing ID without --create
tyk api apply --file new-api.yaml
# Error: API ID not found in file. Use 'tyk api create' or add --create flag
```

### Create Command (Explicit Creation)
```bash
# Create new API from OAS file
tyk api create --file petstore.yaml

# Create with overrides
tyk api create --file petstore.yaml --upstream-url https://api.example.com --version-name v2
```

### Update Command (Explicit Update)
```bash
# Update by API ID flag
tyk api update --api-id abc123 --file updated-api.yaml

# Update using ID from file
tyk api update --file api-with-id.yaml --version-name v3
```

## GitOps Integration Benefits

**Declarative Workflow:**
```bash
# CI/CD pipeline - safe upserts
tyk api apply --file ./apis/user-service.yaml
tyk api apply --file ./apis/order-service.yaml
```

**Explicit Operations:**
```bash
# Manual API creation
tyk api create --file new-service.yaml

# Targeted updates  
tyk api update --api-id abc123 --file updated-service.yaml
```

## Backward Compatibility

- Existing `tyk api create --file` usage continues to work
- Current flag behavior preserved 
- Apply command is additive (no breaking changes)
- Remove placeholder `import` command

This architecture provides both explicit CRUD semantics and GitOps-style declarative management, enabling teams to choose the approach that fits their workflow while maintaining safety and clarity in API operations.