---
title: Configuration
nav_order: 5
---

# Configuration

Where settings come from (highest → lowest)
- Flags: command line overrides everything
- Environment variables: great for CI
- Config file: `~/.config/tyk/cli.toml`

Environments
- Named contexts (dev, staging, prod)
- Each has: dashboard URL, auth token, org ID

Add an environment
```
tyk config add staging --dashboard-url https://staging.example.com --auth-token $STAGING_TOKEN --org-id $STAGING_ORG
```

Make one default
```
tyk config use staging
```

Environment variables (override)
- `TYK_DASH_URL`
- `TYK_AUTH_TOKEN`
- `TYK_ORG_ID`

Examples
- Temporary override
  ```
  TYK_DASH_URL=http://localhost:3000 TYK_AUTH_TOKEN=dev-token TYK_ORG_ID=dev-org tyk api list
  ```
- Script-friendly switch
  ```
  export TYK_DASH_URL=$STAGING_URL
  export TYK_AUTH_TOKEN=$STAGING_TOKEN
  export TYK_ORG_ID=$STAGING_ORG
  tyk api list
  ```

Precedence
`flags > env vars > config file`

