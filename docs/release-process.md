# Release Process

This document describes the release process for DeCube.

## Table of Contents

1. [Release Cycle](#release-cycle)
2. [Release Types](#release-types)
3. [Release Checklist](#release-checklist)
4. [Release Steps](#release-steps)
5. [Post-Release](#post-release)

## Release Cycle

### Versioning

DeCube uses [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible

### Release Schedule

- **Major Releases**: As needed (breaking changes)
- **Minor Releases**: Quarterly (new features)
- **Patch Releases**: As needed (bug fixes)
- **Pre-releases**: Alpha, beta, RC versions

## Release Types

### Major Release (X.0.0)

- Breaking API changes
- Major architecture changes
- Significant feature additions
- Requires migration guide

### Minor Release (0.X.0)

- New features
- Backward compatible
- API additions
- Performance improvements

### Patch Release (0.0.X)

- Bug fixes
- Security patches
- Performance fixes
- Documentation updates

### Pre-Release

- **Alpha**: Early development, unstable
- **Beta**: Feature complete, testing
- **RC**: Release candidate, final testing

## Release Checklist

### Pre-Release

- [ ] All tests passing
- [ ] Documentation updated
- [ ] Changelog updated
- [ ] Version numbers updated
- [ ] Release notes prepared
- [ ] Security review completed
- [ ] Performance benchmarks run

### Release

- [ ] Create release branch
- [ ] Tag release
- [ ] Build binaries
- [ ] Build Docker images
- [ ] Create GitHub release
- [ ] Publish documentation
- [ ] Announce release

### Post-Release

- [ ] Monitor deployment
- [ ] Collect feedback
- [ ] Address issues
- [ ] Update roadmap

## Release Steps

### 1. Prepare Release

```bash
# Update version numbers
# Update CHANGELOG.md
# Update documentation
# Run final tests
```

### 2. Create Release Branch

```bash
git checkout -b release/v0.2.0
git push origin release/v0.2.0
```

### 3. Final Testing

```bash
# Run all tests
make test

# Run integration tests
make test-integration

# Run E2E tests
make test-e2e

# Performance benchmarks
make benchmark
```

### 4. Create Release Tag

```bash
# Create annotated tag
git tag -a v0.2.0 -m "Release v0.2.0"

# Push tag
git push origin v0.2.0
```

### 5. Build Release

```bash
# Build binaries
./scripts/build-release.sh v0.2.0

# Build Docker images
docker build -t decube:v0.2.0 .
```

### 6. Create GitHub Release

```bash
# Use GitHub CLI
gh release create v0.2.0 \
  --title "v0.2.0" \
  --notes-file RELEASE_NOTES.md \
  dist/*
```

### 7. Publish Documentation

```bash
# Update documentation site
# Publish API docs
# Update examples
```

## Release Automation

### GitHub Actions

The release workflow (`.github/workflows/release.yml`) automatically:

1. Builds binaries on tag push
2. Creates GitHub release
3. Builds and pushes Docker images
4. Generates release notes

### Manual Release

If automation fails:

```bash
# Build manually
./scripts/build-release.sh v0.2.0

# Create release manually
gh release create v0.2.0 --title "v0.2.0" --notes "..." dist/*
```

## Release Notes

### Format

```markdown
# Version X.Y.Z - Release Date

## Summary
Brief overview of release.

## What's New
- Feature 1
- Feature 2

## Improvements
- Improvement 1
- Improvement 2

## Bug Fixes
- Fixed issue #123
- Fixed issue #456

## Breaking Changes
- Change 1: Migration guide
- Change 2: Migration guide

## Contributors
- @user1
- @user2
```

## Post-Release

### Monitoring

1. **Monitor Deployments**
   - Watch for errors
   - Monitor performance
   - Check metrics

2. **Collect Feedback**
   - GitHub issues
   - User feedback
   - Performance data

3. **Address Issues**
   - Critical bugs: Hotfix
   - Minor issues: Next release
   - Documentation: Update docs

### Hotfix Process

For critical bugs:

```bash
# Create hotfix branch from release tag
git checkout -b hotfix/v0.2.1 v0.2.0

# Fix issue
# Test fix

# Create patch release
git tag -a v0.2.1 -m "Hotfix v0.2.1"
git push origin v0.2.1
```

## Release Communication

### Channels

1. **GitHub Release**: Primary announcement
2. **Release Notes**: Detailed changes
3. **Documentation**: Updated guides
4. **Community**: Discussions, forums

### Timeline

- **1 week before**: Release candidate
- **Release day**: Announcement
- **1 week after**: Follow-up, feedback

## Best Practices

### Planning

1. **Roadmap Alignment**
   - Follow roadmap
   - Set expectations
   - Communicate changes

2. **Feature Freeze**
   - Freeze features before release
   - Focus on stability
   - Test thoroughly

### Execution

1. **Automate Where Possible**
   - Use CI/CD
   - Automate builds
   - Automate testing

2. **Document Everything**
   - Release notes
   - Migration guides
   - Known issues

3. **Test Thoroughly**
   - All test suites
   - Multiple environments
   - Performance testing

## References

- [CHANGELOG.md](../CHANGELOG.md)
- [RELEASE_NOTES.md](RELEASE_NOTES.md)
- [Contributing Guide](../CONTRIBUTING.md)

---

*Last updated: January 2024*

