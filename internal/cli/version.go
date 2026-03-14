package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// version is set at build time via -ldflags.
var version = "dev"

func init() {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show NautiKube version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("nautikube %s\n", version)
		},
	}

	rootCmd.AddCommand(versionCmd)
}
