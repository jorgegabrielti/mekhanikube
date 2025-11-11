package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jorgegabrielti/nautikube/pkg/types"
)

// Client é o cliente HTTP para comunicação com Ollama
type Client struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

// New cria um novo cliente Ollama
func New(baseURL, model string) *Client {
	return &Client{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // LLMs podem demorar
		},
	}
}

// Explain envia um problema para o Ollama e retorna a explicação
func (c *Client) Explain(ctx context.Context, problem *types.Problem, language string) (string, error) {
	prompt := c.buildPrompt(problem, language)

	req := types.OllamaRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("erro ao criar requisição: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("erro ao criar HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("erro ao fazer requisição: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama retornou status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var ollamaResp types.OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	return ollamaResp.Response, nil
}

// buildPrompt constrói o prompt para o LLM baseado no problema
func (c *Client) buildPrompt(problem *types.Problem, language string) string {
	detailsStr := ""
	if len(problem.Details) > 0 {
		detailsStr = "\nDetalhes adicionais:\n"
		for _, detail := range problem.Details {
			detailsStr += fmt.Sprintf("- %s\n", detail)
		}
	}

	return fmt.Sprintf(`Você é um especialista em Kubernetes. Explique o seguinte problema de forma simples e forneça uma solução prática em português brasileiro.

Tipo de recurso: %s
Namespace: %s
Nome: %s
Erro: %s%s

Forneça:
1. Uma explicação simplificada do problema
2. Passos claros para resolver

Seja conciso e direto.`,
		problem.Kind,
		problem.Namespace,
		problem.Name,
		problem.Error,
		detailsStr,
	)
}

// Health verifica se o Ollama está acessível
func (c *Client) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ollama não está acessível: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama retornou status inválido: %d", resp.StatusCode)
	}

	return nil
}
