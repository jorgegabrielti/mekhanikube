# Contribuindo para o mekhanikube 🔧

Obrigado pelo seu interesse em contribuir com o mekhanikube!

## Como Contribuir

### Reportando Problemas
- Use GitHub Issues para reportar bugs
- Inclua seu SO, versão do Docker e versão do Kubernetes
- Especifique se está usando Mekhanikube v2 ou K8sGPT legacy
- Forneça passos para reproduzir o problema
- Inclua logs relevantes:
  - `docker logs mekhanikube` (v2)
  - `docker logs mekhanikube-k8sgpt` (legacy)
  - `docker logs mekhanikube-ollama`

### Sugerindo Funcionalidades
- Abra uma GitHub Issue com o rótulo "enhancement"
- Descreva o caso de uso e o comportamento esperado
- Explique como isso beneficiaria os usuários

### Pull Requests
1. Faça fork do repositório
2. Crie uma branch de funcionalidade (`git checkout -b feature/funcionalidade-incrivel`)
3. Teste suas alterações localmente
4. Faça commit com mensagens claras (`git commit -m 'Adiciona funcionalidade incrível'`)
5. Envie para seu fork (`git push origin feature/funcionalidade-incrivel`)
6. Abra um Pull Request

### Configuração de Desenvolvimento

#### Desenvolvimento Go (Mekhanikube v2)

```bash
# Clone seu fork
git clone https://github.com/SEU_USUARIO/mekhanikube.git
cd mekhanikube

# Instalar dependências Go
go mod download

# Compilar localmente
go build -o mekhanikube ./cmd/mekhanikube

# Testar localmente (requer cluster K8s ativo)
./mekhanikube analyze --explain --language Portuguese

# Ou executar diretamente
go run ./cmd/mekhanikube/main.go analyze --explain --language Portuguese
```

#### Desenvolvimento Docker

```bash
# Construir imagem Mekhanikube
docker build -f configs/Dockerfile.mekhanikube -t mekhanikube:dev .

# Iniciar stack completa
docker-compose up -d

# Baixar modelo
docker exec mekhanikube-ollama ollama pull llama3.1:8b

# Testar Mekhanikube v2
docker exec mekhanikube mekhanikube analyze --explain --language Portuguese

# Testar K8sGPT legacy (se usar profile)
docker-compose --profile k8sgpt up -d
docker exec mekhanikube-k8sgpt k8sgpt analyze --explain --language Portuguese
```

## Estrutura do Código

```
mekhanikube/
├── cmd/
│   └── mekhanikube/
│       └── main.go              # Entry point, CLI
├── internal/
│   ├── scanner/                 # Scanners de recursos K8s
│   ├── analyzer/                # Lógica de análise
│   └── ollama/                  # Cliente Ollama
├── pkg/
│   └── types/                   # Tipos compartilhados
├── configs/
│   ├── Dockerfile.mekhanikube
│   └── entrypoint-mekhanikube.sh
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
5. Verificar logs sem erros (`docker logs mekhanikube`)

### Testes Funcionais
1. Criar pods com problemas intencionais
2. Executar análise e verificar detecção
3. Testar filtros (`--filter Pod`, `--filter ConfigMap`)
4. Testar namespaces (`-n kube-system`)
5. Testar explicações IA (`--explain`)
6. Testar ambos idiomas (`--language Portuguese`, `--language English`)

## Dúvidas?

Abra uma GitHub Discussion ou Issue!

