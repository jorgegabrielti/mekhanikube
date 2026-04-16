package tui

// translations holds all translatable TUI strings keyed by language code.
var translations = map[string]map[string]string{
	"en": {
		// viewInitConfig
		"init_title":     "Initial Configuration",
		"step_lang":      "Language",
		"step_severity":  "Minimum Severity",
		"step_output":    "Output Format",
		"step_retention": "History Retention",
		"sev_all":        "(all)",
		"help_init":      "  [↑/↓] navigate  [enter] select  [q] quit",

		// retention labels
		"ret_disabled":  "Disabled (do not save history)",
		"ret_30":        "30 days",
		"ret_90":        "90 days (default)",
		"ret_180":       "180 days",
		"ret_365":       "365 days (1 year)",
		"ret_unlimited": "Unlimited (never delete)",

		// viewSetup
		"setup_title": "Select a Kubernetes context:",
		"help_setup":  "  [↑/↓] navigate  [enter] select  [s] settings  [q] quit",

		// viewProgress
		"current_context": "current context",
		"all_namespaces":  "all namespaces",
		"connecting":      "Connecting to cluster",
		"scanning_res":    "Scanning resources",
		"connected":       "Connected to cluster",

		// viewResults
		"ns_label":       "Namespace",
		"problems_found": "Problems found",
		"no_problems":    "No problems detected. Cluster looks healthy!",
		"col_severity":   "SEVERITY",
		"col_resource":   "RESOURCE",
		"col_name":       "NAME",
		"col_score":      "SCORE",
		"col_issue":      "ISSUE",
		"showing":        "Showing",
		"of":             "of",
		"help_results":   "  [↑/↓] navigate  [enter] detail  [e] export  [r] rescan  [c] context  [s] settings  [q] quit",

		// viewDetail
		"detail_title":      "Problem Detail",
		"field_severity":    "Severity",
		"field_resource":    "Resource",
		"field_namespace":   "Namespace",
		"field_name":        "Name",
		"field_score":       "Score",
		"field_issue":       "Issue",
		"field_offending":   "Offending Property",
		"field_explanation": "Explanation",
		"field_remediation": "Remediation",
		"field_details":     "Details",
		"problem_x_of_y":    "Problem %d of %d",
		"help_detail":       "  [↑/↓] next/prev  [1-9] run cmd  [e] export  [esc] back  [c] context  [s] settings  [q] quit",

		// viewOutput
		"running":     "Running...",
		"cmd_output":  "Command Output",
		"no_output":   "(no output)",
		"help_output": "  [↑/↓] scroll  [esc] back  [q] quit",

		// viewReportMenu
		"export_report": "Export Report",
		"full_report":   "Full report",
		"issues":        "issues",
		"current_issue": "Current issue only",
		"help_report":   "  [↑/↓] select  [enter] confirm  [esc] back  [q] quit",

		// viewReportFormat
		"export_format": "Export Format",
		"txt_desc":      "TXT  — plain text",
		"csv_desc":      "CSV  — spreadsheet (Excel / Google Sheets)",
		"pdf_desc":      "PDF  — formatted document",

		// viewReportSaved
		"export_failed": "Export failed",
		"report_saved":  "✓ Report saved",
		"any_key":       "any key to continue",

		// viewError
		"conn_failed": "Connection failed",
		"scan_failed": "Scan failed",
		"error":       "Error",
		"help_error":  "  [r] retry  [c] context  [q] quit",
	},
	"pt": {
		// viewInitConfig
		"init_title":     "Configuração Inicial",
		"step_lang":      "Idioma",
		"step_severity":  "Severidade Mínima",
		"step_output":    "Formato de Saída",
		"step_retention": "Retenção de Histórico",
		"sev_all":        "(todas)",
		"help_init":      "  [↑/↓] navegar  [enter] selecionar  [q] sair",

		// retention labels
		"ret_disabled":  "Desabilitado (não salvar histórico)",
		"ret_30":        "30 dias",
		"ret_90":        "90 dias (padrão)",
		"ret_180":       "180 dias",
		"ret_365":       "365 dias (1 ano)",
		"ret_unlimited": "Ilimitado (nunca apagar)",

		// viewSetup
		"setup_title": "Selecione um contexto Kubernetes:",
		"help_setup":  "  [↑/↓] navegar  [enter] selecionar  [s] configurações  [q] sair",

		// viewProgress
		"current_context": "contexto atual",
		"all_namespaces":  "todos os namespaces",
		"connecting":      "Conectando ao cluster",
		"scanning_res":    "Verificando recursos",
		"connected":       "Conectado ao cluster",

		// viewResults
		"ns_label":       "Namespace",
		"problems_found": "Problemas encontrados",
		"no_problems":    "Nenhum problema detectado. O cluster parece saudável!",
		"col_severity":   "SEVERIDADE",
		"col_resource":   "RECURSO",
		"col_name":       "NOME",
		"col_score":      "SCORE",
		"col_issue":      "PROBLEMA",
		"showing":        "Mostrando",
		"of":             "de",
		"help_results":   "  [↑/↓] navegar  [enter] detalhe  [e] exportar  [r] re-escanear  [c] contexto  [s] configurações  [q] sair",

		// viewDetail
		"detail_title":      "Detalhe do Problema",
		"field_severity":    "Severidade",
		"field_resource":    "Recurso",
		"field_namespace":   "Namespace",
		"field_name":        "Nome",
		"field_score":       "Score",
		"field_issue":       "Problema",
		"field_offending":   "Propriedade Problemática",
		"field_explanation": "Explicação",
		"field_remediation": "Remediação",
		"field_details":     "Detalhes",
		"problem_x_of_y":    "Problema %d de %d",
		"help_detail":       "  [↑/↓] próx/ant  [1-9] executar cmd  [e] exportar  [esc] voltar  [c] contexto  [s] configurações  [q] sair",

		// viewOutput
		"running":     "Executando...",
		"cmd_output":  "Saída do Comando",
		"no_output":   "(sem saída)",
		"help_output": "  [↑/↓] rolar  [esc] voltar  [q] sair",

		// viewReportMenu
		"export_report": "Exportar Relatório",
		"full_report":   "Relatório completo",
		"issues":        "problemas",
		"current_issue": "Apenas problema atual",
		"help_report":   "  [↑/↓] selecionar  [enter] confirmar  [esc] voltar  [q] sair",

		// viewReportFormat
		"export_format": "Formato de Exportação",
		"txt_desc":      "TXT  — texto simples",
		"csv_desc":      "CSV  — planilha (Excel / Google Sheets)",
		"pdf_desc":      "PDF  — documento formatado",

		// viewReportSaved
		"export_failed": "Falha na exportação",
		"report_saved":  "✓ Relatório salvo",
		"any_key":       "qualquer tecla para continuar",

		// viewError
		"conn_failed": "Falha na conexão",
		"scan_failed": "Falha na verificação",
		"error":       "Erro",
		"help_error":  "  [r] tentar novamente  [c] contexto  [q] sair",
	},
}

// t returns the translated string for the given key, falling back to English.
func t(lang, key string) string {
	if m, ok := translations[lang]; ok {
		if s, ok := m[key]; ok {
			return s
		}
	}
	if s, ok := translations["en"][key]; ok {
		return s
	}
	return key
}
