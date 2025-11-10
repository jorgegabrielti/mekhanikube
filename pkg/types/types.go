package types

// Problem representa um problema detectado no cluster
type Problem struct {
	Kind        string   `json:"kind"`        // Pod, Service, Deployment, etc
	Namespace   string   `json:"namespace"`   // Namespace do recurso
	Name        string   `json:"name"`        // Nome do recurso
	Error       string   `json:"error"`       // Descrição do erro
	Explanation string   `json:"explanation"` // Explicação da IA
	Details     []string `json:"details"`     // Detalhes adicionais
}

// String retorna uma representação em string do problema
func (p Problem) String() string {
	return p.Kind + " " + p.Namespace + "/" + p.Name + ": " + p.Error
}

// AnalyzeOptions define as opções para análise
type AnalyzeOptions struct {
	Namespace string   // Namespace específico (vazio = todos)
	Filter    []string // Filtros por tipo de recurso
	Explain   bool     // Se deve explicar com IA
	Language  string   // Idioma das explicações (Portuguese, English, etc)
	NoCache   bool     // Forçar análise sem cache
}

// OllamaRequest representa uma requisição ao Ollama
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// OllamaResponse representa uma resposta do Ollama
type OllamaResponse struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
}
