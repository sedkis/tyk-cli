---
title: Github Actions
parent: Git Ops
nav_order: 1
---

# CI/CD Integration

GitHub Actions (two environments)
```
name: Deploy APIs
on: [push]
jobs:
  deploy:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        env: [staging, prod]
    steps:
      - uses: actions/checkout@v4
      - name: Install Tyk CLI
        run: |
          curl -L "https://github.com/sedkis/tyk-cli/releases/latest/download/tyk-cli_Linux_x86_64.tar.gz" | tar xz
          sudo mv tyk /usr/local/bin/tyk
      - name: Deploy API
        env:
          TYK_DASH_URL: ${{ secrets[matrix.env == 'staging' && 'STAGING_DASH_URL' || 'PROD_DASH_URL'] }}
          TYK_AUTH_TOKEN: ${{ secrets[matrix.env == 'staging' && 'STAGING_AUTH_TOKEN' || 'PROD_AUTH_TOKEN'] }}
          TYK_ORG_ID: ${{ secrets[matrix.env == 'staging' && 'STAGING_ORG_ID' || 'PROD_ORG_ID'] }}
        run: |
          # First deployment (no existing API): add --create
          # tyk api apply --file apis/users.yaml --create
          # Subsequent updates:
          tyk api apply --file apis/users.yaml
```

Smoke test (bash)
```
set -euo pipefail
curl -fsS "$BASE_URL/users/health" >/dev/null
echo "✅ API is healthy"
```

Pro tips
- Keep tokens in repo secrets, not in YAML
- Fail fast with small health checks
- Use a matrix to deploy the same spec to multiple envs
