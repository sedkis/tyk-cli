---
layout: default
title: Home
nav_order: 1
description: "A powerful command-line interface for managing Tyk APIs and configurations"
permalink: /
---

<div class="hero-section">
  <h1>Tyk CLI</h1>
  <p>A powerful command-line interface for managing Tyk APIs and configurations. Built to streamline API lifecycle management with OpenAPI Specification (OAS) support.</p>
  <div class="hero-buttons">
    <a href="{{ site.baseurl }}/getting-started" class="btn btn-primary">Get Started</a>
    <a href="https://github.com/tyktech/tyk-cli" class="btn btn-secondary">View on GitHub</a>
  </div>
</div>

<div class="feature-grid">
  <div class="feature-card">
    <span class="feature-icon">🚀</span>
    <h3>Interactive Setup</h3>
    <p>Get started quickly with guided configuration wizard that walks you through connecting to your Tyk Dashboard</p>
  </div>
  
  <div class="feature-card">
    <span class="feature-icon">🌍</span>
    <h3>Multi-Environment</h3>
    <p>Unified config system with named environments for seamless switching between dev, staging, and production</p>
  </div>
  
  <div class="feature-card">
    <span class="feature-icon">📝</span>
    <h3>OpenAPI First</h3>
    <p>Native support for OpenAPI 3.0 specifications with automatic Tyk gateway extension generation</p>
  </div>
  
  <div class="feature-card">
    <span class="feature-icon">🔧</span>
    <h3>Flexible Config</h3>
    <p>Configure via environment variables, config files, or CLI flags with intuitive precedence system</p>
  </div>
  
  <div class="feature-card">
    <span class="feature-icon">🎨</span>
    <h3>Beautiful CLI</h3>
    <p>Colorful, intuitive command-line experience with helpful prompts and clear error messages</p>
  </div>
  
  <div class="feature-card">
    <span class="feature-icon">✅</span>
    <h3>Well Tested</h3>
    <p>>80% test coverage with comprehensive live environment validation and integration tests</p>
  </div>
</div>

<div class="quick-start">
  <h3>Quick Start</h3>
  <p>Get up and running with Tyk CLI in under 2 minutes:</p>

```bash
# Install via Homebrew (recommended)
brew tap sedkis/tyk && brew install tyk

# Or download directly
curl -L "https://github.com/sedkis/tyk-cli/releases/latest/download/tyk-cli_$(uname -s)_$(uname -m).tar.gz" | tar xz
sudo mv tyk /usr/local/bin/

# Initialize with interactive wizard
tyk init

# Deploy your first API
tyk api create --file my-api.yaml
```
</div>

## 📚 Documentation

<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 1rem; margin: 2rem 0;">
  <a href="{{ site.baseurl }}/getting-started" style="display: block; padding: 1rem; border: 1px solid #e2e8f0; border-radius: 8px; text-decoration: none; color: inherit; transition: all 0.2s ease;">
    <strong>🚀 Getting Started</strong><br>
    <small>Installation, setup, and your first API deployment</small>
  </a>
  
  <a href="{{ site.baseurl }}/configuration" style="display: block; padding: 1rem; border: 1px solid #e2e8f0; border-radius: 8px; text-decoration: none; color: inherit; transition: all 0.2s ease;">
    <strong>⚙️ Configuration</strong><br>
    <small>Environment management and advanced configuration</small>
  </a>
  
  <a href="{{ site.baseurl }}/api-reference" style="display: block; padding: 1rem; border: 1px solid #e2e8f0; border-radius: 8px; text-decoration: none; color: inherit; transition: all 0.2s ease;">
    <strong>📖 API Reference</strong><br>
    <small>Complete command reference and examples</small>
  </a>
  
  <a href="{{ site.baseurl }}/examples/" style="display: block; padding: 1rem; border: 1px solid #e2e8f0; border-radius: 8px; text-decoration: none; color: inherit; transition: all 0.2s ease;">
    <strong>💡 Examples</strong><br>
    <small>Real-world usage patterns and workflows</small>
  </a>
</div>

## 🤝 Community & Support

- **🐛 Issues**: Report bugs and request features on [GitHub Issues](https://github.com/sedkis/tyk-cli/issues)
- **💬 Community**: Join discussions on the [Tyk Community Forum](https://community.tyk.io/)
- **📖 Contributing**: See our [Contributing Guide]({{ site.baseurl }}/CONTRIBUTING) to get involved
- **📧 Documentation**: Help improve these docs by clicking "Edit this page" on any page

---

<div style="text-align: center; margin: 3rem 0; color: #718096;">
  <p>Built with ❤️ by the Tyk community. Licensed under <a href="https://github.com/tyktech/tyk-cli/blob/main/LICENSE">MIT</a>.</p>
</div>