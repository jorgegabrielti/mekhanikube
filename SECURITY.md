# Security Policy

## Supported Versions

The following versions of Mekhanikube are currently being supported with security updates:

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Security Considerations

### Local Deployment Only

Mekhanikube is designed to run **locally** on your infrastructure. It should not be exposed to the public internet.

**Key Security Features**:
- ✅ No external API calls (except model downloads from Ollama)
- ✅ All data stays local
- ✅ No telemetry or tracking
- ✅ Read-only access to Kubernetes cluster
- ✅ Kubeconfig mounted as read-only

### What Mekhanikube Does NOT Do

- ❌ Does not modify your Kubernetes cluster
- ❌ Does not send data to external services
- ❌ Does not store sensitive credentials externally
- ❌ Does not expose APIs publicly

## Reporting a Vulnerability

If you discover a security vulnerability in Mekhanikube, please report it responsibly:

### How to Report

**Email**: [jorgegabrielti@gmail.com](mailto:jorgegabrielti@gmail.com)

**Subject**: `[SECURITY] Brief description of issue`

**Please include**:
1. Description of the vulnerability
2. Steps to reproduce
3. Potential impact
4. Suggested fix (if available)
5. Your contact information (optional, for follow-up)

### What to Expect

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Fix Timeline**: Depends on severity (see below)
- **Public Disclosure**: After fix is released or 90 days (whichever comes first)

### Severity Levels

#### Critical (Fix within 7 days)
- Remote code execution
- Privilege escalation
- Exposure of kubeconfig or credentials
- Cluster modification without authorization

#### High (Fix within 14 days)
- Information disclosure (cluster data)
- Denial of service affecting cluster
- Bypass of access controls

#### Medium (Fix within 30 days)
- Denial of service (local only)
- Information leaks (non-sensitive)
- Configuration issues

#### Low (Fix when possible)
- Minor issues with limited impact
- Documentation errors
- Best practice violations

## Security Best Practices

### For Users

1. **Network Isolation**
   ```yaml
   # Keep containers on host network (default)
   # Or use private Docker network
   network_mode: host
   ```

2. **Kubeconfig Protection**
   ```yaml
   # Always mount as read-only
   volumes:
     - ~/.kube/config:/root/.kube/config:ro
   ```

3. **Regular Updates**
   ```bash
   # Keep Mekhanikube updated
   git pull origin main
   make build
   make restart
   ```

4. **Limit Cluster Access**
   ```bash
   # Use service account with read-only permissions
   # Create dedicated kubeconfig for Mekhanikube
   ```

5. **Monitor Logs**
   ```bash
   # Regularly check logs for anomalies
   make logs
   ```

### For Developers

1. **Dependency Management**
   - Keep base images updated
   - Scan for vulnerabilities regularly
   - Pin dependency versions

2. **Code Review**
   - All PRs require review
   - Security-sensitive changes need extra scrutiny
   - Run security scans in CI/CD

3. **Secret Management**
   - Never commit secrets to git
   - Use `.env` files (already in `.gitignore`)
   - Rotate credentials regularly

4. **Container Security**
   - Run containers as non-root when possible
   - Minimize image size
   - Use official base images
   - Enable security scanning

## Security Testing

### Automated Security Checks

Mekhanikube includes:

- **Trivy vulnerability scanning** (in CI/CD)
- **ShellCheck** for script linting
- **Docker Compose validation**

Run locally:
```bash
# Lint configuration
make lint

# Run tests
make test

# Check health
make health
```

### Manual Security Review

Before each release:
- [ ] Review all dependencies
- [ ] Check for known vulnerabilities
- [ ] Test with minimal permissions
- [ ] Verify no sensitive data exposure
- [ ] Confirm read-only cluster access

## Known Limitations

### 1. Kubeconfig Exposure in Container

**Issue**: Kubeconfig is mounted in container filesystem.

**Mitigation**:
- Mounted as read-only
- Container filesystem is ephemeral
- Not exposed externally

**Recommendation**: Use dedicated kubeconfig with minimal permissions.

### 2. Docker Socket Access (Not Required)

**Status**: Mekhanikube does NOT require Docker socket access.

**If you see requests for `/var/run/docker.sock`**: This is not needed and should not be granted.

### 3. Host Network Mode

**Trade-off**: Host network mode simplifies connectivity but shares host network stack.

**Alternative**: Use bridge mode with explicit port mappings:
```yaml
network_mode: bridge
ports:
  - "11434:11434"
```

## Security Disclosure Policy

### Public Disclosure

Security vulnerabilities will be disclosed:

1. **After a fix is released**
2. **After 90 days** (if no fix is available)
3. **With credit to reporter** (if desired)

### Hall of Fame

We recognize security researchers who responsibly disclose vulnerabilities:

- *Be the first!*

## Compliance

### Data Privacy

Mekhanikube is designed for privacy:
- No data collection
- No external communications
- No telemetry
- GDPR compliant (no personal data processed)

### Audit Trail

For compliance purposes:
```bash
# All actions are logged
docker logs mekhanikube-k8sgpt

# Kubernetes API audit logs (in your cluster)
kubectl logs -n kube-system kube-apiserver-*
```

## Security Resources

- [Docker Security Best Practices](https://docs.docker.com/engine/security/)
- [Kubernetes Security](https://kubernetes.io/docs/concepts/security/)
- [K8sGPT Security](https://docs.k8sgpt.ai/)
- [Ollama Security](https://github.com/ollama/ollama/blob/main/docs/security.md)

## Contact

For security concerns:
- **Email**: [jorgegabrielti@gmail.com](mailto:jorgegabrielti@gmail.com)
- **GitHub Issues**: For non-sensitive issues
- **GitHub Security Advisory**: For responsible disclosure

---

**Last Updated**: 2025-11-09
