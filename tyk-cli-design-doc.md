## Tyk CLI – OAS-first design document

### Audience
- **Developers**: create and manage OAS-native APIs quickly.
- **Platform engineers**: script simple CI/CD pipelines to provision, update, and cleanup APIs.

### Scope (v0)
- **Supported operations**: get, create, import, update, delete APIs (OAS-native).
- **OAS-native**: the source of truth is an OpenAPI document (JSON or YAML).
- **Versioning**: manage OAS API versions (list/create/switch default) in a minimal, opinionated way.
- **Convert**: import an OAS and convert it to a Tyk API definition artifact for migration or inspection.
- **Non-goals (v0)**: dry-run, schema validation, linting, templates, mocking, traffic, keys/policies, complex diff/apply flows.

### Design principles
- **Simple first**: each command does one thing; obvious flags; short help output.
- **CI-friendly**: deterministic exit codes, `--json` output where relevant, env-driven configuration.
- **Idempotent updates**: `update` fully replaces the OAS document of a given API/version.
- **No hidden magic**: `import` behavior is explicit (create vs update is controlled by flags).

### Terminology
- "API" refers to a Tyk OAS-native API managed via the Dashboard Admin API.
- "API definition" refers to the Tyk API Definition document (non-OAS) or the OAS with Tyk extensions.

## CLI overview

### Binary name
`tyk` (primary) with `tyk api ...` subcommands.

### Authentication & target configuration
- **Environment variables** (recommended for CI):
  - `TYK_DASH_URL`: dashboard base URL, e.g. `https://dashboard.example.com`
  - `TYK_AUTH_TOKEN`: dashboard API auth token
  - `TYK_ORG_ID`: organization ID (when required by the endpoint)
- **Flags** (override env): `--dash-url`, `--auth-token`, `--org-id`

### Output & exit codes
- Default: human-readable messages.
- `--json`: machine-friendly JSON for scripting.
- Exit codes:
  - `0` success
  - `1` generic failure (I/O, network, unexpected)
  - `2` bad arguments (missing file, invalid flag combination)
  - `3` not found (API or version)
  - `4` conflict (e.g. creating an API that already exists without `--force`)

## Commands (v0)

### 1) Get
Retrieve an API by ID.

```bash
tyk api get --api-id <apiId> [--json]
```

Behavior:
- Prints the current OAS document for the default version, unless `--version-name` is specified.
- Flags:
  - `--version-name <name>`: get a specific version

Output examples:
- Human: prints a short summary (name, listenPath, defaultVersion) and saves OAS to stdout.
- JSON: `{ "apiId": "...", "version": "...", "oas": { ... } }`

### 2) Create
Create a new API from a local OAS file.

```bash
tyk api create --file openapi.yaml \
  [--upstream-url https://service:8080] [--listen-path /my-api/] \
  [--custom-domain api.example.com] [--set-default]
```

Behavior:
- Creates a new base API with an initial version name derived from the OAS `info.version` unless overridden by `--version-name`.
- Flags:
  - `--file <path.{yaml|yml|json}>` (required)
  - `--version-name <name>`: initial version name; defaults to `info.version` or `v1`.
  - `--upstream-url`, `--listen-path`, `--custom-domain`: optional overrides applied during creation.
  - `--set-default`: mark the version as default (on by default if first version).

Result:
- Returns `apiId` (and version name) on success.

### 3) Import
Create or update from a local OAS without requiring Tyk extensions.

```bash
tyk api import --file openapi.yaml [--create] [--update --api-id <id>] \
  [--version-name v2] [--set-default]
```

Behavior:
- `--create`: create a new API (same as `create`, but accepts plain OAS without Tyk extensions).
- `--update --api-id`: replace the OAS document of the target API/version.
- If both `--update` and `--create` are omitted, command fails with exit code 2.
- Flags:
  - `--file <path>` (required)
  - `--api-id <id>` (required with `--update`)
  - `--version-name <name>`: target or new version; defaults to `info.version`.
  - `--set-default`: switch default version to `version-name` after operation.

Decision guide:
- Use `create` when you are provisioning a brand-new API.
- Use `import --update --api-id` to replace the OAS of an existing API/version.
- Use `import --create` when you only have a plain OAS and want the CLI to create a new API from it (no Tyk extensions needed).

Common mistakes and messages:
- Missing mode: running `tyk api import --file ...` without `--create` or `--update` exits with code 2 and message: "Specify one of --create or --update".
- Both modes: providing both `--create` and `--update` exits with code 2 and message: "Choose either --create or --update, not both".
- Missing `--api-id` in update mode exits with code 2 and message: "--api-id is required with --update".

### 4) Update
Replace the OAS of an existing API/version.

```bash
tyk api update --api-id <id> --file openapi.yaml [--version-name v2] [--set-default]
```

Behavior:
- Updates the specified version (or default if not provided) with the given OAS content.
- Full replace of the stored OAS.

### 5) Delete
Delete an API by ID.

```bash
tyk api delete --api-id <id> [--yes]
```

Behavior:
- Prompts for confirmation unless `--yes` is provided.

### 6) Versioning
Minimal helpers for API versions.

```bash
tyk api versions list --api-id <id> [--json]
tyk api versions create --api-id <id> --new-version-name v2 [--set-default]
tyk api versions switch-default --api-id <id> --version-name v2
```

Behavior:
- `list`: prints available versions and indicates the default.
- `create`: creates a new version linked to the base API. OAS content comes from `--file` if provided, otherwise duplicates from default.
  - Optional: `--file openapi.yaml`
- `switch-default`: marks a version as default.

### 7) Convert
Convert a local OAS to a Tyk API definition artifact (no network call).

```bash
tyk api convert --file openapi.yaml --out api-definition.json [--format apidef|oas-with-tyk]
```

Behavior:
- `--format apidef` (default): produce classic Tyk API Definition JSON.
- `--format oas-with-tyk`: emit OAS with `x-tyk-api-gateway` extensions populated.

## Example flows

### Developer: create and iterate locally
```bash
export TYK_DASH_URL=https://dashboard.example.com
export TYK_AUTH_TOKEN=***
export TYK_ORG_ID=***

tyk api create --file ./openapi.yaml --listen-path /orders/ --set-default --json > create.json
API_ID=$(jq -r .apiId create.json)

# Make changes to openapi.yaml...
tyk api update --api-id "$API_ID" --file ./openapi.yaml
tyk api get --api-id "$API_ID" --json | jq .oas.info.version
```

### Platform engineer: CI pipeline (import-and-update)
```bash
# assumes env vars are set in CI
tyk api import --update --api-id "$API_ID" --file ./openapi.yaml --version-name v2 --set-default --json
```

## Implementation notes

### HTTP API surface (Tyk Dashboard Admin API)
- Create OAS API: POST `/api/apis/oas` (body: OAS JSON/YAML)
- Update OAS API: PUT `/api/apis/oas/{apiId}` (body: OAS JSON/YAML)
- Get OAS API details: GET `/api/apis/oas/{apiId}`
- Delete OAS API: DELETE `/api/apis/oas/{apiId}`
- List OAS versions: GET `/api/apis/oas/{apiId}/versions`
- Create version: POST `/api/apis/oas` with `base_api_id` and `new_version_name` or a version endpoint if available
- Switch default version: PATCH `/api/apis/oas/{apiId}` with `set_default`

Notes:
- The CLI should accept both YAML and JSON and convert to the correct content type automatically.
- Where the API requires wrapper fields (e.g., `base_api_id`, `new_version_name`), the CLI will construct the payload from flags.

### File handling
- Accept `.yaml`, `.yml`, `.json`.
- Preserve order and comments when possible on read/write (YAML-in, YAML-out not guaranteed; we primarily send JSON over the wire and read YAML locally).

### Security
- Never print tokens in logs or errors.
- Support `~/.config/tyk/cli.toml` for defaults (optional in v0), overridden by env, then flags.

### Telemetry (optional, off by default)
- A `--anonymous-usage` flag may be added later; out of scope for v0.

## UX details

### Consistent flags
- `--file`, `--api-id`, `--version-name`, `--set-default`, `--json`, `--yes`.

### Minimal helpful output
- On create/import/update: print `apiId`, `versionName`, and listen path summary.
- On delete: print `Deleted API <id>`.

### Errors
- Network timeouts: show brief message with retry hint; non-zero exit. Example: "Request timed out contacting dashboard; check TYK_DASH_URL or try again".
- 4xx/5xx responses: display concise cause and any server-provided detail. On `--json`, return `{ "error": { "status": <int>, "code": "<string|optional>", "message": "...", "details": { ... } } }`.
- Not found (3): "API <id> not found".
- Conflict (4): when creating an API and server reports a duplicate or conflicting resource, e.g., listen path in use; message includes conflicting field if provided.
- Bad arguments (2): descriptive client-side messages (missing file, both modes set, etc.).

Examples (human output):
- Create conflict: "Cannot create API: listen path /orders/ is already in use (status 409)".
- Update not found: "API abcd123 not found (status 404)".

On `--json`, errors are machine-parseable and include `status`, `message`, and optional `details` map.

### Input validation (v0)
- Minimal client-side validation only:
  - Ensure `--file` exists and is readable.
  - Parse YAML/JSON; surface parser errors with line/column when available.
  - Check required flag combinations (e.g., `--api-id` with `--update`).
- No semantic OAS validation in v0 (no lint/dry-run); server-side validation errors are passed through with helpful context.

## Open questions (can be deferred)
- Should `import` auto-create when `--api-id` is not provided? For v0, keep explicit: require `--create` or `--update`.
- Should we support selecting an API by name? Defer to v1; v0 uses `--api-id`.
- How much of `x-tyk-*` should `convert` populate? v0: minimal sane defaults; advanced mapping later.

## Future work (post v0)
- Validation and linting (`tyk api validate`).
- Diff and apply strategies (`tyk api apply --diff`).
- Policies, keys, mocking, test data seeds.
- Templates/quickstarts for common upstreams.

## Acceptance checklist (v0)
- Get/create/import/update/delete commands implemented with described flags.
- Versioning list/create/switch-default operational.
- Convert produces usable `apidef` and `oas-with-tyk` artifacts.
- Works with env-only configuration in CI.
- `--json` outputs stable fields for scripting.


