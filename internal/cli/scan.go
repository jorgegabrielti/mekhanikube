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
	outputFmt   string
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
	return "<resource-names>"
}



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
		Example: `  nautikube scan
  nautikube scan Pod Deployment
  nautikube scan Cluster --min-severity critical
  nautikube scan -n my-namespace -o json`,
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

			// Validate output format
			validFormats := []string{"table", "csv", "json", "yaml"}
			isValidFormat := false
			for _, f := range validFormats {
				if opts.outputFmt == f {
					isValidFormat = true
					break
				}
			}
			if !isValidFormat {
				return fmt.Errorf("invalid output format: %s (allowed: %s)", opts.outputFmt, strings.Join(validFormats, ", "))
			}

			return runScan(opts)
		},
	}

	scanCmd.Flags().StringVarP(&opts.namespace, "namespace", "n", "", "Scan a specific namespace (default: all)")
	scanCmd.Flags().VarP(&opts.resources, "resource", "r", "Filter by resource type (e.g., Pod,Deployment)")
	scanCmd.Flags().StringVarP(&opts.minSeverity, "min-severity", "s", "", "Minimum severity to display (critical,high,medium,low,info)")
	scanCmd.Flags().StringVarP(&opts.outputFmt, "output", "o", "table", "Output format: table, csv, json, yaml")
	scanCmd.Flags().BoolVar(&opts.noColor, "no-color", false, "Disable colored output")
	scanCmd.Flags().StringVar(&opts.kubeconfig, "kubeconfig", "", "Path to kubeconfig file")
	scanCmd.Flags().StringVar(&opts.kubecontext, "context", "", "Kubernetes context to use")
	scanCmd.Flags().StringVarP(&opts.reportFile, "report-file", "f", "", "Write output to a file instead of stdout (format determined by --output)")

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

	// Output Formatting and Destination
	fmtString := opts.outputFmt
	var outWriter *os.File = os.Stdout

	if opts.reportFile != "" {
		if fmtString == "table" {
			opts.noColor = true // force no color for text files
		}

		file, err := os.Create(opts.reportFile)
		if err != nil {
			return fmt.Errorf("failed to create report file: %w", err)
		}
		defer file.Close()
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
