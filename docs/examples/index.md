# Examples

Real-world examples and common workflows for using the Tyk CLI.

## Quick Examples

### Basic API Management

```bash
# Create an API from OpenAPI spec
tyk api create --file my-api.yaml

# Get API details
tyk api get my-api-123

# Update an existing API
tyk api apply --file updated-api.yaml

# Delete an API
tyk api delete my-api-123 --yes
```

### Environment Management

```bash
# Setup environments
tyk config add dev --dashboard-url http://localhost:3000 --auth-token dev-token --org-id dev-org
tyk config add prod --dashboard-url https://api.company.com --auth-token prod-token --org-id prod-org

# Switch between environments
tyk config use dev
tyk api apply --file api.yaml

tyk config use prod
tyk api apply --file api.yaml
```

## Sample Files

- **[Simple REST API](simple-rest-api.yaml)** - Basic REST API definition
- **[GraphQL API](graphql-api.yaml)** - GraphQL API with authentication
- **[Microservice API](microservice-api.yaml)** - Microservice with rate limiting
- **[Webhook API](webhook-api.yaml)** - Webhook receiver API

## Workflow Examples

- **[CI/CD Integration](cicd-integration.md)** - Automated API deployment
- **[Multi-Environment Setup](multi-environment.md)** - Managing dev/staging/prod
- **[API Versioning](api-versioning.md)** - Version management strategies
- **[GitOps Workflow](gitops-workflow.md)** - Git-based API management

## Advanced Examples

- **[Custom Middleware](custom-middleware.md)** - Adding custom functionality
- **[Authentication Setup](authentication.md)** - Various auth methods
- **[Rate Limiting](rate-limiting.md)** - Traffic control patterns
- **[Monitoring Setup](monitoring.md)** - Analytics and logging