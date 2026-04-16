package output

import (
	"fmt"
	"io"
	"time"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	"gopkg.in/yaml.v3"
)

// YAMLFormatter writes scan results as YAML.
type YAMLFormatter struct{}

type yamlOutput struct {
	Timestamp string              `yaml:"timestamp"`
	Problems  []diagnosis.Problem `yaml:"problems"`
	Summary   diagnosis.Summary   `yaml:"summary"`
}

// Format writes the scan results as YAML.
func (f *YAMLFormatter) Format(w io.Writer, problems []diagnosis.Problem, summary diagnosis.Summary) error {
	out := yamlOutput{
		Timestamp: summary.ScanTime.UTC().Format(time.RFC3339),
		Problems:  problems,
		Summary:   summary,
	}

	if out.Problems == nil {
		out.Problems = []diagnosis.Problem{}
	}

	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2)
	if err := encoder.Encode(out); err != nil {
		return fmt.Errorf("failed to encode YAML: %w", err)
	}
	return nil
}
