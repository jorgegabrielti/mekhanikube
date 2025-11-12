---
name: Release Checklist
about: Template para preparar uma nova release
title: 'Release v[VERSION]'
labels: release
assignees: ''

---

## 📋 Checklist de Release v[VERSION]

### Pré-Release
- [ ] Todas as features/fixes planejadas estão implementadas
- [ ] Testes locais passando (`go test ./...`)
- [ ] Docker build funcionando (`docker-compose build`)
- [ ] Análise com IA funcionando (`docker exec nautikube nautikube analyze --explain`)
- [ ] Documentação atualizada (README, docs/)
- [ ] CHANGELOG.md atualizado com as mudanças

### Release
- [ ] Executar script de release: `.\scripts\release.ps1 -Version [VERSION]`
- [ ] Revisar commit de release
- [ ] Fazer push da tag: `git push origin v[VERSION]`
- [ ] Aguardar GitHub Actions completar o build
- [ ] Verificar release criada no GitHub
- [ ] Testar binários baixados (pelo menos Linux e Windows)
- [ ] Testar Docker image: `docker pull ghcr.io/jorgegabrielti/nautikube:[VERSION]`

### Pós-Release
- [ ] Anunciar release (README badges atualizados automaticamente)
- [ ] Atualizar documentação externa se necessário
- [ ] Fechar issues resolvidas na release
- [ ] Criar milestone para próxima versão

### Testes de Validação
- [ ] Download de binário funciona
- [ ] Binário executa: `./nautikube version`
- [ ] Docker image funciona: `docker run ghcr.io/jorgegabrielti/nautikube:[VERSION] version`
- [ ] Instalação limpa com `git clone` + `docker-compose up -d`

---

## 📝 Notas Adicionais

[Adicione aqui qualquer informação relevante sobre esta release]
