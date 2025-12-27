# Ansible Playbooks for DeCube

Ansible playbooks for deploying and managing DeCube.

## Prerequisites

- Ansible >= 2.9
- SSH access to target servers
- Python 3 on target servers

## Quick Start

```bash
# Install Ansible (if not already installed)
pip install ansible

# Test connectivity
ansible all -i inventory.yml -m ping

# Run playbook
ansible-playbook -i inventory.yml playbook.yml
```

## Configuration

### Inventory

Edit `inventory.yml` to configure your servers:

- Host IP addresses
- SSH user
- Cluster configuration

### Variables

Edit `playbook.yml` to customize:

- DeCube version
- User and group
- Directory paths
- Service configuration

## Tasks

The playbook performs:

1. System package updates
2. Required package installation
3. User and group creation
4. Directory creation
5. Binary download
6. Configuration deployment
7. Systemd service setup
8. Service start

## Templates

Create template files:

- `templates/config.yaml.j2` - Configuration template
- `templates/decube.service.j2` - Systemd service template

## See Also

- [Deployment Guide](../docs/deployment.md)
- [Ansible Documentation](https://docs.ansible.com/)

