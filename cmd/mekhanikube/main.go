package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jorgegabrielti/mekhanikube/internal/analyzer"
	"github.com/jorgegabrielti/mekhanikube/internal/ollama"
	"github.com/jorgegabrielti/mekhanikube/internal/scanner"
	"github.com/jorgegabrielti/mekhanikube/pkg/types"
)

var (
	// Flags globais
	namespace string
	filter    []string
	explain   bool
	language  string
	noCache   bool

	// Configurações Ollama
	ollamaURL   = "http://localhost:11434"
	ollamaModel = "llama3.1:8b"
)

var rootCmd = &cobra.Command{
	Use:   "mekhanikube",
	Short: "Seu mecânico de Kubernetes com IA",
	Long: `Mekhanikube analisa seu cluster Kubernetes, identifica problemas 
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
	Short: "Mostra a versão do Mekhanikube",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Mekhanikube v1.0.0")
	},
}

func init() {
	// Flags do comando analyze
	analyzeCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Namespace específico (vazio = todos)")
	analyzeCmd.Flags().StringSliceVarP(&filter, "filter", "f", []string{}, "Filtrar por tipo de recurso (Pod, ConfigMap, etc)")
	analyzeCmd.Flags().BoolVarP(&explain, "explain", "e", false, "Explicar problemas usando IA")
	analyzeCmd.Flags().StringVarP(&language, "language", "l", "Portuguese", "Idioma das explicações (Portuguese, English)")
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
		Language:  language,
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
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
