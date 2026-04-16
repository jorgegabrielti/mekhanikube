# How to Integrate with CI/CD

NautiKube is heavily designed for automation. You can integrate it into GitHub Actions, GitLab CI, or Jenkins to automatically verify the health of your newly deployed application before marking a deployment as successful.

## 1. Using JSON Output

Instead of the human-readable colorized table, you can ask NautiKube to spit out pure JSON using `-o json`.

```bash
nautikube scan -n staging -o json > scan-results.json
```

The output JSON will be an array of `Problem` objects, followed by a `Summary` block if you parse the total execution.

## 2. Using YAML Output

If your automation tools prefer YAML, NautiKube provides that natively:

```bash
nautikube scan -n staging -o yaml
```

## 3. Disabling Colors for Parsing

If you must use the `table` format in a CI runner, terminal colors (ANSI escape codes) will often break the text. You can disable all color rendering:

```bash
nautikube scan --no-color
```

## 4. Example GitHub Actions Workflow

Here is an example workflow that deploys to staging, waits 30 seconds for pods to attempt starting, and then uses NautiKube to verify if any deployments are failing:

```yaml
name: Deploy and Verify

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Code
        uses: actions/checkout@v4

      - name: Setup Kustomize and Deploy
        run: |
          # Fake setup connecting to your cluster
          kubectl apply -k ./manifests/overlays/staging

      - name: Wait for Rollout
        run: sleep 30

      - name: Install NautiKube
        run: |
          curl -sL https://github.com/jorgegabrielti/nautikube/releases/latest/download/nautikube_linux_amd64.tar.gz | tar xz
          sudo mv nautikube /usr/local/bin/

      - name: Verify Staging Health (High/Critical only)
        run: |
          # Check for high/critical issues
          RESULTS=$(nautikube scan -n staging -s high -o json)
          
          # If RESULTS array is not empty '[]', fail the pipeline
          if [ "$RESULTS" != "[]" ]; then
            echo "Critical problems detected post-deployment!"
            nautikube scan -n staging -s high  # Print table for developer to read
            exit 1
          fi
          echo "Deployment verified healthy."
```
