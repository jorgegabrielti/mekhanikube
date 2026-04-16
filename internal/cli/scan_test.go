package cli

import (
	"bytes"
	"strings"
	"testing"
)

// runScan validates before connecting to the cluster.
// Tests that expect a validation error check for the exact message.
// Tests that expect "valid flags" check that the error is a cluster-connection
// error, not a validation error — proving the flag was accepted.

// isClusterError returns true when err is a cluster-connection failure (not a
// flag-validation failure). This allows tests to assert that valid flag
// combinations pass validation even when no cluster is available.
func isClusterError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "failed to connect") ||
		strings.Contains(msg, "no configuration") ||
		strings.Contains(msg, "kubeconfig") ||
		strings.Contains(msg, "connect:") ||
		strings.Contains(msg, "cluster unreachable") ||
		strings.Contains(msg, "authentication failed") ||
		strings.Contains(msg, "TLS/certificate error")
}

// ---- Validation errors (no cluster needed) ----------------------------------

func TestScanCommand_Validation(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantError string
	}{
		{
			name:      "invalid min-severity",
			args:      []string{"--min-severity", "invalid"},
			wantError: "invalid severity: invalid",
		},
		{
			name:      "positional args and --resource flag are mutually exclusive",
			args:      []string{"Pod", "--resource", "Deployment"},
			wantError: "choose either positional arguments or the --resource flag",
		},
		{
			name:      "unknown flag is rejected",
			args:      []string{"--nonexistent-flag"},
			wantError: "unknown flag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewScanCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantError)
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Errorf("got error %q, want it to contain %q", err.Error(), tt.wantError)
			}
		})
	}
}

// ---- Valid flag combinations (fail only at cluster, not at validation) -------

func TestScanCommand_ValidFlagsAccepted(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "no flags (default scan)",
			args: []string{},
		},
		{
			name: "namespace flag short",
			args: []string{"-n", "default"},
		},
		{
			name: "namespace flag long",
			args: []string{"--namespace", "kube-system"},
		},
		{
			name: "resource flag short single",
			args: []string{"-r", "Pod"},
		},
		{
			name: "resource flag long single",
			args: []string{"--resource", "Deployment"},
		},
		{
			name: "resource flag comma-separated",
			args: []string{"-r", "Pod,Deployment,Service"},
		},
		{
			name: "resource flag repeated",
			args: []string{"-r", "Pod", "-r", "Node"},
		},
		{
			name: "positional resource args",
			args: []string{"Pod", "Deployment"},
		},
		{
			name: "min-severity critical",
			args: []string{"--min-severity", "critical"},
		},
		{
			name: "min-severity high",
			args: []string{"--min-severity", "high"},
		},
		{
			name: "min-severity medium",
			args: []string{"--min-severity", "medium"},
		},
		{
			name: "min-severity low",
			args: []string{"--min-severity", "low"},
		},
		{
			name: "min-severity info",
			args: []string{"--min-severity", "info"},
		},
		{
			name: "no-color flag",
			args: []string{"--no-color"},
		},
		{
			name: "kubeconfig flag",
			args: []string{"--kubeconfig", "/nonexistent/config"},
		},
		{
			name: "context flag",
			args: []string{"--context", "my-context"},
		},
		{
			name: "report-file json extension",
			args: []string{"-f", "out.json"},
		},
		{
			name: "report-file yaml extension",
			args: []string{"-f", "out.yaml"},
		},
		{
			name: "report-file csv extension",
			args: []string{"-f", "out.csv"},
		},
		{
			name: "report-file txt extension (table format)",
			args: []string{"-f", "out.txt"},
		},
		{
			name: "combined namespace and resource",
			args: []string{"-n", "production", "-r", "Pod"},
		},
		{
			name: "combined namespace resource severity",
			args: []string{"-n", "default", "-r", "Pod", "--min-severity", "high"},
		},
		{
			name: "no-color standalone",
			args: []string{"--no-color"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewScanCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			// Either succeeds (real cluster present) or fails only at the
			// cluster-connection stage — NOT at flag validation.
			if err != nil && !isClusterError(err) {
				t.Errorf("unexpected validation error for args %v: %v", tt.args, err)
			}
		})
	}
}

// ---- version command --------------------------------------------------------

func TestVersionCommand(t *testing.T) {
	root := NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"version"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("version command returned error: %v", err)
	}
	// version writes to stdout directly via fmt.Printf; check via output capture
	// is not required — absence of error is sufficient.
}

// ---- help output ------------------------------------------------------------

func TestRootCommand_Help(t *testing.T) {
	root := NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--help"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("--help returned error: %v", err)
	}
}

func TestScanCommand_Help(t *testing.T) {
	cmd := NewScanCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("scan --help returned error: %v", err)
	}
	out := buf.String()
	for _, expected := range []string{"--namespace", "--resource", "--min-severity", "--no-color", "--report-file", "--context", "--kubeconfig"} {
		if !strings.Contains(out, expected) {
			t.Errorf("scan --help output missing flag %q", expected)
		}
	}
}

// ---- formatFromExtension ----------------------------------------------------

func TestFormatFromExtension(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"report.json", "json"},
		{"report.JSON", "json"},
		{"report.yaml", "yaml"},
		{"report.yml", "yaml"},
		{"report.YAML", "yaml"},
		{"report.csv", "csv"},
		{"report.CSV", "csv"},
		{"report.txt", "table"},
		{"report", "table"},
		{"out.log", "table"},
	}
	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := formatFromExtension(tt.filename)
			if got != tt.want {
				t.Errorf("formatFromExtension(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

// ---- filterBySeverity / parseSeverityWeight ---------------------------------

func TestFilterBySeverity(t *testing.T) {
	tests := []struct {
		minSev string
		weight int
	}{
		{"critical", 5},
		{"high", 4},
		{"medium", 3},
		{"low", 2},
		{"info", 1},
		{"invalid", 0},
		{"CRITICAL", 5},
		{"High", 4},
	}
	for _, tt := range tests {
		t.Run(tt.minSev, func(t *testing.T) {
			got := parseSeverityWeight(tt.minSev)
			if got != tt.weight {
				t.Errorf("parseSeverityWeight(%q) = %d, want %d", tt.minSev, got, tt.weight)
			}
		})
	}
}
