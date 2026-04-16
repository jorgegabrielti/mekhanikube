package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jorgegabrielti/nautikube/internal/config"
	"github.com/spf13/cobra"
)

// NewInitCmd creates the `nautikube init` command.
func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Configure default preferences",
		Long: `Interactively configure NautiKube's default preferences.
Settings are saved to ~/.nautikube/config.yaml and used as
defaults for all subsequent commands. CLI flags always override
the saved configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit()
		},
	}
}

func runInit() error {
	cfg := config.Load()
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("NautiKube — Initial Configuration")
	fmt.Println(strings.Repeat("─", 40))
	fmt.Println()

	// Language
	cfg.Language = promptChoice(reader, "Language", cfg.Language, []string{"en", "pt"})

	// Minimum severity
	sevChoices := []string{"critical", "high", "medium", "low", "info", ""}
	if cfg.Severity == "" {
		cfg.Severity = promptChoice(reader, "Minimum severity (empty = show all)", "show all", sevChoices)
		if cfg.Severity == "show all" {
			cfg.Severity = ""
		}
	} else {
		cfg.Severity = promptChoice(reader, "Minimum severity (empty = show all)", cfg.Severity, sevChoices)
	}

	// Output format
	cfg.Output = promptChoice(reader, "Default output format", cfg.Output, []string{"table", "yaml", "csv", "json"})

	if err := config.Save(cfg); err != nil {
		return err
	}

	path, _ := config.Path()
	fmt.Println()
	fmt.Printf("✅ Config saved to %s\n", path)

	return nil
}

// promptChoice asks the user to pick a value. Shows current default in brackets.
func promptChoice(reader *bufio.Reader, label, current string, choices []string) string {
	display := strings.Join(choices, "/")
	if current == "" {
		fmt.Printf("  %s [%s]: ", label, display)
	} else {
		fmt.Printf("  %s [%s] (%s): ", label, display, current)
	}

	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)

	if line == "" {
		return current
	}

	lower := strings.ToLower(line)
	for _, c := range choices {
		if strings.ToLower(c) == lower {
			return c
		}
	}

	// Invalid input — keep current
	fmt.Printf("    Invalid choice '%s', keeping '%s'\n", line, current)
	return current
}
