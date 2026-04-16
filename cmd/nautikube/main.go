package main

import (
	"fmt"
	"os"

	"github.com/jorgegabrielti/nautikube/internal/cli"
	"github.com/spf13/cobra"
)

func main() {
	// Disable Cobra's Windows mousetrap so double-clicking nautikube.exe in
	// Explorer opens the TUI directly instead of showing a "use cmd.exe" prompt.
	// Go console binaries already get a console window from Windows automatically.
	cobra.MousetrapHelpText = ""
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
