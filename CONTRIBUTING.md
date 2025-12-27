# Contributing to DeCube

Thank you for your interest in contributing to DeCube! This document provides guidelines and information for contributors.

## Code of Conduct

This project follows a code of conduct to ensure a welcoming environment for all contributors. By participating, you agree to:

- Be respectful and inclusive
- Focus on constructive feedback
- Accept responsibility for mistakes
- Show empathy towards other contributors
- Help create a positive community

## How to Contribute

### 1. Fork and Clone
```bash
git clone https://github.com/your-username/DeCub.git
cd DeCub
git remote add upstream https://github.com/REChain-Network-Solutions/DeCub.git
```

### 2. Set Up Development Environment
Follow the [setup guide](docs/setup.md) to get your development environment running.

### 3. Create a Branch
```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/issue-number-description
```

### 4. Make Changes
- Write clear, concise commit messages
- Follow the existing code style
- Add tests for new functionality
- Update documentation as needed
- Ensure all tests pass

### 5. Test Your Changes
```bash
# Run unit tests
make test-unit

# Run integration tests
make test-integration

# Test manually
docker-compose up -d
# ... test your changes
```

### 6. Submit a Pull Request
```bash
git push origin feature/your-feature-name
```
Then create a PR on GitHub with:
- Clear title and description
- Reference any related issues
- Screenshots/videos for UI changes
- Test results

## Development Guidelines

### Code Style
- **Go**: Follow standard Go formatting (`go fmt`)
- **Rust**: Use `rustfmt` and `clippy`
- **Documentation**: Use Markdown for docs, Go doc comments for code
- **Commits**: Use conventional commits format

### Testing
- Unit tests for all new code
- Integration tests for component interactions
- Chaos tests for resilience
- Performance benchmarks where applicable

### Documentation
- Update READMEs for component changes
- Add API documentation for new endpoints
- Include setup instructions for new features
- Update architecture diagrams if needed

## Areas for Contribution

### High Priority
- [ ] Enhance gossip synchronization (see TODO.md)
- [ ] Improve consensus mechanisms
- [ ] Add comprehensive monitoring
- [ ] Performance optimizations

### Medium Priority
- [ ] Additional CRDT implementations
- [ ] More storage backends
- [ ] CLI tool enhancements
- [ ] Documentation improvements

### Low Priority
- [ ] Alternative language implementations
- [ ] UI/dashboard development
- [ ] Mobile client development
- [ ] Research integrations

## Issue Reporting

### Bug Reports
Please include:
- Steps to reproduce
- Expected vs actual behavior
- Environment details (OS, versions)
- Logs and error messages
- Screenshots if applicable

### Feature Requests
Please include:
- Use case description
- Proposed implementation
- Benefits and impact
- Alternative solutions considered

## Review Process

1. **Automated Checks**: CI/CD runs tests and linting
2. **Code Review**: At least one maintainer review required
3. **Testing**: Additional testing may be requested
4. **Approval**: Maintainers approve and merge

## Getting Help

- **Documentation**: Check docs/ directory and READMEs
- **Issues**: Search existing issues before creating new ones
- **Discussions**: Use GitHub Discussions for questions
- **Community**: Join our community channels

## Recognition

Contributors are recognized through:
- GitHub contributor statistics
- Mention in release notes
- Community acknowledgments
- Potential co-authorship on publications

Thank you for contributing to DeCube and helping build the future of decentralized computing!
