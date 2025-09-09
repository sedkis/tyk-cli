---
title: Getting Started
nav_order: 2
---

# Getting Started

Welcome! You’ll be productive in ~3 minutes.

## 1) Install

macOS, using brew
  ```
  brew tap sedkis/tyk && brew install tyk
  ```
Or Linux/macOS (tarball)
  ```
  curl -L "https://github.com/sedkis/tyk-cli/releases/latest/download/tyk-cli_$(uname -s)_$(uname -m).tar.gz" | tar xz
  sudo mv tyk /usr/local/bin/
  ```

2) Say hi
```
tyk --help
```

## 3) Create your first environment config
```
tyk init
## or
tyk config add dev --dashboard-url http://localhost:3000 --auth-token dev-token --org-id dev-org

## then
tyk config use dev
```

## 4) Create your first API
Create from scratch:
```bash
tyk api create --name httpbin --upstream-url http://httpbingo.org
```

Response:
```bash
✓ API created successfully!
  API ID:         8acf2c7c0d6d4bf3707b429afeaed791
  Name:           httpbin
  Version:        v1
  Listen Path:    /httpbin/
  Upstream URL:   http://httpbingo.org
  Default Version: v1

Next steps:
  tyk api get 8acf2c7c0d6d4bf3707b429afeaed791                           # View full configuration
  tyk api get 8acf2c7c0d6d4bf3707b429afeaed791 --oas-only > api.yaml    # Export for editing
```

## OAS Workflows
```
tyk api import-oas --file path/to/my-api.yaml
```

5) Check it worked
- See your API in the Tyk Dashboard
- Hit a simple endpoint or health route

Tips
- Keep tokens out of shell history by using env vars; see {{ site.baseurl }}/configuration
- Start with a small OAS to keep feedback tight
- Use --dry-run if you want a no-changes preview (when available)
