package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jorgegabrielti/nautikube/internal/analyzer"
	"github.com/jorgegabrielti/nautikube/internal/ollama"
	"github.com/jorgegabrielti/nautikube/internal/scanner"
	"github.com/jorgegabrielti/nautikube/pkg/types"
)

const (
	// Version do NautiKube
	Version = "2.0.3"
)

var (
	// Flags globais
	namespace string
	filter    []string
	explain   bool
	noCache   bool

	// Configurações Ollama (podem ser sobrescritas por variáveis de ambiente)
	ollamaURL   string
	ollamaModel string
)

func init() {
	// Configurações padrão com suporte a variáveis de ambiente
	ollamaURL = getEnv("OLLAMA_HOST", "http://host.docker.internal:11434")
	ollamaModel = getEnv("OLLAMA_MODEL", "llama3.1:8b")
}

// getEnv retorna o valor de uma variável de ambiente ou o valor padrão
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

var rootCmd = &cobra.Command{
	Use:   "nautikube",
	Short: "Seu navegador de diagnósticos Kubernetes com IA",
	Long: `NautiKube analisa seu cluster Kubernetes, identifica problemas 
e explica em linguagem simples usando IA local.`,
	SilenceUsage: true,
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analisa o cluster e identifica problemas",
	Long:  `Escaneia recursos do cluster Kubernetes e identifica problemas comuns`,
	RunE:  runAnalyze,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Mostra a versão do NautiKube",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("NautiKube v%s\n", Version)
		fmt.Printf("Engine: Go (nativo)\n")
		fmt.Printf("Ollama URL: %s\n", ollamaURL)
		fmt.Printf("Ollama Model: %s\n", ollamaModel)
	},
}

func initCommands() {
	// Flags do comando analyze
	analyzeCmd.Flags().StringVarP(&namespace, "namespace", "n", getEnv("NAUTIKUBE_DEFAULT_NAMESPACE", ""), "Namespace específico (vazio = todos)")
	analyzeCmd.Flags().StringSliceVarP(&filter, "filter", "f", []string{}, "Filtrar por tipo de recurso (Pod, ConfigMap, etc)")
	analyzeCmd.Flags().BoolVarP(&explain, "explain", "e", getEnv("NAUTIKUBE_EXPLAIN", "false") == "true", "Explicar problemas usando IA")
	analyzeCmd.Flags().BoolVar(&noCache, "no-cache", false, "Forçar análise sem cache")

	// Adiciona comandos à raiz
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(versionCmd)
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Cria scanner
	scan, err := scanner.New()
	if err != nil {
		return fmt.Errorf("erro ao criar scanner: %w", err)
	}

	// Cria cliente Ollama (apenas se --explain estiver habilitado)
	var ollamaClient *ollama.Client
	if explain {
		ollamaClient = ollama.New(ollamaURL, ollamaModel)

		// Verifica se Ollama está acessível
		if err := ollamaClient.Health(ctx); err != nil {
			return fmt.Errorf("ollama não está disponível: %w\nCertifique-se que o Ollama está rodando em %s", err, ollamaURL)
		}
	}

	// Cria analyzer
	analyze := analyzer.New(scan, ollamaClient)

	// Executa análise
	opts := types.AnalyzeOptions{
		Namespace: namespace,
		Filter:    filter,
		Explain:   explain,
		Language:  "Portuguese", // Sempre em português
		NoCache:   noCache,
	}

	fmt.Println("Analisando cluster...")
	problems, err := analyze.Analyze(ctx, opts)
	if err != nil {
		return fmt.Errorf("erro durante análise: %w", err)
	}

	// Exibe resultados
	if len(problems) == 0 {
		fmt.Println("✅ Nenhum problema encontrado!")
		return nil
	}

	fmt.Printf("\n🔍 Encontrados %d problema(s):\n\n", len(problems))

	for i, problem := range problems {
		fmt.Printf("%d: %s %s/%s\n", i, problem.Kind, problem.Namespace, problem.Name)
		fmt.Printf("- Error: %s\n", problem.Error)

		if explain && problem.Explanation != "" {
			fmt.Printf("- IA: %s\n", problem.Explanation)
		}

		if len(problem.Details) > 0 {
			fmt.Println("- Detalhes:")
			for _, detail := range problem.Details {
				fmt.Printf("  - %s\n", detail)
			}
		}

		fmt.Println()
	}

	return nil
}

func main() {
	// Inicializa comandos
	initCommands()

	// Executa comando raiz
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
		os.Exit(1)
	}
}
