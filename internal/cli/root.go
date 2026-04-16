package cli

import (
	"os"

	"github.com/charmbracelet/x/term"
	"github.com/jorgegabrielti/nautikube/internal/config"
	"github.com/jorgegabrielti/nautikube/internal/tui"
	"github.com/spf13/cobra"
)

// NewRootCmd creates a new root command.
func NewRootCmd() *cobra.Command {
	var namespace string

	cmd := &cobra.Command{
		Use:   "nautikube",
		Short: "Kubernetes cluster diagnostic tool",
		Long: `NautiKube scans your Kubernetes cluster, detects problems,
prioritizes them by severity score (0-100), and provides
actionable remediation commands.`,
		SilenceUsage:  false,
		SilenceErrors: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			if term.IsTerminal(os.Stdout.Fd()) {
				cfg := config.Load()
				return tui.Run(tui.Options{Namespace: namespace, Lang: cfg.Language})
			}
			return cmd.Help()
		},
	}

	cmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes namespace to scan (default: all namespaces)")

	cmd.AddCommand(NewScanCmd())
	cmd.AddCommand(NewVersionCmd())
	cmd.AddCommand(NewInitCmd())
	cmd.AddCommand(NewConfigCmd())

	return cmd
}

// Execute runs the root command.
func Execute() error {
	return NewRootCmd().Execute()
}
