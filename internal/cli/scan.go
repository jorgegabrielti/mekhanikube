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
	resources   []string
	minSeverity string
	outputFmt   string
	noColor     bool
	kubeconfig  string
	kubecontext string
}

func init() {
	opts := &scanOptions{}

	scanCmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan the cluster for problems",
		Long: `Scan your Kubernetes cluster for common problems such as crashing pods,
unavailable deployments, services without endpoints, and unhealthy nodes.

Results are sorted by severity score (highest first) and include
actionable remediation commands.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScan(cmd, args, opts)
		},
	}

	scanCmd.Flags().StringVarP(&opts.namespace, "namespace", "n", "", "Scan a specific namespace (default: all)")
	scanCmd.Flags().StringSliceVarP(&opts.resources, "resource", "r", nil, "Filter by resource type (e.g., Pod,Deployment)")
	scanCmd.Flags().StringVarP(&opts.minSeverity, "min-severity", "s", "", "Minimum severity to display (critical,high,medium,low,info)")
	scanCmd.Flags().StringVarP(&opts.outputFmt, "output", "o", "table", "Output format: table, json, yaml")
	scanCmd.Flags().BoolVar(&opts.noColor, "no-color", false, "Disable colored output")
	scanCmd.Flags().StringVar(&opts.kubeconfig, "kubeconfig", "", "Path to kubeconfig file")
	scanCmd.Flags().StringVar(&opts.kubecontext, "context", "", "Kubernetes context to use")

	rootCmd.AddCommand(scanCmd)
}

func runScan(cmd *cobra.Command, args []string, opts *scanOptions) error {
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

	// Format and output
	summary := diagnosis.NewSummary(problems)
	formatter := output.NewFormatter(opts.outputFmt)
	if tf, ok := formatter.(*output.TableFormatter); ok {
		tf.NoColor = opts.noColor
	}

	return formatter.Format(os.Stdout, problems, summary)
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
