package cli

import (
	"fmt"
	"strings"

	"github.com/jorgegabrielti/nautikube/internal/config"
	"github.com/spf13/cobra"
)

// NewConfigCmd creates the `nautikube config` command group.
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage NautiKube configuration",
	}

	cmd.AddCommand(newConfigShowCmd())
	return cmd
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show resolved configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigShow()
		},
	}
}

func runConfigShow() error {
	cfg := config.Load()
	path, _ := config.Path()

	fmt.Println("NautiKube — Resolved Configuration")
	fmt.Println(strings.Repeat("─", 40))

	severity := cfg.Severity
	if severity == "" {
		severity = "(show all)"
	}

	fmt.Printf("  Language:         %s\n", cfg.Language)
	fmt.Printf("  Minimum severity: %s\n", severity)
	fmt.Printf("  Output format:    %s\n", cfg.Output)
	fmt.Println()
	fmt.Printf("  Config file: %s\n", path)

	return nil
}
