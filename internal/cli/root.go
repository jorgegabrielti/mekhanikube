package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nautikube",
	Short: "Kubernetes cluster diagnostic tool",
	Long: `NautiKube scans your Kubernetes cluster, detects problems,
prioritizes them by severity score (0-100), and provides
actionable remediation commands.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
