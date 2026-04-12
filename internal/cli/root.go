package cli

import (
	"os"

	"github.com/charmbracelet/x/term"
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
				return tui.Run(tui.Options{Namespace: namespace})
			}
			return cmd.Help()
		},
	}

	cmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes namespace to scan (default: all namespaces)")

	cmd.AddCommand(NewScanCmd())
	cmd.AddCommand(NewVersionCmd())

	return cmd
}

// Execute runs the root command.
func Execute() error {
	return NewRootCmd().Execute()
}
