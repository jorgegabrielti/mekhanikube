# Security Policy

## Supported Versions

| Version | Supported | Notes |
|---------|-----------|-------|
| 1.0.x   | ✅         | Current stable release |
| < 1.0   | ❌         | Not supported |

## Security Model

NautiKube is designed as a **read-only diagnostic tool**. It performs no writes to your Kubernetes cluster.

### What NautiKube Does

- ✅ Read-only access to the Kubernetes API (Pods, Deployments, Services, Nodes, Events)
- ✅ All processing runs locally — no external API calls
- ✅ No telemetry, tracking, or data collection
- ✅ Single static binary — no runtime dependencies

### What NautiKube Does NOT Do

- ❌ Does not modify any cluster resources
- ❌ Does not send data to external services
- ❌ Does not store credentials
- ❌ Does not require privileged access

## Reporting a Vulnerability

If you discover a security vulnerability in NautiKube, please report it responsibly.

### How to Report

**Email**: [jorgegabrielti@gmail.com](mailto:jorgegabrielti@gmail.com)

**Subject**: `[SECURITY] Brief description`

**Please include**:
1. Description of the vulnerability
2. Steps to reproduce
3. Potential impact
4. Suggested fix (if available)

### Response Timeline

| Severity | Target Fix Time |
|----------|-----------------|
| **Critical** — RCE, credential exposure, unauthorized cluster modification | 7 days |
| **High** — information disclosure, access control bypass | 14 days |
| **Medium** — local DoS, non-sensitive information leak | 30 days |
| **Low** — minor issues, best practice violations | Best effort |

### Disclosure Policy

- Vulnerabilities are disclosed **after a fix is released**, or after **90 days** (whichever comes first).
- Credit is given to the reporter upon request.

## Best Practices

### For Users

1. **Use a dedicated kubeconfig** with read-only permissions:
   ```bash
   # Create a read-only ClusterRole for NautiKube
   kubectl create clusterrolebinding nautikube-reader \
     --clusterrole=view \
     --serviceaccount=default:nautikube
   ```

2. **Keep NautiKube updated**:
   ```bash
   # Check your version
   nautikube version

   # Download latest from GitHub Releases
   ```

3. **Restrict namespace access** when possible:
   ```bash
   # Scan only specific namespaces
   nautikube scan -n production
   ```

### For Developers

1. **Dependency management** — run `govulncheck ./...` regularly
2. **Code review** — all PRs require review; security-sensitive changes need extra scrutiny
3. **No secrets in code** — never commit credentials or kubeconfig files
4. **Quality gates** — `make check` runs lint, vet, vuln, and tests before merge

## Automated Security Checks

NautiKube CI includes:

- **golangci-lint** — static analysis with 50+ linters
- **govulncheck** — known CVE scanning for Go dependencies
- **Race detector** — data race detection in tests

```bash
# Run locally
make lint          # Static analysis
make vuln          # Vulnerability scan
make test          # Tests with race detector
```

## Data Privacy

NautiKube is fully compliant with data privacy requirements:
- No data collection of any kind
- No external communications
- No telemetry or analytics
- GDPR compatible (no personal data processed)

## Security Resources

- [Kubernetes Security](https://kubernetes.io/docs/concepts/security/)
- [Go Vulnerability Database](https://vuln.go.dev/)

## Contact

- **Email**: [jorgegabrielti@gmail.com](mailto:jorgegabrielti@gmail.com)
- **GitHub Issues**: For non-sensitive problems
- **GitHub Security Advisory**: For responsible disclosure

---

**Last Updated**: 2026-03-14
