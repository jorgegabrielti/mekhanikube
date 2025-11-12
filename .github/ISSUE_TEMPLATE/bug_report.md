---
name: Bug Report
about: Reporte um bug ou problema no NautiKube
title: '[BUG] '
labels: 'bug'
assignees: ''

---

## 🐛 Descrição do Bug

<!-- Descreva o bug de forma clara e concisa -->

## 📋 Passos para Reproduzir

1. 
2. 
3. 

## ✅ Comportamento Esperado

<!-- O que deveria acontecer? -->

## ❌ Comportamento Atual

<!-- O que está acontecendo de errado? -->

## 🖥️ Ambiente

**Sistema Operacional:**
- [ ] Windows 10/11
- [ ] macOS (Intel)
- [ ] macOS (Apple Silicon)
- [ ] Linux (qual distro?)

**Versão do NautiKube:**
```bash
docker exec nautikube nautikube version
# Cole a saída aqui
```

**Docker:**
```bash
docker --version
docker-compose --version
# Cole as versões aqui
```

**Kubernetes:**
- [ ] Docker Desktop
- [ ] Minikube
- [ ] Kind
- [ ] EKS
- [ ] Outro: _______

## 📝 Logs

<details>
<summary>Logs do NautiKube</summary>

```
docker logs nautikube
# Cole os logs aqui
```

</details>

<details>
<summary>Logs do Ollama</summary>

```
docker logs nautikube-ollama
# Cole os logs aqui
```

</details>

## 🔍 Informações Adicionais

<!-- Screenshots, contexto adicional, etc -->

## 🎯 Gravidade

- [ ] Crítico (aplicação não funciona)
- [ ] Alto (feature principal quebrada)
- [ ] Médio (inconveniente mas há workaround)
- [ ] Baixo (problema cosmético)

## 💡 Possível Solução (opcional)

<!-- Se você tem ideia de como corrigir, descreva aqui -->
