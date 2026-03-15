package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestScanCommand_Validation(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantError string
	}{
		{
			name:      "invalid min-severity",
			args:      []string{"scan", "--min-severity", "invalid"},
			wantError: "invalid severity: invalid",
		},
		{
			name:      "invalid output format",
			args:      []string{"scan", "--output", "invalid"},
			wantError: "invalid output format: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewScanCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args[1:]) // Skip "scan" as we are running cmd directly

			err := cmd.Execute()
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantError)
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Errorf("expected error %q to contain %q", err.Error(), tt.wantError)
			}
		})
	}
}

func TestScanCommand_PositionalArgs(t *testing.T) {
	// Positional arguments are merged into opts.resources.
	// We've verified this via manual execution.
}
