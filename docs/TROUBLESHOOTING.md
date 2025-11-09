# Troubleshooting Guide

## Common Issues and Solutions

### 1. K8sGPT Cannot Connect to Kubernetes API

**Symptoms**:
```
Error: unable to connect to Kubernetes cluster
```

**Causes and Solutions**:

#### A. Kubeconfig Not Found or Invalid

```bash
# Check if kubeconfig exists
ls ~/.kube/config

# Verify kubeconfig is valid
kubectl cluster-info

# Check container can see the file
docker exec mekhanikube-k8sgpt ls -l /root/.kube/config
```

**Fix**: Ensure kubeconfig path in `docker-compose.yml` is correct:
```yaml
volumes:
  - C:/Users/${USERNAME}/.kube/config:/root/.kube/config:ro
```

#### B. Wrong API Server Address

**Issue**: Kubeconfig uses `127.0.0.1` which doesn't work in containers.

**Fix**: The entrypoint.sh automatically fixes this, but verify:
```bash
docker exec mekhanikube-k8sgpt cat /root/.kube/config_mod
```

Should show `host.docker.internal` instead of `127.0.0.1`.

#### C. Kubernetes API Not Accessible

```bash
# Test from K8sGPT container
docker exec mekhanikube-k8sgpt kubectl cluster-info

# Test DNS resolution
docker exec mekhanikube-k8sgpt nslookup host.docker.internal
```

**Fix**: Ensure your Kubernetes cluster is running:
```bash
# For Docker Desktop Kubernetes
docker info | grep -i kubernetes

# For Minikube
minikube status
```

---

### 2. Ollama API Not Responding

**Symptoms**:
```
Error: failed to connect to Ollama API
Connection refused on localhost:11434
```

**Diagnosis**:
```bash
# Check Ollama container status
docker ps | grep ollama

# Check Ollama logs
docker logs mekhanikube-ollama

# Test Ollama API
curl http://localhost:11434/api/tags
```

**Solutions**:

#### A. Container Not Running
```bash
# Restart Ollama
docker-compose restart ollama

# Check health
docker inspect mekhanikube-ollama | grep Health -A 10
```

#### B. Model Not Loaded
```bash
# List installed models
docker exec mekhanikube-ollama ollama list

# Install model if missing
make install-model MODEL=gemma:7b
```

#### C. Port Conflict
```bash
# Check if port 11434 is in use
netstat -ano | findstr :11434  # Windows
lsof -i :11434                  # Linux/Mac

# Change port in .env:
OLLAMA_PORT=11435
```

---

### 3. K8sGPT Backend Not Configured

**Symptoms**:
```
Error: no backend configured
```

**Diagnosis**:
```bash
# Check K8sGPT auth status
docker exec mekhanikube-k8sgpt k8sgpt auth list
```

**Fix**:
```bash
# Reconfigure backend
make change-model MODEL=gemma:7b

# Or manually:
docker exec mekhanikube-k8sgpt k8sgpt auth add --backend ollama --model gemma:7b --baseurl http://localhost:11434
docker exec mekhanikube-k8sgpt k8sgpt auth default -p ollama
```

---

### 4. Model Download Fails

**Symptoms**:
```
Error: failed to pull model
```

**Causes**:

#### A. No Internet Connection
```bash
# Test connectivity
docker exec mekhanikube-ollama ping -c 3 ollama.ai
```

#### B. Insufficient Disk Space
```bash
# Check Docker disk usage
docker system df

# Check volume size
docker volume inspect mekhanikube-ollama-data
```

**Fix**:
```bash
# Clean up unused resources
make prune

# Or more aggressive:
docker system prune -a --volumes
```

#### C. Model Name Typo
```bash
# List available models at: https://ollama.ai/library

# Correct examples:
make install-model MODEL=gemma:7b
make install-model MODEL=mistral
make install-model MODEL=llama2
```

---

### 5. Container Won't Start

**Symptoms**:
```
Error response from daemon: container exited immediately
```

**Diagnosis**:
```bash
# Check logs
docker-compose logs

# Specific container logs
docker logs mekhanikube-k8sgpt
docker logs mekhanikube-ollama

# Check Docker Compose config
docker-compose config
```

**Common Fixes**:

#### A. Syntax Error in docker-compose.yml
```bash
# Validate configuration
docker-compose config -q
```

#### B. Volume Mount Fails
```bash
# Windows: Check path format
# Correct: C:/Users/username/.kube/config
# Wrong: C:\Users\username\.kube\config

# Linux/Mac: Check permissions
chmod 644 ~/.kube/config
```

#### C. Port Already in Use
```bash
# Find what's using the port
netstat -ano | findstr :11434  # Windows
sudo lsof -i :11434            # Linux/Mac

# Kill the process or change port
```

---

### 6. Analysis Returns No Issues

**Symptoms**:
```
No problems detected
```

**This might be normal!** But verify:

#### A. Check Specific Namespaces
```bash
# List all namespaces
docker exec mekhanikube-k8sgpt kubectl get namespaces

# Analyze specific namespace
make analyze-ns NAMESPACE=kube-system
```

#### B. Check Available Filters
```bash
# List all analyzers
make filters

# Try specific resources
make analyze-pods
make analyze-services
```

#### C. Verify Cluster Has Resources
```bash
# Check if cluster has workloads
docker exec mekhanikube-k8sgpt kubectl get all --all-namespaces
```

---

### 7. AI Explanations Not Working

**Symptoms**:
```
Issues detected but no AI explanations
```

**Checks**:

#### A. Verify --explain Flag
```bash
# Must include --explain
make analyze  # Includes --explain

# Or manually:
docker exec mekhanikube-k8sgpt k8sgpt analyze --explain
```

#### B. Check Backend Connection
```bash
# Verify backend is active
docker exec mekhanikube-k8sgpt k8sgpt auth list

# Should show:
# Active: true
# Provider: ollama
```

#### C. Model Compatibility
```bash
# Some models may not work well
# Recommended models:
make change-model MODEL=gemma:7b
make change-model MODEL=mistral
```

---

### 8. Slow Performance

**Symptoms**:
- Analysis takes >5 minutes
- System becomes unresponsive

**Optimizations**:

#### A. Use Smaller Model
```bash
# Faster but less accurate
make change-model MODEL=tinyllama

# Balanced
make change-model MODEL=gemma:7b
```

#### B. Limit Scope
```bash
# Analyze one namespace
make analyze-ns NAMESPACE=default

# Analyze specific resource type
make analyze-pods
```

#### C. Allocate More Resources
```yaml
# In docker-compose.yml, add:
services:
  ollama:
    deploy:
      resources:
        limits:
          memory: 8G
          cpus: '4'
```

---

### 9. Volume Permission Issues

**Symptoms** (Linux/Mac):
```
Error: permission denied
```

**Fix**:
```bash
# Check volume ownership
docker volume inspect mekhanikube-ollama-data

# Reset permissions
docker-compose down -v
docker-compose up -d
```

---

### 10. Network Issues on Windows

**Symptoms**:
```
Error: cannot resolve host.docker.internal
```

**Fix**:
```bash
# Ensure Docker Desktop is using WSL 2
wsl --set-default-version 2

# Or switch to bridge network mode in docker-compose.yml:
network_mode: bridge

# And update ports:
ports:
  - "11434:11434"
```

---

## Debugging Commands

### Check Overall Health
```bash
make health
```

### View All Logs
```bash
make logs
```

### Interactive Shell Access
```bash
# K8sGPT container
make shell-k8sgpt

# Ollama container
make shell-ollama
```

### Test Connectivity
```bash
# From K8sGPT to Ollama
docker exec mekhanikube-k8sgpt curl -f http://localhost:11434/api/tags

# From K8sGPT to Kubernetes
docker exec mekhanikube-k8sgpt kubectl get nodes
```

### Reset Everything
```bash
# Nuclear option - removes all data
make clean
make clean-models

# Then rebuild
make setup
```

---

## Getting Help

If you're still stuck:

1. **Check Logs**: `make logs`
2. **Run Health Check**: `make health`
3. **Run Tests**: `make test`
4. **Search Issues**: [GitHub Issues](https://github.com/jorgegabrielti/mekhanikube/issues)
5. **Open New Issue**: Include:
   - OS and Docker version
   - Output of `make health`
   - Relevant logs
   - Steps to reproduce

---

## Prevention Tips

### Before Starting

- [ ] Ensure Docker Desktop is running
- [ ] Verify Kubernetes cluster is accessible
- [ ] Check available disk space (min 10GB)
- [ ] Confirm kubeconfig path is correct

### Best Practices

- Use `make setup` for initial installation
- Run `make health` regularly
- Keep Docker Desktop updated
- Don't modify volumes manually
- Back up important configurations

### Regular Maintenance

```bash
# Weekly
make prune              # Clean unused resources

# Monthly
docker system prune -a  # Deep clean

# Before updates
make clean              # Full reset
```
