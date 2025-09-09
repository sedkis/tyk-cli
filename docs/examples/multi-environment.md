---
title: Multi-Environment
parent: Git Ops
nav_order: 2
---

# Multi-Environment

Define environments once
```
tyk config add dev --dashboard-url http://localhost:3000 --auth-token dev-token --org-id dev-org
tyk config add staging --dashboard-url https://staging.api.company.com --auth-token $STAGING_TOKEN --org-id $STAGING_ORG
tyk config add prod --dashboard-url https://api.company.com --auth-token $PROD_TOKEN --org-id $PROD_ORG
```

Deploy the same API everywhere
```
tyk config use dev && tyk api apply --file apis/users.yaml
tyk config use staging && tyk api apply --file apis/users.yaml
tyk config use prod && tyk api apply --file apis/users.yaml
```

Promote with a quick check
- Grab the spec to compare
  ```
  tyk api get <api-id> --oas-only > users.dev.yaml
  ```

Tips
- Include the environment in versioned API IDs if you want isolation
- Or use the same ID everywhere for a “one artifact, many envs” flow
