package output

import (
	"io"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
)

// Formatter defines the interface for output formatting.
type Formatter interface {
	// Format writes the scan results to the given writer.
	Format(w io.Writer, problems []diagnosis.Problem, summary diagnosis.Summary) error
}

// NewFormatter creates a Formatter for the given format name.
// Supported formats: "table" (default), "json", "yaml".
func NewFormatter(format string) Formatter {
	switch format {
	case "json":
		return &JSONFormatter{}
	case "yaml":
		return &YAMLFormatter{}
	default:
		return &TableFormatter{NoColor: false}
	}
}
