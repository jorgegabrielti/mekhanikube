# Contribuindo para o NautiKube 🔧

Obrigado pelo seu interesse em contribuir com o NautiKube!

## 📋 Índice

- [Como Contribuir](#como-contribuir)
- [Workflow de Desenvolvimento](#workflow-de-desenvolvimento)
- [Padrões de Código](#padrões-de-código)
- [Testes](#testes)
- [Documentação](#documentação)

---

## 🤝 Como Contribuir

### Reportando Bugs

Use o [template de bug report](.github/ISSUE_TEMPLATE/bug_report.md) e inclua:
- ✅ Sistema operacional e versão
- ✅ Versão do NautiKube (`docker exec nautikube nautikube version`)
- ✅ Versão do Docker e Kubernetes
- ✅ Passos para reproduzir
- ✅ Logs relevantes:
  ```bash
  docker logs nautikube
  docker logs nautikube-ollama
  ```

### Sugerindo Features

Use o [template de feature request](.github/ISSUE_TEMPLATE/feature_request.md) e descreva:
- 💡 O problema que a feature resolve
- 🎯 Casos de uso práticos
- 📋 Comportamento esperado

### Enviando Pull Requests

**Leia primeiro:** [docs/WORKFLOW.md](docs/WORKFLOW.md) - GitHub Flow completo

**Resumo rápido:**

1. **Fork** o repositório
2. **Clone** seu fork:
   ```bash
   git clone https://github.com/SEU_USUARIO/nautikube.git
   cd nautikube
   ```
3. **Crie branch** a partir da `main`:
   ```bash
   git checkout -b feature/minha-feature
   ```
4. **Desenvolva** seguindo os padrões
5. **Teste** localmente
6. **Commit** usando [Conventional Commits](https://www.conventionalcommits.org/):
   ```bash
   git commit -m "feat: adicionar suporte para X"
   ```
7. **Push** para seu fork:
   ```bash
   git push origin feature/minha-feature
   ```
8. **Abra PR** usando o [template](.github/PULL_REQUEST_TEMPLATE.md)

### Configuração de Desenvolvimento

#### Desenvolvimento Go (NautiKube v2)

```bash
# Clone seu fork
git clone https://github.com/SEU_USUARIO/NautiKube.git
cd NautiKube

# Instalar dependências Go
go mod download

# Compilar localmente
go build -o NautiKube ./cmd/NautiKube

# Testar localmente (requer cluster K8s ativo)
./NautiKube analyze --explain --language Portuguese

# Ou executar diretamente
go run ./cmd/NautiKube/main.go analyze --explain --language Portuguese
```

#### Desenvolvimento Docker

```bash
# Construir imagem NautiKube
docker build -f configs/Dockerfile.NautiKube -t NautiKube:dev .

# Iniciar stack completa
docker-compose up -d

# Baixar modelo
docker exec NautiKube-ollama ollama pull llama3.1:8b

# Testar NautiKube v2
docker exec NautiKube NautiKube analyze --explain --language Portuguese

# Testar K8sGPT legacy (se usar profile)
docker-compose --profile k8sgpt up -d
docker exec NautiKube-k8sgpt k8sgpt analyze --explain --language Portuguese
```

## Estrutura do Código

```
NautiKube/
├── cmd/
│   └── NautiKube/
│       └── main.go              # Entry point, CLI
├── internal/
│   ├── scanner/                 # Scanners de recursos K8s
│   ├── analyzer/                # Lógica de análise
│   └── ollama/                  # Cliente Ollama
├── pkg/
│   └── types/                   # Tipos compartilhados
├── configs/
│   ├── Dockerfile.NautiKube
│   └── entrypoint-NautiKube.sh
└── docs/                        # Documentação
```

## Estilo de Código

### Go
- Siga [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` para formatação
- Execute `go vet` antes de commitar
- Mantenha funções pequenas e focadas
- Documente funções públicas

### Shell Scripts
- Siga recomendações do ShellCheck
- Use `set -e` para parar em erros
- Adicione comentários explicativos

### Docker
- Use builds multi-estágio
- Minimize camadas de imagem
- Use `.dockerignore` apropriadamente
- Prefira imagens Alpine para tamanho reduzido

### Documentação
- Mantenha README.md atualizado
- Documente novas features em docs/
- Atualize CHANGELOG.md
- Use português para documentação brasileira

## Testes

Antes de enviar um PR:

### Testes Go
```bash
# Compilar código
go build ./...

# Verificar imports
go mod tidy
go mod verify

# Lint (se tiver golangci-lint instalado)
golangci-lint run
```

### Testes Docker
1. Construir imagens sem erros
2. Testar com cluster Kubernetes local (Docker Desktop, Minikube, Kind)
3. Verificar todos os comandos do README.md
4. Testar cenários de erro (cluster offline, Ollama offline)
5. Verificar logs sem erros (`docker logs NautiKube`)

### Testes Funcionais
1. Criar pods com problemas intencionais
2. Executar análise e verificar detecção
3. Testar filtros (`--filter Pod`, `--filter ConfigMap`)
4. Testar namespaces (`-n kube-system`)
5. Testar explicações IA (`--explain`)
6. Testar ambos idiomas (`--language Portuguese`, `--language English`)

## Dúvidas?

Abra uma GitHub Discussion ou Issue!

