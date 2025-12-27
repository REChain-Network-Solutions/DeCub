# GitHub Actions Workflows

This directory contains GitHub Actions workflows for CI/CD automation.

## Available Workflows

### CI (`ci.yml`)
Runs on every push and pull request to `main` and `develop` branches.

**Jobs:**
- **Lint**: Runs golangci-lint on all Go code
- **Test**: Runs tests for all components with coverage reporting
- **Build**: Builds all components to verify compilation
- **Docker**: Builds Docker images (only pushes on non-PR events)
- **Security**: Runs Trivy vulnerability scanner

**Features:**
- Matrix strategy for testing multiple components in parallel
- Automatic component detection (skips components without go.mod)
- Error handling with `continue-on-error` for non-critical steps
- Code coverage upload to Codecov

### Release (`release.yml`)
Runs when a version tag is pushed (e.g., `v1.0.0`).

**Jobs:**
- **Release**: Builds binaries for multiple platforms and creates GitHub release
- **Docker**: Builds and pushes Docker images to GitHub Container Registry

**Features:**
- Multi-platform binary builds (Linux, macOS, Windows)
- Automatic checksum generation
- Release notes generation
- Docker image tagging with version, major.minor, and SHA

### CodeQL Analysis (`codeql.yml`)
Runs static code analysis for security vulnerabilities.

**Schedule:** Weekly on Sundays, also on push/PR

**Features:**
- Automated security scanning
- Go language analysis
- Security and quality queries

### Stale Issues (`stale.yml`)
Automatically marks stale issues and PRs.

**Schedule:** Daily

**Features:**
- Marks issues stale after 60 days of inactivity
- Marks PRs stale after 30 days of inactivity
- Closes stale items after 7 additional days
- Exempts pinned, security, bug, and enhancement labels

### Dependabot Auto-merge (`dependabot-auto-merge.yml`)
Automatically merges Dependabot PRs that pass checks.

**Features:**
- Only merges minor/patch updates
- Requires all checks to pass
- Uses squash merge

## Workflow Configuration

### Required Secrets

For full functionality, configure these secrets in GitHub:

- `DOCKER_USERNAME`: Docker Hub username (optional, for Docker builds)
- `DOCKER_PASSWORD`: Docker Hub password (optional, for Docker builds)
- `CODECOV_TOKEN`: Codecov token (optional, for coverage reporting)

### Permissions

Workflows use minimal required permissions:
- `contents: read` - Read repository contents
- `contents: write` - Create releases (release workflow only)
- `packages: write` - Push Docker images (release workflow only)
- `security-events: write` - Upload security scan results

## Testing Workflows Locally

You can test workflows locally using [act](https://github.com/nektos/act):

```bash
# Install act
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash

# Test CI workflow
act push

# Test specific job
act -j test
```

## Troubleshooting

### Workflow Fails on Component Without go.mod
- This is expected behavior - components without go.mod are automatically skipped
- The workflow includes checks to handle missing components gracefully

### Docker Build Fails
- Ensure Dockerfile exists in the repository root
- Check that Docker secrets are configured (optional)
- Workflow will continue even if Docker build fails (non-blocking)

### Test Failures
- Tests run with `continue-on-error: true` to allow partial failures
- Check individual component test results in the workflow summary
- Fix failing tests and push again

### Coverage Upload Fails
- This is non-blocking (uses `fail_ci_if_error: false`)
- Configure `CODECOV_TOKEN` secret for better coverage reporting
- Coverage files are still generated locally

## Customization

To customize workflows:

1. Edit the workflow file in `.github/workflows/`
2. Test locally with `act` if possible
3. Push changes and monitor workflow runs
4. Adjust based on results

## Best Practices

- Keep workflows fast (use caching, parallel jobs)
- Make non-critical steps non-blocking
- Use matrix strategies for parallel execution
- Include error handling for optional steps
- Document any required secrets or configuration

