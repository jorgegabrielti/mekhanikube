// Package cli — exhaustive tests for root, version, and scan commands.
//
// Strategy:
//   - Tests that validate flag *registration* use ParseFlags() — no network, always fast.
//   - Tests that validate *RunE validation errors* use Execute() — errors are returned
//     before any cluster connection is attempted.
//   - Tests for pure logic functions (parseSeverityWeight, filterBySeverity,
//     formatFromExtension) call the functions directly.
//   - Tests that would reach ScanAll (real K8s API calls) are NOT included here;
//     those belong to integration tests (//go:build integration).
package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	"github.com/jorgegabrielti/nautikube/internal/scanner"
)

// ---------------------------------------------------------------------------
// Root command
// ---------------------------------------------------------------------------

func TestRootCommand_ShortHelp(t *testing.T) {
	root := NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"-h"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("-h returned error: %v", err)
	}
	out := buf.String()
	for _, expected := range []string{"scan", "version", "NautiKube"} {
		if !strings.Contains(out, expected) {
			t.Errorf("root -h output missing %q", expected)
		}
	}
}

func TestRootCommand_NoArgs_ShowsUsage(t *testing.T) {
	root := NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{})

	// Root with no args should not return an error (Cobra shows usage by default).
	err := root.Execute()
	if err != nil {
		t.Fatalf("root with no args returned error: %v", err)
	}
}

func TestRootCommand_UnknownSubcommand(t *testing.T) {
	root := NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"nonexistent-command"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for unknown subcommand, got nil")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("got error %q, want it to contain 'unknown command'", err.Error())
	}
}

// ---------------------------------------------------------------------------
// version command
// ---------------------------------------------------------------------------

func TestVersionCommand_NoError(t *testing.T) {
	root := NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"version"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("version command returned error: %v", err)
	}
}

func TestVersionCommand_ShortHelp(t *testing.T) {
	root := NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"version", "-h"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("version -h returned error: %v", err)
	}
}

func TestVersionCommand_UnknownFlag(t *testing.T) {
	root := NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"version", "--nonexistent"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for unknown flag on version, got nil")
	}
}

// ---------------------------------------------------------------------------
// scan --help output completeness
// ---------------------------------------------------------------------------

func TestScanCommand_ShortHelp(t *testing.T) {
	cmd := NewScanCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"-h"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("scan -h returned error: %v", err)
	}
}

func TestScanCommand_Help_ShortFlags(t *testing.T) {
	cmd := NewScanCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})

	_ = cmd.Execute()
	out := buf.String()
	for _, flag := range []string{"-n", "-r", "-s", "-f"} {
		if !strings.Contains(out, flag) {
			t.Errorf("scan --help output missing short flag %q", flag)
		}
	}
}

func TestScanCommand_Help_ContainsExamples(t *testing.T) {
	cmd := NewScanCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})

	_ = cmd.Execute()
	out := buf.String()
	for _, snippet := range []string{"nautikube scan", "--min-severity", "--kubeconfig", "--no-color"} {
		if !strings.Contains(out, snippet) {
			t.Errorf("scan --help missing example snippet %q", snippet)
		}
	}
}

// parseFlagsOK is a test helper that parses flags on a fresh scan command and
// asserts the flags are accepted (no "unknown flag" error). It does NOT execute
// RunE and therefore never reaches the Kubernetes client or network.
func parseFlagsOK(t *testing.T, args []string) {
	t.Helper()
	cmd := NewScanCmd()
	if err := cmd.ParseFlags(args); err != nil {
		t.Errorf("ParseFlags(%v) failed: %v", args, err)
	}
}

// ---------------------------------------------------------------------------
// --min-severity flag registration (both -s short and --min-severity long)
// ---------------------------------------------------------------------------

func TestScanCommand_MinSeverity_FlagRegistration(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
	}{
		{"short critical", []string{"-s", "critical"}},
		{"short high", []string{"-s", "high"}},
		{"short medium", []string{"-s", "medium"}},
		{"short low", []string{"-s", "low"}},
		{"short info", []string{"-s", "info"}},
		{"long critical", []string{"--min-severity", "critical"}},
		{"long high", []string{"--min-severity", "high"}},
		{"long medium", []string{"--min-severity", "medium"}},
		{"long low", []string{"--min-severity", "low"}},
		{"long info", []string{"--min-severity", "info"}},
		// Case variants — flag accepts any string; RunE validates casing.
		{"upper CRITICAL", []string{"-s", "CRITICAL"}},
		{"mixed Critical", []string{"--min-severity", "Critical"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parseFlagsOK(t, tt.args)
		})
	}
}

// ---------------------------------------------------------------------------
// --report-file / -f flag registration — all extensions
// ---------------------------------------------------------------------------

func TestScanCommand_ReportFile_FlagRegistration(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
	}{
		{"short json", []string{"-f", "out.json"}},
		{"short yaml", []string{"-f", "out.yaml"}},
		{"short yml", []string{"-f", "out.yml"}},
		{"short csv", []string{"-f", "out.csv"}},
		{"short txt", []string{"-f", "out.txt"}},
		{"short no-ext", []string{"-f", "out"}},
		{"long json", []string{"--report-file", "out.json"}},
		{"long yaml", []string{"--report-file", "out.yaml"}},
		{"long yml", []string{"--report-file", "out.yml"}},
		{"long csv", []string{"--report-file", "out.csv"}},
		{"long txt", []string{"--report-file", "out.txt"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parseFlagsOK(t, tt.args)
		})
	}
}

// ---------------------------------------------------------------------------
// --namespace / -n flag registration
// ---------------------------------------------------------------------------

func TestScanCommand_Namespace_FlagRegistration(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
	}{
		{"short default", []string{"-n", "default"}},
		{"short kube-system", []string{"-n", "kube-system"}},
		{"short monitoring", []string{"-n", "monitoring"}},
		{"long default", []string{"--namespace", "default"}},
		{"long hyphenated", []string{"--namespace", "my-app-ns"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parseFlagsOK(t, tt.args)
		})
	}
}

// ---------------------------------------------------------------------------
// --kubeconfig and --context flag registration
// ---------------------------------------------------------------------------

func TestScanCommand_KubeconfigContext_FlagRegistration(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
	}{
		{"kubeconfig path", []string{"--kubeconfig", "/home/user/.kube/config"}},
		{"kubeconfig nonexistent", []string{"--kubeconfig", "/nonexistent/path"}},
		{"context staging", []string{"--context", "staging"}},
		{"context production", []string{"--context", "production"}},
		{"kubeconfig + context", []string{"--kubeconfig", "/path/to/config", "--context", "dev"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parseFlagsOK(t, tt.args)
		})
	}
}

// ---------------------------------------------------------------------------
// --no-color flag registration
// ---------------------------------------------------------------------------

func TestScanCommand_NoColor_FlagRegistration(t *testing.T) {
	t.Parallel()
	parseFlagsOK(t, []string{"--no-color"})
}

// ---------------------------------------------------------------------------
// --resource / -r flag registration — all 25 scanner names
// ---------------------------------------------------------------------------

// allResourceNames lists every scanner Name() registered in DefaultRegistry.
// This slice is the authoritative contract: if a scanner is renamed, this test
// will fail and catch the breaking change.
var allResourceNames = []string{
	"Pod",
	"Deployment",
	"Service",
	"Node",
	"Event",
	"Cluster",
	"ConfigMap",
	"Secret",
	"PersistentVolume",
	"PersistentVolumeClaim",
	"ServiceAccount",
	"StatefulSet",
	"DaemonSet",
	"ReplicaSet",
	"Job",
	"CronJob",
	"Ingress",
	"NetworkPolicy",
	"HorizontalPodAutoscaler",
	"Role",
	"ClusterRole",
	"RoleBinding",
	"ClusterRoleBinding",
	"PodDisruptionBudget",
	"ResourceQuota",
}

func TestScanCommand_AllResourceNames_ShortFlag_Registration(t *testing.T) {
	t.Parallel()
	for _, resource := range allResourceNames {
		t.Run(resource, func(t *testing.T) {
			t.Parallel()
			parseFlagsOK(t, []string{"-r", resource})
		})
	}
}

func TestScanCommand_AllResourceNames_LongFlag_Registration(t *testing.T) {
	t.Parallel()
	for _, resource := range allResourceNames {
		t.Run(resource, func(t *testing.T) {
			t.Parallel()
			parseFlagsOK(t, []string{"--resource", resource})
		})
	}
}

func TestScanCommand_ResourceFlag_CommaSeparated_AllResources(t *testing.T) {
	t.Parallel()
	all := strings.Join(allResourceNames, ",")
	parseFlagsOK(t, []string{"-r", all})
}

func TestScanCommand_ResourceFlag_Repeated(t *testing.T) {
	t.Parallel()
	parseFlagsOK(t, []string{"-r", "Pod", "-r", "Deployment", "-r", "Node", "-r", "Service"})
}

// ---------------------------------------------------------------------------
// Registry contract — allResourceNames must match DefaultRegistry exactly
// ---------------------------------------------------------------------------

func TestDefaultRegistry_NamesMatchContract(t *testing.T) {
	t.Parallel()
	reg := scanner.DefaultRegistry()
	registeredNames := reg.Names()

	// Every name in our contract must exist in the registry.
	registered := make(map[string]bool, len(registeredNames))
	for _, n := range registeredNames {
		registered[n] = true
	}
	for _, name := range allResourceNames {
		if !registered[name] {
			t.Errorf("allResourceNames contains %q but it is not in DefaultRegistry", name)
		}
	}

	// Registry must not have more names than our contract (catch new undocumented scanners).
	if len(registeredNames) != len(allResourceNames) {
		t.Errorf("DefaultRegistry has %d scanners, allResourceNames has %d; keep them in sync",
			len(registeredNames), len(allResourceNames))
	}
}

// ---------------------------------------------------------------------------
// Complex flag combinations (ParseFlags — no network)
// ---------------------------------------------------------------------------

func TestScanCommand_ComplexCombinations_FlagRegistration(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "namespace + resource + severity + no-color",
			args: []string{"-n", "production", "-r", "Pod", "-s", "critical", "--no-color"},
		},
		{
			name: "all filtering flags long form",
			args: []string{"--namespace", "default", "--resource", "Deployment,Service", "--min-severity", "high"},
		},
		{
			name: "kubeconfig + context + namespace + severity",
			args: []string{"--kubeconfig", "/nonexistent/config", "--context", "staging", "-n", "staging", "-s", "medium"},
		},
		{
			name: "resource short + severity short",
			args: []string{"-r", "Node", "-s", "high"},
		},
		{
			name: "report-file json + namespace + severity",
			args: []string{"-f", "out.json", "-n", "default", "-s", "low"},
		},
		{
			name: "report-file yaml + no-color",
			args: []string{"-f", "out.yaml", "--no-color"},
		},
		{
			name: "multiple resources + severity",
			args: []string{"-r", "Pod,Deployment,Node,Service", "--min-severity", "critical"},
		},
		{
			name: "context + resource + report-file",
			args: []string{"--context", "production", "-r", "Cluster", "-f", "out.csv"},
		},
		{
			name: "all long flags combined",
			args: []string{
				"--namespace", "kube-system",
				"--resource", "Pod",
				"--min-severity", "critical",
				"--no-color",
				"--kubeconfig", "/path/to/config",
				"--context", "my-context",
				"--report-file", "report.json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parseFlagsOK(t, tt.args)
		})
	}
}

// ---------------------------------------------------------------------------
// Additional validation error cases
// ---------------------------------------------------------------------------

func TestScanCommand_Validation_Extended(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantError string
	}{
		{
			name:      "uppercase invalid severity",
			args:      []string{"--min-severity", "INVALID"},
			wantError: "invalid severity: INVALID",
		},
		{
			name:      "mixed-case invalid severity",
			args:      []string{"--min-severity", "Extreme"},
			wantError: "invalid severity: Extreme",
		},
		{
			name:      "positional args and long resource flag mutually exclusive",
			args:      []string{"Pod", "--resource", "Node"},
			wantError: "choose either positional arguments or the --resource flag",
		},
		{
			name:      "positional args and short resource flag mutually exclusive",
			args:      []string{"Service", "-r", "Deployment"},
			wantError: "choose either positional arguments or the --resource flag",
		},
		{
			name:      "unknown long flag",
			args:      []string{"--verbose"},
			wantError: "unknown flag",
		},
		{
			name:      "unknown short flag",
			args:      []string{"-z"},
			wantError: "unknown shorthand flag",
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

// ---------------------------------------------------------------------------
// parseSeverityWeight — exhaustive per-case coverage
// ---------------------------------------------------------------------------

func TestParseSeverityWeight_AllCases(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"critical", 5},
		{"CRITICAL", 5},
		{"Critical", 5},
		{"cRiTiCaL", 5},
		{"high", 4},
		{"HIGH", 4},
		{"High", 4},
		{"medium", 3},
		{"MEDIUM", 3},
		{"Medium", 3},
		{"low", 2},
		{"LOW", 2},
		{"Low", 2},
		{"info", 1},
		{"INFO", 1},
		{"Info", 1},
		{"", 0},
		{"warning", 0},
		{"extreme", 0},
		{"0", 0},
		{"none", 0},
	}
	for _, tt := range tests {
		t.Run(tt.input+"="+strings.Repeat("0", tt.want)[0:0]+strings.Repeat("x", tt.want), func(t *testing.T) {
			got := parseSeverityWeight(tt.input)
			if got != tt.want {
				t.Errorf("parseSeverityWeight(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// filterBySeverity — actual Problem filtering
// ---------------------------------------------------------------------------

func makeProblems() []diagnosis.Problem {
	p := func(sev diagnosis.Severity) diagnosis.Problem {
		return diagnosis.Problem{
			Resource:  "Pod",
			Namespace: "default",
			Name:      "test-pod",
			Issue:     "test issue",
			Severity:  sev,
		}
	}
	return []diagnosis.Problem{
		p(diagnosis.Critical),
		p(diagnosis.High),
		p(diagnosis.Medium),
		p(diagnosis.Low),
		p(diagnosis.Info),
	}
}

func TestFilterBySeverity_ActualProblems(t *testing.T) {
	all := makeProblems()

	tests := []struct {
		minSev    string
		wantCount int
	}{
		{"critical", 1},
		{"high", 2},
		{"medium", 3},
		{"low", 4},
		{"info", 5},
	}

	for _, tt := range tests {
		t.Run(tt.minSev, func(t *testing.T) {
			got := filterBySeverity(all, tt.minSev)
			if len(got) != tt.wantCount {
				t.Errorf("filterBySeverity(%q) returned %d problems, want %d", tt.minSev, len(got), tt.wantCount)
			}
		})
	}
}

func TestFilterBySeverity_EmptyInput(t *testing.T) {
	got := filterBySeverity(nil, "critical")
	if len(got) != 0 {
		t.Errorf("filterBySeverity with nil input: got %d, want 0", len(got))
	}
}

func TestFilterBySeverity_AllFilteredOut(t *testing.T) {
	problems := []diagnosis.Problem{
		{Severity: diagnosis.Low},
		{Severity: diagnosis.Info},
	}
	got := filterBySeverity(problems, "critical")
	if len(got) != 0 {
		t.Errorf("expected 0 problems after filtering low/info with critical, got %d", len(got))
	}
}

func TestFilterBySeverity_NoneFiltered(t *testing.T) {
	problems := makeProblems()
	got := filterBySeverity(problems, "info")
	if len(got) != len(problems) {
		t.Errorf("filterBySeverity(info) should keep all %d problems, got %d", len(problems), len(got))
	}
}

func TestFilterBySeverity_SeveritiesPreserved(t *testing.T) {
	all := makeProblems()
	got := filterBySeverity(all, "high")

	for _, p := range got {
		if p.Severity.Weight() < 4 {
			t.Errorf("filtered problem has severity %q (weight %d) below high (weight 4)", p.Severity, p.Severity.Weight())
		}
	}
}

// ---------------------------------------------------------------------------
// formatFromExtension — extended cases
// ---------------------------------------------------------------------------

func TestFormatFromExtension_Extended(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"/tmp/report.json", "json"},
		{"/var/log/out.yaml", "yaml"},
		{"/var/log/out.yml", "yaml"},
		{"./results/scan.csv", "csv"},
		{"C:\\Users\\user\\report.JSON", "json"},
		{"C:\\Users\\user\\report.YAML", "yaml"},
		{"C:\\Users\\user\\report.YML", "yaml"},
		{"C:\\Users\\user\\report.CSV", "csv"},
		{"report.TXT", "table"},
		{"report.log", "table"},
		{"report.xml", "table"},
		{"report.html", "table"},
		{".json", "json"},
		{".yaml", "yaml"},
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
