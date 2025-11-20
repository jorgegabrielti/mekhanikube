## ✨ Novidades

- **Detecção Avançada de Provedor**: O NautiKube agora identifica visualmente o tipo de cluster (AWS EKS, Azure AKS, Google GKE, Local).
- **Conectividade Resiliente**: Nova lógica de conexão em Go com múltiplas estratégias de fallback (In-Cluster > Config Mod > Home > Env).
- **Troubleshooting Inteligente**: Dicas de resolução de problemas baseadas no tipo de erro e provedor.

## 🔧 Melhorias

- Interface de inicialização mais informativa com ícones e detalhes do ambiente.
- Mantida a correção crítica de manipulação de YAML (Python) da v2.0.4.

## 🚀 Como usar

```bash
# Clone o repositório
git clone https://github.com/jorgegabrielti/nautikube.git
cd nautikube

# Inicie os serviços
docker-compose up -d

# Execute uma análise
docker exec nautikube nautikube analyze --explain
```

## ✅ Testes Realizados

- ✅ Detecção visual de cluster local
- ✅ Conectividade via Go client (múltiplas estratégias)
- ✅ Análise completa funcional
- ✅ Build Docker bem sucedido
