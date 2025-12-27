# DeCube Documentation Index

Welcome to the DeCube documentation! This index provides an overview of all available documentation and guides for the DeCube decentralized compute platform.

## Getting Started

- **[README.md](../README.md)** - Project overview, architecture summary, and quick start guide
- **[Getting Started Guide](getting-started.md)** - Quick start for new users
- **[Setup Guide](setup.md)** - Detailed installation and configuration instructions
- **[Deployment Guide](deployment.md)** - Production deployment strategies and configurations
- **[Kubernetes Guide](kubernetes.md)** - Kubernetes deployment guide
- **[Migration Guide](migration.md)** - Upgrading and migrating between versions
- **[Integration Guide](integration.md)** - Integrating with other systems

## Architecture & Design

- **[Architecture Overview](architecture.md)** - Comprehensive system architecture and component descriptions
- **[API Documentation](api.md)** - API overview and service documentation
- **[API Reference](api-reference.md)** - Complete REST and gRPC API reference
- **[Network Protocol](network-protocol.md)** - Network protocol specifications
- **[Performance Guide](performance.md)** - Performance characteristics and metrics
- **[Performance Tuning](performance-tuning.md)** - Performance optimization strategies

## Operations & Maintenance

- **[Troubleshooting Guide](troubleshooting.md)** - Common issues, diagnosis, and resolution steps
- **[Monitoring Guide](monitoring.md)** - Monitoring and observability configuration
- **[Observability Guide](observability.md)** - Comprehensive observability setup
- **[Backup & Recovery Guide](backup-recovery.md)** - Data backup and disaster recovery procedures
- **[Disaster Recovery Plan](disaster-recovery.md)** - Comprehensive disaster recovery procedures
- **[Operational Runbook](operational-runbook.md)** - Day-to-day operational procedures
- **[Scaling Guide](scaling.md)** - Horizontal and vertical scaling strategies
- **[Capacity Planning](capacity-planning.md)** - Capacity planning and resource management
- **[Cost Optimization](cost-optimization.md)** - Cost optimization strategies
- **[Security Hardening Guide](security-hardening.md)** - Security best practices and hardening

## Development

- **[Getting Started Guide](getting-started.md)** - Quick start for new users
- **[Development Guide](development.md)** - Development practices and workflows
- **[Contributing Guide](../CONTRIBUTING.md)** - Development workflow and contribution guidelines
- **[Code of Conduct](../CODE_OF_CONDUCT.md)** - Community standards and behavior expectations
- **[Security Policy](../SECURITY.md)** - Security reporting and vulnerability management
- **[Testing Guide](testing.md)** - Testing strategies and procedures
- **[Release Process](release-process.md)** - Release procedures and checklist

## Reference

- **[API Reference](api-reference.md)** - Complete REST and gRPC API reference
- **[Network Protocol](network-protocol.md)** - Network protocol documentation
- **[FAQ](faq.md)** - Frequently asked questions
- **[Glossary](glossary.md)** - Terms and concepts
- **[Changelog](../CHANGELOG.md)** - Version history and release notes
- **[License](../LICENSE)** - Project licensing information
- **[Roadmap](roadmap.md)** - Project roadmap and future plans
- **[Community Guide](COMMUNITY.md)** - Community guidelines and resources
- **[Adopters](ADOPTERS.md)** - Organizations using DeCube

## Component Documentation

### Core Components
- **[Catalog Service](components/catalog.md)** - CRDT-based metadata management
- **[Gossip Protocol](components/gossip.md)** - State synchronization and peer discovery
- **[Consensus Layer](components/consensus.md)** - RAFT and BFT consensus
- **[Storage Layer](components/storage.md)** - CAS, object storage, and distributed ledger

### Supporting Services
- **etcd** - Distributed configuration store
- **MinIO** - Object storage backend
- **PostgreSQL** - Relational data storage (if applicable)

## CLI Tools

- **decubectl** - Primary CLI for DeCube operations
- **rechainctl** - CLI for ReChain-specific operations

## Configuration Examples

- **[Example Configuration](../config/config.example.yaml)** - Example configuration template
- **[Production Configuration](../config/config.production.yaml)** - Production configuration template
- **[Development Configuration](../config/config.development.yaml)** - Development configuration template
- **[Docker Compose](../config/docker-compose.yml)** - Development environment configuration
- **[Docker Production](../docker-compose.production.yml)** - Production Docker Compose
- **[Kubernetes Manifests](../k8s/)** - Kubernetes deployment manifests
- **[Terraform Configuration](../terraform/)** - Infrastructure as Code
- **[Ansible Playbooks](../ansible/)** - Server automation
- **[Prometheus Config](../config/prometheus.yml)** - Prometheus monitoring
- **[Alert Rules](../config/alerts.yml)** - Alerting configuration
- **[OpenAPI Specification](../api/openapi.yaml)** - OpenAPI 3.0 specification

## Vendor-Specific Deployments

- **[AWS Deployment](vendor-deployments/aws.md)** - Amazon Web Services deployment
- **[GCP Deployment](vendor-deployments/gcp.md)** - Google Cloud Platform deployment
- **[Azure Deployment](vendor-deployments/azure.md)** - Microsoft Azure deployment

## Compliance

- **[GDPR Compliance](compliance/gdpr.md)** - GDPR compliance guide
- **[SOC 2 Compliance](compliance/soc2.md)** - SOC 2 Type II compliance guide

## Testing & Quality Assurance

- **[Testing Guide](testing.md)** - Testing strategies and procedures
- **[Test Plan](../TEST_PLAN.md)** - Testing strategy and coverage requirements
- **[Integration Tests](../tests/)** - Automated test suites
- **[Benchmarks](../benchmarks/)** - Performance benchmarks and testing

## Community & Support

- **GitHub Issues** - Bug reports and feature requests
- **GitHub Discussions** - Community discussions and Q&A
- **Slack Channel** - Real-time community support (#decube on kubernetes.slack.com)
- **Mailing List** - dev@decube.io for development discussions

## External Resources

- **[Project Website](https://decube.io)** - Official project website
- **[Documentation Site](https://docs.decube.io)** - Online documentation
- **[Blog](https://blog.decube.io)** - Technical blog and announcements
- **[YouTube Channel](https://youtube.com/decube)** - Video tutorials and demos

## Quick Links

- [GitHub Repository](https://github.com/REChain-Network-Solutions/DeCub)
- [Docker Hub](https://hub.docker.com/u/decube)
- [Artifact Registry](https://console.cloud.google.com/artifacts)

## Contributing to Documentation

Documentation is maintained in the `docs/` directory. To contribute:

1. Follow the [Contributing Guide](../CONTRIBUTING.md)
2. Use Markdown format for all documentation
3. Include code examples where appropriate
4. Test documentation changes locally
5. Submit pull requests for review

## Documentation Standards

- Use clear, concise language
- Include practical examples
- Maintain consistent formatting
- Keep information up-to-date
- Cross-reference related documents

---

*Last updated: 2024-01-15*

For questions or suggestions about this documentation, please [open an issue](https://github.com/REChain-Network-Solutions/DeCub/issues) or start a [discussion](https://github.com/REChain-Network-Solutions/DeCub/discussions).
