# CI/CD Integration

Examples of integrating Tyk CLI into your CI/CD pipelines for automated API deployment.

## GitHub Actions

### Basic Workflow

```yaml
# .github/workflows/api-deploy.yml
name: Deploy API

on:
  push:
    branches: [main]
    paths: ['api/**']

jobs:
  deploy:
    runs-on: ubuntu-latest
    
    steps:
    - name: Checkout
      uses: actions/checkout@v3
      
    - name: Install Tyk CLI
      run: |
        curl -L "https://github.com/sedkis/tyk-cli/releases/latest/download/tyk-cli_Linux_x86_64.tar.gz" | tar xz
        sudo mv tyk /usr/local/bin/
        
    - name: Deploy to Staging
      env:
        TYK_DASH_URL: ${{ secrets.STAGING_DASH_URL }}
        TYK_AUTH_TOKEN: ${{ secrets.STAGING_AUTH_TOKEN }}
        TYK_ORG_ID: ${{ secrets.STAGING_ORG_ID }}
      run: |
        tyk api apply --file api/api.yaml
        
    - name: Deploy to Production
      if: github.ref == 'refs/heads/main'
      env:
        TYK_DASH_URL: ${{ secrets.PROD_DASH_URL }}
        TYK_AUTH_TOKEN: ${{ secrets.PROD_AUTH_TOKEN }}
        TYK_ORG_ID: ${{ secrets.PROD_ORG_ID }}
      run: |
        tyk api apply --file api/api.yaml
```

### Multi-Environment Workflow

```yaml
# .github/workflows/multi-env-deploy.yml
name: Multi-Environment Deploy

on:
  push:
    branches: [develop, staging, main]
    paths: ['apis/**']

jobs:
  deploy:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        include:
          - branch: develop
            environment: dev
            dash_url_secret: DEV_DASH_URL
            auth_token_secret: DEV_AUTH_TOKEN
            org_id_secret: DEV_ORG_ID
          - branch: staging
            environment: staging
            dash_url_secret: STAGING_DASH_URL
            auth_token_secret: STAGING_AUTH_TOKEN
            org_id_secret: STAGING_ORG_ID
          - branch: main
            environment: production
            dash_url_secret: PROD_DASH_URL
            auth_token_secret: PROD_AUTH_TOKEN
            org_id_secret: PROD_ORG_ID
    
    steps:
    - name: Checkout
      uses: actions/checkout@v3
      
    - name: Install Tyk CLI
      run: |
        curl -L "https://github.com/sedkis/tyk-cli/releases/latest/download/tyk-cli_Linux_x86_64.tar.gz" | tar xz
        sudo mv tyk /usr/local/bin/
        
    - name: Deploy APIs
      if: github.ref == format('refs/heads/{0}', matrix.branch)
      env:
        TYK_DASH_URL: ${{ secrets[matrix.dash_url_secret] }}
        TYK_AUTH_TOKEN: ${{ secrets[matrix.auth_token_secret] }}
        TYK_ORG_ID: ${{ secrets[matrix.org_id_secret] }}
      run: |
        for api_file in apis/*.yaml; do
          echo "Deploying $api_file to ${{ matrix.environment }}"
          tyk api apply --file "$api_file"
        done
```

## GitLab CI

### Basic Pipeline

```yaml
# .gitlab-ci.yml
stages:
  - validate
  - deploy-staging
  - deploy-production

variables:
  CLI_VERSION: "latest"

before_script:
  - curl -L "https://github.com/sedkis/tyk-cli/releases/${CLI_VERSION}/download/tyk-cli_Linux_x86_64.tar.gz" | tar xz
  - mv tyk /usr/local/bin/

validate-apis:
  stage: validate
  script:
    - |
      for api_file in apis/*.yaml; do
        echo "Validating $api_file"
        # Add validation logic here
        yaml-lint "$api_file"
      done
  only:
    changes:
      - apis/**

deploy-staging:
  stage: deploy-staging
  script:
    - |
      export TYK_DASH_URL=$STAGING_DASH_URL
      export TYK_AUTH_TOKEN=$STAGING_AUTH_TOKEN
      export TYK_ORG_ID=$STAGING_ORG_ID
      
      for api_file in apis/*.yaml; do
        echo "Deploying $api_file to staging"
        tyk api apply --file "$api_file"
      done
  only:
    - develop
    - staging
  environment:
    name: staging

deploy-production:
  stage: deploy-production
  script:
    - |
      export TYK_DASH_URL=$PROD_DASH_URL
      export TYK_AUTH_TOKEN=$PROD_AUTH_TOKEN
      export TYK_ORG_ID=$PROD_ORG_ID
      
      for api_file in apis/*.yaml; do
        echo "Deploying $api_file to production"
        tyk api apply --file "$api_file"
      done
  only:
    - main
  environment:
    name: production
  when: manual  # Require manual approval
```

## Jenkins Pipeline

### Declarative Pipeline

```groovy
// Jenkinsfile
pipeline {
    agent any
    
    environment {
        CLI_VERSION = 'latest'
    }
    
    stages {
        stage('Install CLI') {
            steps {
                sh '''
                    curl -L "https://github.com/sedkis/tyk-cli/releases/${CLI_VERSION}/download/tyk-cli_Linux_x86_64.tar.gz" | tar xz
                    sudo mv tyk /usr/local/bin/
                '''
            }
        }
        
        stage('Validate APIs') {
            when {
                changeset "apis/**"
            }
            steps {
                sh '''
                    for api_file in apis/*.yaml; do
                        echo "Validating $api_file"
                        yamllint "$api_file"
                    done
                '''
            }
        }
        
        stage('Deploy to Staging') {
            when {
                anyOf {
                    branch 'develop'
                    branch 'staging'
                }
            }
            environment {
                TYK_DASH_URL = credentials('staging-dash-url')
                TYK_AUTH_TOKEN = credentials('staging-auth-token')
                TYK_ORG_ID = credentials('staging-org-id')
            }
            steps {
                sh '''
                    for api_file in apis/*.yaml; do
                        echo "Deploying $api_file to staging"
                        tyk api apply --file "$api_file"
                    done
                '''
            }
        }
        
        stage('Deploy to Production') {
            when {
                branch 'main'
            }
            environment {
                TYK_DASH_URL = credentials('prod-dash-url')
                TYK_AUTH_TOKEN = credentials('prod-auth-token')
                TYK_ORG_ID = credentials('prod-org-id')
            }
            steps {
                input message: 'Deploy to production?', ok: 'Deploy'
                sh '''
                    for api_file in apis/*.yaml; do
                        echo "Deploying $api_file to production"
                        tyk api apply --file "$api_file"
                    done
                '''
            }
        }
    }
    
    post {
        success {
            echo 'Deployment completed successfully!'
        }
        failure {
            echo 'Deployment failed!'
        }
    }
}
```

## Docker Integration

### Dockerfile for CI

```dockerfile
FROM alpine:latest

RUN apk add --no-cache curl bash

# Install Tyk CLI
RUN curl -L "https://github.com/sedkis/tyk-cli/releases/latest/download/tyk-cli_Linux_x86_64.tar.gz" | tar xz && \
    mv tyk /usr/local/bin/

WORKDIR /workspace
COPY apis/ ./apis/

CMD ["tyk", "api", "apply", "--file", "apis/"]
```

### Docker Compose for Multi-Environment

```yaml
# docker-compose.yml
version: '3.8'

services:
  deploy-staging:
    build: .
    environment:
      - TYK_DASH_URL=${STAGING_DASH_URL}
      - TYK_AUTH_TOKEN=${STAGING_AUTH_TOKEN}
      - TYK_ORG_ID=${STAGING_ORG_ID}
    volumes:
      - ./apis:/workspace/apis
    profiles: ["staging"]

  deploy-production:
    build: .
    environment:
      - TYK_DASH_URL=${PROD_DASH_URL}
      - TYK_AUTH_TOKEN=${PROD_AUTH_TOKEN}
      - TYK_ORG_ID=${PROD_ORG_ID}
    volumes:
      - ./apis:/workspace/apis
    profiles: ["production"]
```

Usage:
```bash
# Deploy to staging
docker-compose --profile staging up deploy-staging

# Deploy to production  
docker-compose --profile production up deploy-production
```

## Azure DevOps

### Pipeline YAML

```yaml
# azure-pipelines.yml
trigger:
  branches:
    include:
    - main
    - develop
  paths:
    include:
    - apis/*

pool:
  vmImage: 'ubuntu-latest'

variables:
  CLI_VERSION: 'latest'

stages:
- stage: Validate
  displayName: 'Validate APIs'
  jobs:
  - job: ValidateAPIs
    displayName: 'Validate API Specifications'
    steps:
    - script: |
        for api_file in apis/*.yaml; do
          echo "Validating $api_file"
          yamllint "$api_file"
        done
      displayName: 'Validate YAML files'

- stage: DeployStaging
  displayName: 'Deploy to Staging'
  condition: in(variables['Build.SourceBranch'], 'refs/heads/develop', 'refs/heads/staging')
  dependsOn: Validate
  jobs:
  - deployment: DeployToStaging
    displayName: 'Deploy APIs to Staging'
    environment: 'staging'
    strategy:
      runOnce:
        deploy:
          steps:
          - script: |
              curl -L "https://github.com/sedkis/tyk-cli/releases/$(CLI_VERSION)/download/tyk-cli_Linux_x86_64.tar.gz" | tar xz
              sudo mv tyk /usr/local/bin/
            displayName: 'Install Tyk CLI'
          - script: |
              export TYK_DASH_URL=$(STAGING_DASH_URL)
              export TYK_AUTH_TOKEN=$(STAGING_AUTH_TOKEN)
              export TYK_ORG_ID=$(STAGING_ORG_ID)
              
              for api_file in apis/*.yaml; do
                echo "Deploying $api_file to staging"
                tyk api apply --file "$api_file"
              done
            displayName: 'Deploy to Staging'

- stage: DeployProduction
  displayName: 'Deploy to Production'
  condition: eq(variables['Build.SourceBranch'], 'refs/heads/main')
  dependsOn: DeployStaging
  jobs:
  - deployment: DeployToProduction
    displayName: 'Deploy APIs to Production'
    environment: 'production'
    strategy:
      runOnce:
        deploy:
          steps:
          - script: |
              curl -L "https://github.com/sedkis/tyk-cli/releases/$(CLI_VERSION)/download/tyk-cli_Linux_x86_64.tar.gz" | tar xz
              sudo mv tyk /usr/local/bin/
            displayName: 'Install Tyk CLI'
          - script: |
              export TYK_DASH_URL=$(PROD_DASH_URL)
              export TYK_AUTH_TOKEN=$(PROD_AUTH_TOKEN)
              export TYK_ORG_ID=$(PROD_ORG_ID)
              
              for api_file in apis/*.yaml; do
                echo "Deploying $api_file to production"
                tyk api apply --file "$api_file"
              done
            displayName: 'Deploy to Production'
```

## Best Practices

### Security
- Store credentials as encrypted secrets
- Use separate tokens for each environment
- Rotate tokens regularly
- Never commit secrets to version control

### Error Handling
```bash
# Check for deployment success
if tyk api apply --file api.yaml; then
  echo "✅ API deployed successfully"
else
  echo "❌ API deployment failed"
  exit 1
fi
```

### Rollback Strategy
```bash
# Backup before deployment
tyk api get my-api --json > backup-$(date +%Y%m%d-%H%M%S).json

# Deploy new version
if ! tyk api apply --file new-api.yaml; then
  echo "Deployment failed, rolling back..."
  tyk api apply --file backup-*.json
  exit 1
fi
```

### Validation
```bash
# Validate API spec before deployment
if ! yamllint api.yaml; then
  echo "Invalid YAML syntax"
  exit 1
fi

# Test API after deployment
sleep 5  # Wait for API to be active
if ! curl -f "https://api.company.com/health"; then
  echo "Health check failed"
  exit 1
fi
```