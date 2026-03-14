package output

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
)

// TableFormatter writes scan results as a colorized terminal table.
type TableFormatter struct {
	NoColor bool
}

// Format writes the scan results in human-readable table format.
func (f *TableFormatter) Format(w io.Writer, problems []diagnosis.Problem, summary diagnosis.Summary) error {
	if f.NoColor || os.Getenv("NO_COLOR") != "" {
		color.NoColor = true
	}

	bold := color.New(color.Bold)
	bold.Fprintln(w, "NautiKube Scan Results")
	bold.Fprintln(w, "======================")
	fmt.Fprintln(w)

	if len(problems) == 0 {
		green := color.New(color.FgGreen)
		green.Fprintln(w, "✅ No problems found. Your cluster looks healthy!")
		return nil
	}

	for _, p := range problems {
		f.formatProblem(w, p)
		fmt.Fprintln(w)
	}

	f.formatSummary(w, summary)
	return nil
}

func (f *TableFormatter) formatProblem(w io.Writer, p diagnosis.Problem) {
	icon := severityIcon(p.Severity)
	c := severityColor(p.Severity)

	// Header line: icon + severity + resource + namespace/name
	c.Fprintf(w, "%s %-8s  %s  %s/%s\n", icon, p.Severity, p.Resource, p.Namespace, p.Name)

	// Score
	fmt.Fprintf(w, "   Score: %d/100\n", p.Score)

	// Issue
	fmt.Fprintf(w, "   Issue: %s\n", p.Issue)

	// Remediation
	if len(p.Remediation) > 0 {
		fmt.Fprintf(w, "   Fix:   %s\n", p.Remediation[0])
		for _, cmd := range p.Remediation[1:] {
			fmt.Fprintf(w, "          %s\n", cmd)
		}
	}

	// Details
	for _, detail := range p.Details {
		fmt.Fprintf(w, "   • %s\n", detail)
	}
}

func (f *TableFormatter) formatSummary(w io.Writer, s diagnosis.Summary) {
	fmt.Fprintln(w, strings.Repeat("─", 40))

	parts := []string{
		fmt.Sprintf("Summary: %d problems found", s.Total),
	}
	fmt.Fprintln(w, parts[0])

	var counts []string
	if s.Critical > 0 {
		counts = append(counts, color.RedString("🔴 Critical: %d", s.Critical))
	}
	if s.High > 0 {
		counts = append(counts, color.YellowString("🟠 High: %d", s.High))
	}
	if s.Medium > 0 {
		counts = append(counts, color.YellowString("🟡 Medium: %d", s.Medium))
	}
	if s.Low > 0 {
		counts = append(counts, color.BlueString("🔵 Low: %d", s.Low))
	}
	if s.Info > 0 {
		counts = append(counts, fmt.Sprintf("⚪ Info: %d", s.Info))
	}

	if len(counts) > 0 {
		fmt.Fprintf(w, "  %s\n", strings.Join(counts, "  "))
	}
}

func severityIcon(s diagnosis.Severity) string {
	switch s {
	case diagnosis.Critical:
		return "🔴"
	case diagnosis.High:
		return "🟠"
	case diagnosis.Medium:
		return "🟡"
	case diagnosis.Low:
		return "🔵"
	case diagnosis.Info:
		return "⚪"
	default:
		return "❓"
	}
}

func severityColor(s diagnosis.Severity) *color.Color {
	switch s {
	case diagnosis.Critical:
		return color.New(color.FgRed, color.Bold)
	case diagnosis.High:
		return color.New(color.FgYellow)
	case diagnosis.Medium:
		return color.New(color.FgYellow)
	case diagnosis.Low:
		return color.New(color.FgBlue)
	default:
		return color.New(color.FgWhite)
	}
}
