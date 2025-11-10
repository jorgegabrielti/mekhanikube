<div align="center">

<img src="assets/logo.png" alt="NautiKube Logo" width="800"/>

**Seu navegador de diagnósticos Kubernetes com IA**

[![Licença: MIT](https://img.shields.io/badge/Licen%C3%A7a-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Vers�o](https://img.shields.io/badge/vers%C3%A3o-2.0.0-blue.svg)](https://github.com/jorgegabrielti/NautiKube/releases)
[![Go](https://img.shields.io/badge/Go-1.21-00ADD8.svg)](https://golang.org/)

Ferramenta pr�pria de an�lise de clusters Kubernetes com **IA local**  
Totalmente local � Privado � Performance otimizada � 100% em portugu�s

[Come�ar](#-in�cio-r�pido) � [Documenta��o](docs/) � [Contribuir](CONTRIBUTING.md)

</div>

---

##  O que faz?

Escaneia seu cluster Kubernetes, identifica problemas e **explica em linguagem simples** usando IA local via Ollama.

```bash
# Execute uma an�lise
docker exec NautiKube NautiKube analyze --explain --language Portuguese
```

**Exemplo de sa�da:**
```
?? Encontrados 2 problema(s):

0: Pod default/nginx-5d5d5d5d-xxx
- Error: Container nginx in CrashLoopBackOff
- IA: Este container est� reiniciando continuamente. Isso geralmente acontece 
  quando o processo principal dentro do container falha. Verifique os logs com 
  'kubectl logs nginx-5d5d5d5d-xxx' para identificar o erro espec�fico.
```

---

##  In�cio R�pido

### Pr�-requisitos
- Docker & Docker Compose
- Cluster Kubernetes ativo
- ~8GB de espa�o livre

### Instala��o

```bash
# 1. Clone o reposit�rio
git clone https://github.com/jorgegabrielti/NautiKube.git
cd NautiKube

# 2. Inicie os servi�os
docker-compose up -d

# 3. Baixe o modelo de IA (primeira vez - ~4.7GB)
docker exec NautiKube-ollama ollama pull llama3.1:8b

# 4. Pronto! Analisar cluster
docker exec NautiKube NautiKube analyze --explain --language Portuguese
```

> ?? **Novo!** N�o � mais necess�rio configurar backend. O NautiKube detecta e conecta automaticamente ao Ollama!

---

##  Comandos �teis

```bash
# An�lise r�pida (sem IA)
docker exec NautiKube NautiKube analyze

# An�lise completa em portugu�s com explica��es da IA
docker exec NautiKube NautiKube analyze --explain --language Portuguese

# An�lise completa em ingl�s
docker exec NautiKube NautiKube analyze --explain --language English

# Analisar namespace espec�fico
docker exec NautiKube NautiKube analyze -n kube-system --explain --language Portuguese

# Filtrar por tipo de recurso
docker exec NautiKube NautiKube analyze --filter Pod --explain --language Portuguese
docker exec NautiKube NautiKube analyze --filter ConfigMap

# Ver vers�o
docker exec NautiKube NautiKube version

# Listar modelos Ollama instalados
docker exec NautiKube-ollama ollama list

# Ver status dos containers
docker-compose ps
```

---

##  Modelos Dispon�veis

| Modelo | Tamanho | Velocidade | Qualidade | Portugu�s | Recomendado para |
|--------|---------|------------|-----------|-----------|------------------|
| **llama3.1:8b** ? | 4.7GB | Bom | Excelente | ????? | **Recomendado (PT-BR)** |
| **gemma2:9b** | 5.4GB | M�dio | Excelente | ????? | Melhor qualidade |
| **qwen2.5:7b** | 4.7GB | R�pido | Muito Boa | ???? | Velocidade |
| **mistral** | 4.1GB | M�dio | Boa | ??? | Uso geral |
| **tinyllama** | 1.1GB | Muito R�pido | B�sica | ?? | Scans r�pidos |

> ?? **llama3.1:8b** � o modelo padr�o por oferecer excelente suporte ao portugu�s brasileiro

**Trocar modelo:**
```bash
# Instalar outro modelo no Ollama
docker exec NautiKube-ollama ollama pull gemma2:9b

# Atualizar vari�vel de ambiente e reiniciar
# Edite .env e mude OLLAMA_MODEL=gemma2:9b
docker-compose restart NautiKube
```

---

##  Por que NautiKube pr�prio?

Desenvolvemos nossa pr�pria solu��o nativa em Go por diversos motivos:

| Aspecto | Antes | Agora | Benef�cio |
|---------|-------|-------|-----------|
| **Performance** | Startup 30s | Startup <10s | ? 3x mais r�pido |
| **Tamanho** | ~200MB | ~80MB | ?? 60% menor |
| **Configura��o** | 3 passos | Autom�tica | ?? Plug & play |
| **C�digo** | Depend�ncia externa | C�digo pr�prio | ?? Controle total |
| **Features** | Limitadas | Customiz�veis | ?? Expans�vel |
| **Manuten��o** | Dependente upstream | Independente | ? Autonomia |

**Principais vantagens:**
- ???? Suporte nativo ao portugu�s (n�o precisa flag --language)
- ?? Interface CLI mais simples e direta
- ? Detec��o autom�tica do Ollama (sem configura��o manual)
- ?? Performance otimizada para clusters pequenos e m�dios
- ?? Facilidade para adicionar novos tipos de an�lise

---

##  Solu��o de Problemas

**Container n�o inicia?**
```bash
docker-compose logs NautiKube
```

**Ollama n�o responde?**
```bash
docker logs NautiKube-ollama
docker exec NautiKube-ollama ollama list
```

**NautiKube n�o acessa o cluster?**
```bash
docker exec NautiKube kubectl get nodes
docker exec NautiKube cat /root/.kube/config_mod
```

**Erro "connection refused"?**
Certifique-se que seu cluster Kubernetes est� rodando:
```bash
kubectl cluster-info
```

---

##  Documenta��o

-  [Arquitetura](docs/ARCHITECTURE.md) - Como funciona internamente
-  [Solu��o de Problemas](docs/TROUBLESHOOTING.md) - Problemas comuns e solu��es
-  [Perguntas Frequentes](docs/FAQ.md) - D�vidas mais comuns
-  [Como Contribuir](CONTRIBUTING.md) - Guia para contribui��es

---

##  Licen�a

Licen�a MIT - consulte o arquivo [LICENSE](LICENSE) para mais detalhes.

---

##  Cr�ditos

- [Ollama](https://ollama.ai/) - Plataforma de modelos de linguagem locais
- [Kubernetes](https://kubernetes.io/) - Sistema de orquestra��o de cont�ineres

---
