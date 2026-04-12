package main

import (
	"fmt"
	"os"

	"github.com/jorgegabrielti/nautikube/internal/cli"
	"github.com/spf13/cobra"
)

func main() {
	ensureConsole()
	// Disable Cobra's Windows mousetrap: we allocate a console ourselves,
	// so we don't want the "open cmd.exe" splash screen on double-click.
	cobra.MousetrapHelpText = ""
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
