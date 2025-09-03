# Tyk CLI Implementation Plan

Based on the design document, this TODO plan outlines the implementation phases for the Tyk CLI v0.

## 🎉 Recent Progress Summary

**✅ Unified Configuration/Environment System (Latest)**
- Implemented unified config system where environments ARE the configuration
- Interactive environment switching with arrow key navigation
- Enhanced config commands with beautiful color output and status indicators
- Added survey/v2 integration for user-friendly CLI interactions
- Fixed variable shadowing bug in environment selection

**✅ Phase 1 Foundation Complete**
- Unified configuration system with named environments in TOML
- Comprehensive HTTP client with live Tyk Dashboard integration
- Complete test suite with >80% coverage including live environment validation
- Modern CLI framework with Cobra and enhanced user experience

## Phase 1: Foundation & Setup ✅ COMPLETED

### [✅] 1. Project Structure & Dependencies
- [✅] Initialize Go module with proper versioning
- [✅] Add required dependencies (cobra, viper, yaml.v3, survey, fatih/color, etc.)
- [✅] Set up project directory structure:
  - `cmd/` - CLI commands
  - `internal/` - Internal packages
  - `pkg/` - Public packages (if any)
  - `test/` - Integration tests
- [✅] Configure build system and Makefile

### [✅] 2. Unified Configuration/Environment System
- [✅] **Unified Design**: Environments ARE configurations (no redundancy)
- [✅] Named environment support in `~/.config/tyk/cli.toml`:
  ```toml
  default_environment = "dev"
  
  [environments.dev]
  dashboard_url = "http://localhost:3000"
  auth_token = "dev-token"
  org_id = "dev-org-id"
  
  [environments.prod]
  dashboard_url = "https://api.tyk.io"
  auth_token = "prod-token"
  org_id = "prod-org-id"
  ```
- [✅] Environment variable overrides: `TYK_DASH_URL`, `TYK_AUTH_TOKEN`, `TYK_ORG_ID`
- [✅] Flag overrides: `--dash-url`, `--auth-token`, `--org-id`
- [✅] Config commands: `tyk config list/use/current/set`
- [✅] Interactive environment switching with beautiful UI

### [✅] 3. HTTP Client for Tyk Dashboard API
- [✅] Implement HTTP client with proper authentication
- [✅] Add timeout and retry logic
- [✅] Support for OAS API endpoints:
  - `POST /api/apis/oas` - Create OAS API
  - `PUT /api/apis/oas/{apiId}` - Update OAS API
  - `GET /api/apis/oas/{apiId}` - Get OAS API details
  - `DELETE /api/apis/oas/{apiId}` - Delete OAS API
  - `GET /api/apis/oas/{apiId}/versions` - List versions
  - `PATCH /api/apis/oas/{apiId}` - Switch default version

### [✅] 4. File Handling System
- [✅] Support for `.yaml`, `.yml`, `.json` file formats
- [✅] YAML/JSON parsing with proper error handling
- [✅] Line/column error reporting for parse failures
- [✅] Content type detection and conversion
- [✅] File existence and readability validation

### [✅] 5. CLI Framework Setup
- [✅] Initialize Cobra CLI with `tyk` as root command
- [✅] Set up command structure: `tyk api <subcommand>`
- [✅] Implement global flags: `--json`, `--help`
- [✅] Add version command and build info
- [✅] **Enhanced**: Interactive unified config/environment management with colors

## Phase 2: Core API Commands ✅ COMPLETED

### [✅] 6. `tyk api get` Command ✅ COMPLETED
- [✅] Implement basic get functionality by API ID
- [✅] Add `--version-name` flag for specific versions
- [✅] Support `--json` output format
- [✅] Human-readable output with summary + OAS to stdout
- [✅] Error handling for API not found (exit code 3)
- [✅] **Enhanced**: Proper OAS document parsing with Tyk extensions
- [✅] **Tested**: Works with live API ID `b84fe1a04e5648927971c0557971565c`

### [ ] 7. `tyk api create` Command (Refactor for Explicit Creation)
- [ ] **Refactor**: Simplify to explicit creation only (always generates new ID)
- [ ] Required flag: `--file <path>` with OAS document
- [ ] Ignore any existing `x-tyk-api-gateway.info.id` in file
- [ ] Optional flags:
  - [ ] `--version-name` (defaults to info.version or v1)
  - [ ] `--upstream-url`, `--listen-path`, `--custom-domain` (overrides)
  - [ ] `--set-default` (default: true for first version)
- [ ] Return new apiId and version name on success
- [ ] Handle creation conflicts (exit code 4)

### [ ] 8. `tyk api apply` Command (Declarative Upsert)
- [ ] Implement file-based declarative upsert
- [ ] Extract API ID from `x-tyk-api-gateway.info.id` extension
- [ ] Upsert logic: ID present → update, ID missing → error or create with `--create`
- [ ] Required flag: `--file`
- [ ] Optional flags: `--create`, `--version-name`, `--set-default`
- [ ] Error handling for missing ID without `--create` flag

### [ ] 9. `tyk api update` Command (Explicit Update)
- [ ] Implement explicit API update operation
- [ ] Required: `--api-id` (flag) OR ID in file via `x-tyk-api-gateway.info.id`
- [ ] Required: `--file` with OAS document
- [ ] Optional flags: `--version-name`, `--set-default`
- [ ] PUT operation to Dashboard API
- [ ] Handle API/version not found errors

### [ ] 10. `tyk api delete` Command
- [ ] Implement API deletion by ID
- [ ] Required flag: `--api-id`
- [ ] Confirmation prompt (unless `--yes` provided)
- [ ] Success message: "Deleted API <id>"
- [ ] Handle API not found errors


## Phase 3: Enhanced Command Features

### [ ] 11. `tyk api convert` Command
- [ ] Local-only OAS conversion (no network calls)
- [ ] Required flag: `--file`
- [ ] Optional flags: `--out` (output file), `--format`
- [ ] Support formats:
  - `apidef` (default) - classic Tyk API Definition JSON
  - `oas-with-tyk` - OAS with x-tyk-api-gateway extensions
- [ ] Output to stdout if no `--out` specified

### [ ] 12. JSON Output Support
- [ ] Implement `--json` flag for all applicable commands
- [ ] Consistent JSON structure across commands
- [ ] Machine-parseable error format:
  ```json
  {
    "error": {
      "status": 404,
      "code": "not_found",
      "message": "API abcd123 not found",
      "details": {}
    }
  }
  ```

### [ ] 13. Exit Code System
- [ ] Implement proper exit codes:
  - `0` - Success
  - `1` - Generic failure (I/O, network, unexpected)
  - `2` - Bad arguments (missing file, invalid flags)
  - `3` - Not found (API or version)
  - `4` - Conflict (duplicate resource)
- [ ] Ensure consistent exit codes across all commands

### [ ] 14. Error Handling & UX
- [ ] User-friendly error messages for common scenarios
- [ ] Network timeout handling with retry hints
- [ ] 4xx/5xx response handling with server details
- [ ] Client-side validation messages
- [ ] Never print auth tokens in logs/errors

## Phase 4: Testing & Documentation

### [✅] 15. Unit Tests
- [✅] Test unified config/environment system and precedence
- [✅] Test file parsing and validation
- [✅] Test command flag parsing and validation
- [✅] Test HTTP client error handling
- [✅] Test output formatting (human and JSON)
- [✅] Achieve >80% test coverage

### [✅] 16. Integration Tests
- [✅] Set up test environment with mock Tyk Dashboard
- [✅] Test complete command workflows
- [✅] Test error scenarios and edge cases
- [✅] Test file I/O operations
- [✅] Validate exit codes in various scenarios
- [✅] **Enhanced**: Live multi-environment testing against provided Tyk Dashboard

### [ ] 17. Documentation & Examples
- [ ] Update README.md with usage examples
- [ ] Create example OAS files for testing
- [ ] Document unified environment/config variable setup
- [ ] Add example CI/CD pipeline configurations
- [ ] Create troubleshooting guide

## Acceptance Criteria

- [ ] Core CRUD commands implemented: `get`, `create`, `update`, `delete`
- [ ] Declarative `apply` command with GitOps workflow support
- [ ] API ID-based upsert logic working correctly
- [ ] Convert command produces usable artifacts
- [ ] Environment variable overrides work in CI (no config file needed)
- [ ] `--json` outputs stable, scriptable fields
- [ ] All exit codes implemented correctly
- [ ] Comprehensive error handling with helpful messages

## Future Work (Post v0)

- [ ] Versioning commands (`tyk api versions list/create/switch-default`)
- [ ] OAS validation and linting (`tyk api validate`)
- [ ] Diff output for apply command (`tyk api apply --diff`)
- [ ] Support for policies, keys, and mocking
- [ ] Template system for common upstreams
- [ ] API selection by name (not just ID)
- [ ] Telemetry and usage analytics