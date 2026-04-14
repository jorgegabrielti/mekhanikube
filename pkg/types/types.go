package types

// Severity representa o nível de severidade de um problema
type Severity string

const (
	Critical Severity = "CRITICAL" // Problema crítico que afeta produção
	High     Severity = "HIGH"     // Problema grave que requer atenção imediata
	Medium   Severity = "MEDIUM"   // Problema moderado que deve ser resolvido
	Low      Severity = "LOW"      // Problema menor ou aviso
	Info     Severity = "INFO"     // Informacional apenas
)

// Problem representa um problema detectado no cluster
type Problem struct {
	Kind        string   `json:"kind"`        // Pod, Service, Deployment, etc
	Namespace   string   `json:"namespace"`   // Namespace do recurso
	Name        string   `json:"name"`        // Nome do recurso
	Error       string   `json:"error"`       // Descrição do erro
	Explanation string   `json:"explanation"` // Explicação da IA
	Details     []string `json:"details"`     // Detalhes adicionais
	Severity    Severity `json:"severity"`    // Nível de severidade
	Score       int      `json:"score"`       // Score numérico (0-100)
}

// String retorna uma representação em string do problema
func (p Problem) String() string {
	return p.Kind + " " + p.Namespace + "/" + p.Name + ": " + p.Error
}

// CalculateScore calcula o score numérico baseado na severidade e contexto
func (p *Problem) CalculateScore() {
	// Score base por severidade
	baseScore := map[Severity]int{
		Critical: 90,
		High:     70,
		Medium:   50,
		Low:      30,
		Info:     10,
	}

	score := baseScore[p.Severity]

	// Ajustes contextuais (+10 pontos cada)
	if p.Namespace == "kube-system" || p.Namespace == "default" {
		score += 10 // Namespaces críticos
	}

	if p.Kind == "Pod" && (ContainsCaseInsensitive(p.Error, "CrashLoopBackOff") ||
		ContainsCaseInsensitive(p.Error, "ImagePullBackOff") ||
		ContainsCaseInsensitive(p.Error, "OOMKilled")) {
		score += 10 // Problemas críticos de Pod
	}

	if p.Kind == "Service" && ContainsCaseInsensitive(p.Error, "no endpoints") {
		score += 10 // Service sem endpoints
	}

	// Garantir range 0-100
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	p.Score = score
}

// ContainsCaseInsensitive verifica se uma string contém outra (case-insensitive)
func ContainsCaseInsensitive(s, substr string) bool {
	return IndexCaseInsensitive(s, substr) >= 0
}

// IndexCaseInsensitive encontra substr em s ignorando case
func IndexCaseInsensitive(s, substr string) int {
	s = ToLower(s)
	substr = ToLower(substr)
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// ToLower converte string para lowercase
func ToLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c = c + ('a' - 'A')
		}
		result[i] = c
	}
	return string(result)
}

// ContainsString verifica se uma string está em um slice
func ContainsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
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
