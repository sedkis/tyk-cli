---
title: Manage APIs
nav_order: 4
---

# Manage APIs

Here’s how to create, import, edit, and update APIs with the `tyk api` command — using small, copy‑pasteable steps.

Before you start
- Pick an environment: `tyk config use dev`
- Or set per‑run vars: `TYK_DASH_URL=... TYK_AUTH_TOKEN=... TYK_ORG_ID=...`

### Create your first API (from scratch)
```
tyk api create \
  --name "httpbin" \
  --upstream-url https://httpbin.org

# Find its ID
tyk api list
```
- `create` makes a minimal API you can call under `/petstore/` right away.
- Inspect it with `tyk api get <api-id>` (includes Tyk extensions).

### Import your first Open API Spec (OAS)
Create a tiny spec and import it.
```
cat > petstore.json <<'JSON'
{
  "openapi": "3.0.3",
  "info": { "title": "Petstore API", "version": "1.0.0" },
  "servers": [ { "url": "https://example.com" } ],
  "paths": {
    "/pets": { "get": { "summary": "List pets", "responses": { "200": { "description": "OK" } } } },
    "/pets/{id}": { "get": { "summary": "Get pet by ID", "parameters": [ { "name": "id", "in": "path", "required": true, "schema": { "type": "string" } } ], "responses": { "200": { "description": "OK" } } } }
  }
}
JSON

tyk api import-oas --file petstore.json
```
- `import-oas` creates a new API from a clean OAS (no Tyk config required).

## Edit Existing Tyk OAS API

It's simple to edit an existing Tyk API:

```
# 1) Export the full definition (with Tyk extensions)
tyk api get <api-id> > foo.yaml

# 2) Edit foo.yaml (change listenPath, upstream, middleware, etc.)

# 3) Apply your changes
tyk api apply --file foo.yaml
```

## Patch OAS Docs on an API

imagine you have made a change to the OAS definition.  Perhaps it is auto-generated from a code change, or it was designed in an API designer. 

We can natively patch the existing Tyk reverse proxy definition for this API.

```
# 1) Export an OAS spec
tyk api get <api-id> --oas-only > openapi.yaml

# 2) Edit with your favorite OAS tools

# 3) Update only the spec on the existing API
tyk api update-oas <api-id> --file openapi.yaml
```


## Handy commands
- Inspect: `tyk api get <api-id> [--oas-only]`
- Browse: `tyk api list` (add `--page 2` or `--interactive`)
- Delete: `tyk api delete <api-id> --yes`
- Global flags: `--dash-url`, `--auth-token`, `--org-id`, `--json`
