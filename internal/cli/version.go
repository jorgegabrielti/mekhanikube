package cli

import (
	"fmt"

	"github.com/jorgegabrielti/nautikube/internal/version"
	"github.com/spf13/cobra"
)

// NewVersionCmd creates a new version command.
func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show NautiKube version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("nautikube %s\n", version.Version)
		},
	}
}
