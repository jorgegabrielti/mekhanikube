# Frequently Asked Questions (FAQ)

## General Questions

### What is Mekhanikube?

Mekhanikube is a containerized solution that combines K8sGPT and Ollama to provide AI-powered analysis of Kubernetes clusters. It identifies issues, explains their causes, and suggests solutions using local LLM models.

### Why "Mekhanikube"?

**Mekhani** (Greek: μηχανικός) = mechanic + **kube** (Kubernetes) = Your Kubernetes mechanic!

### Is it free?

Yes! Mekhanikube is open-source under the MIT License. All components (K8sGPT, Ollama) are also free and open-source.

### Does it send my data anywhere?

No! Everything runs locally on your machine. Your cluster data never leaves your infrastructure. No telemetry, no external API calls.

---

## Installation & Setup

### What are the system requirements?

**Minimum**:
- Docker & Docker Compose
- 2 CPU cores
- 4GB RAM
- 10GB disk space
- Active Kubernetes cluster

**Recommended**:
- 4+ CPU cores
- 8GB+ RAM
- 20GB+ disk space

### What operating systems are supported?

- ✅ Windows 10/11 (with Docker Desktop)
- ✅ macOS (Intel & Apple Silicon)
- ✅ Linux (any distribution with Docker)

### Can I use it with any Kubernetes cluster?

Yes! Mekhanikube works with:
- Local clusters (Docker Desktop, Minikube, Kind)
- Cloud clusters (EKS, GKE, AKS)
- On-premise clusters
- Any cluster accessible via kubeconfig

### How long does setup take?

- First time: ~15-20 minutes (including model download)
- Subsequent starts: ~30 seconds
- Model changes: ~5-10 minutes per model

---

## Usage Questions

### Which AI model should I use?

| Model | Best For | Speed | Quality | Size |
|-------|----------|-------|---------|------|
| **gemma:7b** | Balanced (recommended) | Medium | Good | 4.8GB |
| **mistral** | Good explanations | Medium | Good | 4.1GB |
| **tinyllama** | Quick scans | Fast | Basic | 1.1GB |
| **llama2:13b** | Best quality | Slow | Excellent | 7.4GB |

Start with `gemma:7b` - it offers the best balance.

### Can I use multiple models?

Yes! Install multiple models and switch between them:

```bash
# Install additional models
make install-model MODEL=mistral
make install-model MODEL=tinyllama

# Switch active model
make change-model MODEL=mistral
```

### How often should I run analysis?

**Recommended schedule**:
- **After deployments**: Check for issues immediately
- **Daily**: Routine health checks
- **Before releases**: Validate cluster state
- **When alerts fire**: Investigate root cause

### Can I analyze only specific resources?

Yes! Use filters:

```bash
make analyze-pods        # Only Pods
make analyze-services    # Only Services
make filters             # List all filters
```

Or specific namespaces:

```bash
make analyze-ns NAMESPACE=production
```

### What types of issues can it detect?

K8sGPT analyzes:
- **Pods**: CrashLoopBackOff, ImagePullBackOff, OOMKilled
- **Services**: Endpoint issues, selector mismatches
- **Deployments**: Replica issues, update problems
- **PVCs**: Binding failures, storage issues
- **Ingress**: Configuration errors
- **StatefulSets**: Ordering issues
- **HPA**: Scaling problems
- And more!

---

## Technical Questions

### How does it work?

1. K8sGPT scans your Kubernetes cluster via the Kubernetes API
2. Built-in analyzers identify problems (e.g., pod not starting)
3. K8sGPT sends the problem context to Ollama
4. Ollama's LLM generates a human-readable explanation
5. Results are displayed with problem description, AI explanation, and suggested fixes

### Does it modify my cluster?

**No!** Mekhanikube is read-only. It:
- ✅ Reads cluster state
- ✅ Analyzes configurations
- ✅ Generates reports
- ❌ Never makes changes
- ❌ Never deletes resources
- ❌ Never applies configurations

### What permissions does it need?

K8sGPT requires **read-only** access to cluster resources. The same permissions as `kubectl get` commands.

### Can I run it in CI/CD?

Yes! Example:

```yaml
# GitLab CI
k8s-analysis:
  script:
    - make setup
    - make analyze > report.txt
  artifacts:
    paths:
      - report.txt
```

### Can I export results?

Yes, use K8sGPT's output options:

```bash
# JSON format
docker exec mekhanikube-k8sgpt k8sgpt analyze --explain --output json

# Save to file
docker exec mekhanikube-k8sgpt k8sgpt analyze --explain > analysis.txt
```

---

## Troubleshooting

### Why is it slow?

**Possible causes**:
1. **Large model**: Try `tinyllama` for faster responses
2. **Many resources**: Use filters or namespace scoping
3. **Limited RAM**: Allocate more memory to Docker
4. **CPU bottleneck**: Close other applications

**Optimization**:
```bash
# Use smaller model
make change-model MODEL=tinyllama

# Limit scope
make analyze-ns NAMESPACE=default
make analyze-pods
```

### It says "no issues found" but I know there are problems

1. **Check namespace**: Default is all namespaces
   ```bash
   make analyze-ns NAMESPACE=your-namespace
   ```

2. **Try different filters**: Some issues need specific analyzers
   ```bash
   make filters
   make analyze-pods
   ```

3. **Verify cluster access**:
   ```bash
   docker exec mekhanikube-k8sgpt kubectl get pods --all-namespaces
   ```

### Ollama keeps downloading models

Models are stored in Docker volumes. If you run `make clean-models` or `docker-compose down -v`, models are deleted.

**Preserve models**:
```bash
# Stop without removing volumes
make down

# Or only restart
make restart
```

### Can I use an external Ollama instance?

Yes! Modify `docker-compose.yml`:

```yaml
k8sgpt:
  environment:
    - OLLAMA_BASEURL=http://your-ollama-server:11434
```

Then remove the Ollama service definition.

---

## Advanced Usage

### Can I customize K8sGPT analyzers?

K8sGPT uses built-in analyzers. To enable/disable:

```bash
# List available filters
make filters

# Use specific filters
docker exec mekhanikube-k8sgpt k8sgpt analyze --filter=Pod,Service --explain
```

### Can I use a different LLM backend?

Yes! K8sGPT supports:
- Ollama (local) - default
- OpenAI (cloud)
- Azure OpenAI (cloud)
- LocalAI (local alternative)

Example for OpenAI:
```bash
docker exec mekhanikube-k8sgpt k8sgpt auth add \
  --backend openai \
  --model gpt-4 \
  --password YOUR_API_KEY
```

### How do I backup my configuration?

```bash
# Backup Ollama models
docker run --rm \
  -v mekhanikube-ollama-data:/data \
  -v $(pwd):/backup \
  alpine tar czf /backup/ollama-backup.tar.gz /data

# Backup K8sGPT config
docker run --rm \
  -v mekhanikube-k8sgpt-config:/data \
  -v $(pwd):/backup \
  alpine tar czf /backup/k8sgpt-backup.tar.gz /data
```

### Can I run multiple instances?

Yes, but change container names to avoid conflicts:

```bash
# In .env file
CONTAINER_NAME_OLLAMA=mekhanikube-ollama-2
CONTAINER_NAME_K8SGPT=mekhanikube-k8sgpt-2
OLLAMA_PORT=11435
```

### How do I update to the latest version?

```bash
# Pull latest code
git pull origin main

# Rebuild containers
make build

# Restart services
make restart
```

---

## Performance & Optimization

### How much disk space do I need?

- **Base installation**: ~500MB (containers)
- **Per model**: 1-10GB depending on model
- **Logs**: ~100MB (grows over time)
- **Recommendation**: 20GB free space

### Can I limit resource usage?

Yes! Edit `docker-compose.yml`:

```yaml
services:
  ollama:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 4G
```

### Which model is fastest?

Speed ranking (fastest to slowest):
1. `tinyllama` - ~2-5 seconds per explanation
2. `gemma:7b` - ~5-10 seconds per explanation
3. `mistral` - ~8-15 seconds per explanation
4. `llama2:13b` - ~15-30 seconds per explanation

### Can I use GPU acceleration?

Yes, if you have NVIDIA GPU:

1. Install [NVIDIA Container Toolkit](https://github.com/NVIDIA/nvidia-docker)
2. Modify `docker-compose.yml`:

```yaml
services:
  ollama:
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: 1
              capabilities: [gpu]
```

---

## Security & Privacy

### Is my kubeconfig safe?

Yes:
- Mounted as **read-only** in container
- Never modified or exposed
- Only used for API access
- Container is isolated

### What data is collected?

**None!** Mekhanikube:
- ❌ No telemetry
- ❌ No analytics
- ❌ No external connections (except model downloads)
- ❌ No data sharing

### Can I use it in production?

Yes, but:
- ✅ It's read-only (safe)
- ✅ No cluster modifications
- ⚠️ Ensure adequate resources
- ⚠️ Test in dev/staging first
- ⚠️ Monitor resource usage

### Should I commit my .env file?

**NO!** The `.env` file may contain sensitive information. It's already in `.gitignore`.

---

## Contributing & Support

### How can I contribute?

See [CONTRIBUTING.md](../CONTRIBUTING.md) for:
- Reporting bugs
- Suggesting features
- Submitting pull requests
- Improving documentation

### Where do I report bugs?

Open an issue on [GitHub Issues](https://github.com/jorgegabrielti/mekhanikube/issues) with:
- OS and Docker version
- Output of `make health`
- Steps to reproduce
- Error messages/logs

### Can I request new features?

Yes! Open a GitHub Issue with:
- Feature description
- Use case
- Expected behavior
- Example usage

### How do I get help?

1. Check this FAQ
2. Read [TROUBLESHOOTING.md](TROUBLESHOOTING.md)
3. Search [existing issues](https://github.com/jorgegabrielti/mekhanikube/issues)
4. Open a new issue
5. Join discussions

---

## Roadmap & Future

### What's planned for future versions?

- Web UI dashboard
- Slack/Teams integrations
- Custom analyzer plugins
- Historical analysis tracking
- Kubernetes operator mode
- Multi-cluster support

### Can I sponsor the project?

Not yet, but stay tuned! Meanwhile, contributions and GitHub stars are appreciated! ⭐

---

## Comparison with Other Tools

### Mekhanikube vs kubectl

- **kubectl**: Low-level commands, manual interpretation
- **Mekhanikube**: Automated analysis with AI explanations

### Mekhanikube vs K9s

- **K9s**: Interactive TUI for cluster management
- **Mekhanikube**: Automated problem detection with AI

### Mekhanikube vs Lens

- **Lens**: GUI desktop IDE for Kubernetes
- **Mekhanikube**: CLI tool with AI analysis

### Mekhanikube vs Prometheus/Grafana

- **Prometheus/Grafana**: Metrics and monitoring
- **Mekhanikube**: Issue detection and explanation

**They complement each other!** Use Mekhanikube for diagnostics alongside your existing tools.

---

## Additional Resources

- 📖 [Architecture Documentation](ARCHITECTURE.md)
- 🔧 [Troubleshooting Guide](TROUBLESHOOTING.md)
- 🤝 [Contributing Guidelines](../CONTRIBUTING.md)
- 📝 [Changelog](../CHANGELOG.md)
- 🐙 [GitHub Repository](https://github.com/jorgegabrielti/mekhanikube)
- 🔗 [K8sGPT Documentation](https://docs.k8sgpt.ai/)
- 🦙 [Ollama Documentation](https://github.com/ollama/ollama)

---

**Didn't find your answer?** Open an issue on GitHub!
