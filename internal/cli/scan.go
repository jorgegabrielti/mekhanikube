package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	"github.com/jorgegabrielti/nautikube/internal/k8s"
	"github.com/jorgegabrielti/nautikube/internal/output"
	"github.com/jorgegabrielti/nautikube/internal/scanner"
)

type scanOptions struct {
	namespace   string
	resources   resourceListValue
	minSeverity string
	noColor     bool
	kubeconfig  string
	kubecontext string
	reportFile  string
}

// resourceListValue implements pflag.Value to provide a better UX type name
type resourceListValue []string

func (s *resourceListValue) String() string {
	return strings.Join(*s, ",")
}

func (s *resourceListValue) Set(val string) error {
	*s = append(*s, strings.Split(val, ",")...)
	return nil
}

func (s *resourceListValue) Type() string {
	return "<RESOURCE>"
}

// namedStringValue is a pflag.Value that lets us control the type placeholder
// shown in --help output (e.g. <NAMESPACE> instead of "string").
type namedStringValue struct {
	ptr      *string
	typeName string
}

func (s *namedStringValue) String() string       { return *s.ptr }
func (s *namedStringValue) Set(val string) error { *s.ptr = val; return nil }
func (s *namedStringValue) Type() string         { return s.typeName }

// NewScanCmd creates a new scan command.
func NewScanCmd() *cobra.Command {
	opts := &scanOptions{}

	scanCmd := &cobra.Command{
		Use:   "scan [resource...]",
		Short: "Scan the cluster for problems",
		Long: `Scan your Kubernetes cluster for common problems such as crashing pods,
unavailable deployments, services without endpoints, and unhealthy nodes.

Arguments:
  [resource...]  Optional list of resource types to scan (e.g., Pod, Deployment, Cluster).
                 Equivalent to using the --resource flag.

Results are sorted by severity score (highest first) and include
actionable remediation commands.`,
		Example: `  # Scan all resources in all namespaces
  nautikube scan

  # Scan specific resource types
  nautikube scan Pod Deployment Node
  nautikube scan Cluster --min-severity critical
  nautikube scan -r Pod,Deployment,Service

  # Scope to a namespace and filter by severity
  nautikube scan -n production
  nautikube scan -n kube-system --min-severity critical
  nautikube scan -n default -r Pod --min-severity high

  # Export results to a file (format is detected from the file extension)
  nautikube scan -f report.json
  nautikube scan -f report.yaml -n default
  nautikube scan -f report.csv --min-severity medium

  # Use a specific kubeconfig or context
  nautikube scan --kubeconfig ~/.kube/config --context staging
  nautikube scan --context production --min-severity high

  # Disable color (useful for CI/CD pipelines)
  nautikube scan --no-color`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Prevent confusion by making positional args and --resource flag mutually exclusive
			if len(args) > 0 && cmd.Flags().Changed("resource") {
				return fmt.Errorf("choose either positional arguments or the --resource flag, but not both (e.g., 'nautikube scan Pod' OR 'nautikube scan --resource Pod')")
			}

			// Merge positional args into resources filter
			for _, arg := range args {
				if arg != "" {
					opts.resources = append(opts.resources, arg)
				}
			}

			// Validate min-severity
			if opts.minSeverity != "" && parseSeverityWeight(opts.minSeverity) == 0 {
				return fmt.Errorf("invalid severity: %s (allowed: critical, high, medium, low, info)", opts.minSeverity)
			}

			return runScan(opts)
		},
	}

	scanCmd.Flags().VarP(&namedStringValue{ptr: &opts.namespace, typeName: "<NAMESPACE>"}, "namespace", "n", "Scan a specific namespace (default: all namespaces)")
	scanCmd.Flags().VarP(&opts.resources, "resource", "r", "Filter by resource type, comma-separated (e.g., Pod,Deployment,Node)")
	scanCmd.Flags().VarP(&namedStringValue{ptr: &opts.minSeverity, typeName: "<LEVEL>"}, "min-severity", "s", "Minimum severity level to display: critical, high, medium, low, info")
	scanCmd.Flags().BoolVar(&opts.noColor, "no-color", false, "Disable colored output (automatically set when writing to a file)")
	scanCmd.Flags().Var(&namedStringValue{ptr: &opts.kubeconfig, typeName: "<FILE>"}, "kubeconfig", "Path to kubeconfig file (default: $KUBECONFIG or ~/.kube/config)")
	scanCmd.Flags().Var(&namedStringValue{ptr: &opts.kubecontext, typeName: "<KUBE-CONTEXT>"}, "context", "Kubernetes context to use (default: current context)")
	scanCmd.Flags().VarP(&namedStringValue{ptr: &opts.reportFile, typeName: "<FILE>"}, "report-file", "f", "Write scan results to FILE (format auto-detected from extension: .json, .yaml, .csv, .txt)")

	return scanCmd
}

func runScan(opts *scanOptions) error {
	ctx := context.Background()

	// Create K8s client
	var k8sOpts []k8s.Option
	if opts.kubeconfig != "" {
		k8sOpts = append(k8sOpts, k8s.WithKubeconfig(opts.kubeconfig))
	}
	if opts.kubecontext != "" {
		k8sOpts = append(k8sOpts, k8s.WithContext(opts.kubecontext))
	}

	client, err := k8s.NewClient(k8sOpts...)
	if err != nil {
		return fmt.Errorf("failed to connect to cluster: %w", err)
	}

	// Verify cluster connectivity and credentials before scanning
	if err := k8s.CheckConnection(ctx, client); err != nil {
		return err
	}

	// Load knowledge base
	kb, err := diagnosis.NewKnowledgeBase()
	if err != nil {
		return fmt.Errorf("failed to load knowledge base: %w", err)
	}

	// Create scanner registry with all scanners
	registry := scanner.DefaultRegistry()

	// Run scan
	problems, err := registry.ScanAll(ctx, client, opts.namespace, kb, opts.resources)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	// Filter by minimum severity
	if opts.minSeverity != "" {
		problems = filterBySeverity(problems, opts.minSeverity)
	}

	// Determine output format and destination
	fmtString := "table"
	var outWriter *os.File = os.Stdout

	if opts.reportFile != "" {
		fmtString = formatFromExtension(opts.reportFile)
		opts.noColor = true // force no color for file output

		file, err := os.Create(opts.reportFile)
		if err != nil {
			return fmt.Errorf("failed to create report file: %w", err)
		}
		defer func() { _ = file.Close() }()
		outWriter = file
	}

	formatter := output.NewFormatter(fmtString)
	if tf, ok := formatter.(*output.TableFormatter); ok {
		tf.NoColor = opts.noColor
	}

	err = formatter.Format(outWriter, problems, diagnosis.NewSummary(problems))
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	if opts.reportFile != "" {
		fmt.Printf("Report successfully generated: %s\n", opts.reportFile)
	}
	return nil
}

// formatFromExtension infers the output format from the report file extension.
func formatFromExtension(filename string) string {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".json"):
		return "json"
	case strings.HasSuffix(lower, ".yaml"), strings.HasSuffix(lower, ".yml"):
		return "yaml"
	case strings.HasSuffix(lower, ".csv"):
		return "csv"
	default:
		return "table"
	}
}

func filterBySeverity(problems []diagnosis.Problem, minSev string) []diagnosis.Problem {
	minWeight := parseSeverityWeight(minSev)
	var filtered []diagnosis.Problem
	for _, p := range problems {
		if p.Severity.Weight() >= minWeight {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

func parseSeverityWeight(s string) int {
	switch strings.ToLower(s) {
	case "critical":
		return 5
	case "high":
		return 4
	case "medium":
		return 3
	case "low":
		return 2
	case "info":
		return 1
	default:
		return 0
	}
}
