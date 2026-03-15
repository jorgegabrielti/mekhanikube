package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCmd creates a new root command.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nautikube",
		Short: "Kubernetes cluster diagnostic tool",
		Long: `NautiKube scans your Kubernetes cluster, detects problems,
prioritizes them by severity score (0-100), and provides
actionable remediation commands.`,
		SilenceUsage:  false,
		SilenceErrors: false,
	}

	cmd.AddCommand(NewScanCmd())
	cmd.AddCommand(NewVersionCmd())

	return cmd
}

// Execute runs the root command.
func Execute() error {
	return NewRootCmd().Execute()
}
