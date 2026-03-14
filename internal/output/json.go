package output

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
)

// JSONFormatter writes scan results as JSON.
type JSONFormatter struct{}

type jsonOutput struct {
	Timestamp string              `json:"timestamp"`
	Problems  []diagnosis.Problem `json:"problems"`
	Summary   diagnosis.Summary   `json:"summary"`
}

// Format writes the scan results as JSON.
func (f *JSONFormatter) Format(w io.Writer, problems []diagnosis.Problem, summary diagnosis.Summary) error {
	out := jsonOutput{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Problems:  problems,
		Summary:   summary,
	}

	// Use empty slice instead of nil for clean JSON output
	if out.Problems == nil {
		out.Problems = []diagnosis.Problem{}
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(out); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	return nil
}
