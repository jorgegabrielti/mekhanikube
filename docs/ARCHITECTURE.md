# Architecture

## System Overview

Mekhanikube is a containerized solution that combines K8sGPT and Ollama to provide AI-powered Kubernetes cluster analysis.

```
┌─────────────────────────────────────────────────────────────┐
│                      Host Machine                            │
│                                                              │
│  ┌────────────────────┐         ┌────────────────────┐     │
│  │  Kubernetes Cluster│         │   Docker Host      │     │
│  │                    │         │                    │     │
│  │  - Pods            │         │  ┌──────────────┐ │     │
│  │  - Services        │         │  │   Ollama     │ │     │
│  │  - Deployments     │◄────────┼──┤   Container  │ │     │
│  │  - etc.            │ K8s API │  │              │ │     │
│  │                    │         │  │  - Gemma:7b  │ │     │
│  └────────────────────┘         │  │  - Models    │ │     │
│           ▲                      │  └──────┬───────┘ │     │
│           │                      │         │         │     │
│           │ kubeconfig           │         │ HTTP    │     │
│           │ (mounted)            │         │ API     │     │
│           │                      │         │         │     │
│  ┌────────┴────────┐             │  ┌──────▼───────┐ │     │
│  │  ~/.kube/config │             │  │   K8sGPT     │ │     │
│  └─────────────────┘             │  │   Container  │ │     │
│                                  │  │              │ │     │
│                                  │  │  - Analysis  │ │     │
│                                  │  │  - Filtering │ │     │
│                                  │  └──────────────┘ │     │
│                                  └────────────────────┘     │
└─────────────────────────────────────────────────────────────┘
```

## Component Details

### K8sGPT Container

**Base Image**: Alpine Linux  
**Build Process**: Multi-stage build from official K8sGPT source  
**Purpose**: Analyzes Kubernetes cluster and identifies issues

**Key Features**:
- Reads kubeconfig from mounted volume
- Automatically adjusts configuration for container networking
- Supports multiple resource types (Pods, Services, Deployments, etc.)
- Filters and namespace-scoped analysis
- Communicates with Ollama for AI explanations

**Configuration**:
- `/root/.kube/config`: Original kubeconfig (read-only mount)
- `/root/.kube/config_mod`: Modified kubeconfig for container
- `/root/.config/k8sgpt`: Persistent K8sGPT configuration

### Ollama Container

**Base Image**: Official Ollama image  
**Purpose**: Runs LLM models locally for generating explanations

**Key Features**:
- REST API on port 11434
- Persistent model storage
- Supports multiple models (Gemma, Mistral, Llama2, etc.)
- Compatible with OpenAI API format

**Storage**:
- `/root/.ollama`: Model storage (~5-10GB per model)
- Persistent Docker volume

## Data Flow

1. **Analysis Request**:
   ```
   User → make analyze → K8sGPT container
   ```

2. **Cluster Scan**:
   ```
   K8sGPT → Kubernetes API → Resource Data
   ```

3. **Issue Detection**:
   ```
   K8sGPT → Built-in analyzers → Problem identification
   ```

4. **AI Explanation**:
   ```
   K8sGPT → Ollama API → LLM Model → Explanation
   ```

5. **Result Display**:
   ```
   K8sGPT → Console → User
   ```

## Network Architecture

### Host Network Mode (Default)

```
┌──────────────────────────────────────┐
│           Host Network               │
│                                      │
│  ┌────────────┐    ┌──────────────┐│
│  │  Ollama    │    │   K8sGPT     ││
│  │  :11434    │◄───┤  localhost   ││
│  └────────────┘    └──────────────┘│
│         ▲                           │
│         │                           │
│         ├───────────────────────────┤
│         │     host.docker.internal  │
│         ▼                           │
│  ┌────────────┐                    │
│  │ Kubernetes │                    │
│  │    API     │                    │
│  └────────────┘                    │
└──────────────────────────────────────┘
```

**Benefits**:
- Simplified networking
- Direct access to localhost services
- No port mapping conflicts
- Kubernetes API accessible via host.docker.internal

### Bridge Network Mode (Alternative)

```
┌──────────────────────────────────────┐
│       Docker Bridge Network          │
│                                      │
│  ┌────────────┐    ┌──────────────┐│
│  │  Ollama    │◄───┤   K8sGPT     ││
│  │  :11434    │    │              ││
│  └────┬───────┘    └──────────────┘│
│       │                             │
└───────┼─────────────────────────────┘
        │ Port 11434
        ▼
┌──────────────┐
│     Host     │
└──────────────┘
```

## Volume Management

### Persistent Volumes

1. **mekhanikube-ollama-data**
   - Stores LLM models
   - Size: 5-20GB depending on models
   - Location: Docker managed volume
   - Backup: `docker run --rm -v mekhanikube-ollama-data:/data -v $(pwd):/backup alpine tar czf /backup/ollama-backup.tar.gz /data`

2. **mekhanikube-k8sgpt-config**
   - Stores K8sGPT configuration
   - Size: <10MB
   - Contains: Auth tokens, backend configuration
   - Location: Docker managed volume

## Security Considerations

### Kubeconfig Access

- Mounted as **read-only** to prevent modification
- Original config preserved, working copy created in container
- Never exposed externally

### Network Security

- No external API calls (fully local)
- Ollama API not exposed to internet by default
- Container-to-container communication only

### Data Privacy

- All data stays local
- No telemetry or external connections
- Cluster data never leaves your infrastructure

## Scaling and Performance

### Resource Requirements

**Minimum**:
- CPU: 2 cores
- RAM: 4GB
- Disk: 10GB

**Recommended**:
- CPU: 4+ cores
- RAM: 8GB+
- Disk: 20GB+

### Model Performance

| Model | Size | Speed | Quality | RAM Usage |
|-------|------|-------|---------|-----------|
| TinyLlama | 1.1GB | Fast | Basic | ~2GB |
| Gemma:7b | 4.8GB | Medium | Good | ~8GB |
| Mistral | 4.1GB | Medium | Good | ~8GB |
| Llama2:13b | 7.4GB | Slow | Excellent | ~16GB |

### Optimization Tips

1. **Model Selection**: Use smaller models for faster responses
2. **Namespace Filtering**: Analyze specific namespaces to reduce scope
3. **Resource Filtering**: Focus on specific resource types
4. **Caching**: K8sGPT caches backend configuration

## Integration Points

### CI/CD Integration

```bash
# GitLab CI example
k8s-analysis:
  script:
    - make analyze > analysis.txt
    - make test
  artifacts:
    paths:
      - analysis.txt
```

### Monitoring Integration

```bash
# Export metrics to monitoring system
make analyze --output json | jq '.problems | length' | \
  curl -X POST -d @- http://prometheus-pushgateway:9091/metrics/job/k8sgpt
```

### Alerting Integration

```bash
# Fail pipeline if critical issues found
ISSUES=$(make analyze --filter=Pod | grep -c "Error")
if [ $ISSUES -gt 0 ]; then
  exit 1
fi
```

## Troubleshooting Architecture

See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for common issues and solutions related to:
- Network connectivity
- Volume permissions
- Kubeconfig mounting
- Model loading issues
