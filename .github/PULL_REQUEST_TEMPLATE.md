## 📝 Descrição

<!-- Descreva a mudança de forma clara e concisa -->

## 🎯 Tipo de Mudança

Marque o tipo de mudança:

- [ ] 🚀 Nova feature (funcionalidade nova)
- [ ] 🐛 Bug fix (correção de bug)
- [ ] 💥 Breaking change (mudança que quebra compatibilidade)
- [ ] 📝 Documentação
- [ ] 🎨 Style (formatação, não afeta código)
- [ ] ♻️ Refactor (não é fix nem feature)
- [ ] ⚡ Performance
- [ ] ✅ Testes

## 🔗 Issue Relacionada

<!-- Se houver issue relacionada, referencie aqui -->
Closes #(issue_number)

## 📋 Checklist

Antes de submeter o PR, verifique:

- [ ] Código compila sem erros (`go build ./...`)
- [ ] Testes locais passaram (`go test ./...`)
- [ ] Docker build funciona (`docker-compose build`)
- [ ] Documentação atualizada (README, docs/)
- [ ] CHANGELOG.md atualizado na seção `[Unreleased]`
- [ ] Commits seguem [Conventional Commits](https://www.conventionalcommits.org/)
- [ ] Code review próprio realizado

## 🧪 Como Testar

<!-- Instruções passo-a-passo para testar as mudanças -->

```bash
# Exemplo:
git checkout feature/gpu-support
docker-compose build
docker-compose up -d
docker exec nautikube nautikube version
```

## 📸 Screenshots (se aplicável)

<!-- Adicione screenshots se a mudança afetar UI ou output visual -->

## 💭 Contexto Adicional

<!-- Qualquer informação adicional que ajude na revisão -->

## ⚠️ Breaking Changes

<!-- Se marcou "Breaking change" acima, descreva EXATAMENTE o que quebra e como migrar -->

---

**Checklist do Revisor:**
- [ ] Código está claro e bem estruturado
- [ ] Testes cobrem as mudanças
- [ ] Documentação está completa
- [ ] Não há problemas de segurança óbvios
