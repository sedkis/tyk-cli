# Changelog

All notable changes to this project will be documented here.

## Unreleased

### Changed
- `tyk api apply` is now fully idempotent and acts as an upsert:
  - If `x-tyk-api-gateway.info.id` is present, `apply` updates the API if it exists; otherwise it creates a new API preserving the provided API ID.
  - If the ID is missing, `apply` automatically creates a new API.
- Behavior focuses on the API ID defined in the OAS (we do not care about DB IDs).

### Removed
- The `--create` flag has been removed from `tyk api apply`. Creation now happens automatically when the API is missing.

### Docs
- Updated README and examples to reflect idempotent apply behavior and removal of `--create`.

### Migration Notes
- Remove any usage of `tyk api apply --create` in your scripts/pipelines and use `tyk api apply --file <path>` instead.

