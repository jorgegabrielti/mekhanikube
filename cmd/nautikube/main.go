package main

import (
	"fmt"
	"os"

	"github.com/jorgegabrielti/nautikube/internal/cli"
)

func main() {
	ensureConsole()
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
