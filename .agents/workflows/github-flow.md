---
description: Implementação obrigatória do fluxo de versionamento (GitHub Flow + SemVer)
---

# GitHub Flow & Versioning Strategy

Este projeto segue o **GitHub Flow** com **Conventional Commits** e **Semantic Versioning**. Agentes e desenvolvedores **nunca** devem fazer commits diretos na branch `main`.

## Regras de Ouro

1. **A `main` é estrita**: Nenhum código deve entrar na `main` sem passar por um Pull Request e pelo CI (`make check`).
2. **Vida curta**: Branches de feature devem ser focadas e pequenas.
3. **Conventional Commits**: O go-releaser usa os commits para gerar o `CHANGELOG.md`.

## Workflow de Desenvolvimento

### 1. Iniciar Trabalho
Crie uma branch descritiva a partir da `main` atualizada.
```bash
git checkout main
git pull origin main
git checkout -b <tipo>/<nome-curto>
# Exemplos: feat/ingress-scanner, fix/nil-pointer-score, docs/update-readme
```

### 2. Implementar e Testar (Spec-Driven)
Escreva o código e os testes.
Antes de commitar, você **deve** garantir a qualidade:
```bash
make check
```

### 3. Commitar (Conventional Commits)
Siga o formato `<tipo>[escopo opcional]: <descrição em inglês ou contexto claro>`.
```bash
git add .
git commit -m "feat(scanner): add support for Ingress resources" -m "This adds detection for missing TLS secrets."
```

*Tipos válidos:* `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`, `ci`.

### 4. Abrir Pull Request
Envie a branch e abra um PR contra a `main`.
```bash
git push -u origin HEAD
```
Preencha o `.github/pull_request_template.md` no GitHub.

### 5. Finalização das Tarefas pelo Agente
Se você (Agente de IA) receber um pedido para implementar uma feature ou resolver um bug:
1. **NÃO USE** o terminal para commitar direto na `main`.
2. Peça autorização ao usuário para criar a branch `feat/...` ou `fix/...`.
3. Escreva o código.
4. Rode os testes (`make check`).
5. Gere os comandos de commit via ferramenta CLI ou informe o usuário como prosseguir com o Push/PR.
