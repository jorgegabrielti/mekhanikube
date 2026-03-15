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

	f.renderBanner(w, summary)
	fmt.Fprintln(w)

	// Separate Info level from actual problems
	var realProblems []diagnosis.Problem
	var infoItems []diagnosis.Problem
	for _, p := range problems {
		if p.Severity == diagnosis.Info {
			infoItems = append(infoItems, p)
		} else {
			realProblems = append(realProblems, p)
		}
	}

	if len(realProblems) == 0 {
		green := color.New(color.FgGreen)
		green.Fprintln(w, "✅ Your cluster looks healthy!")
	} else {
		for _, p := range realProblems {
			f.formatProblem(w, p)
			fmt.Fprintln(w)
		}
	}

	if len(infoItems) > 0 {
		fmt.Fprintln(w, strings.Repeat("─", 40))
		fmt.Fprintln(w, "Health Report:")
		for _, item := range infoItems {
			fmt.Fprintf(w, "  %-12s %-15s %s\n", item.Name, "("+item.Resource+")", item.Issue)
		}
		fmt.Fprintln(w)
	}

	if len(realProblems) > 0 {
		f.formatSummary(w, summary)
	}
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

	// Explanation (Cause)
	// Explanation (Cause)
	if p.Explanation != "" {
		lines := strings.Split(p.Explanation, "\n")
		fmt.Fprintf(w, "   Cause: %s\n", lines[0])
		for _, line := range lines[1:] {
			if strings.TrimSpace(line) != "" {
				fmt.Fprintf(w, "          %s\n", line)
			}
		}
	}

	for _, detail := range p.Details {
		fmt.Fprintf(w, "          (%s)\n", detail)
	}

	if p.OffendingProperty != "" {
		fmt.Fprintf(w, "          > Problematic Config: %s\n", p.OffendingProperty)
	}

	// Remediation (Diagnostic Commands)
	var fixes []string
	if p.MutativeFix != "" {
		fixes = append(fixes, p.MutativeFix)
	}
	fixes = append(fixes, p.Remediation...)

	if len(fixes) > 0 {
		fmt.Fprintf(w, "   Fix:   %s\n", fixes[0])
		for _, cmd := range fixes[1:] {
			fmt.Fprintf(w, "          %s\n", cmd)
		}
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

func (f *TableFormatter) renderBanner(w io.Writer, summary diagnosis.Summary) {
	cyan := color.New(color.FgCyan, color.Bold)
	white := color.New(color.FgWhite, color.Bold)

	const width = 68
	fmt.Fprintln(w, "  ╒════════════════════════════════════════════════════════════════════╕")
	fmt.Fprintln(w, "  │                                                                    │")

	lines := []string{
		"    _   __            __   _ __  __      __        ",
		"   / | / /___ ___  __/ /_ (_) /_/ /_  __/ /_  ___  ",
		"  /  |/ / __ / / / / __// / // / / / / __ \\/ _ \\ ",
		" / /|  / /_/ / /_/ / /_ / / , < / /_/ / /_/ /  __/ ",
		"/_/ |_/\\__,_/\\__,_/\\__//_/_/|_|\\__,_/_.___/\\___/  ",
	}

	for _, line := range lines {
		padding := (width - len(line)) / 2
		if padding < 0 {
			padding = 0
		}
		fmt.Fprint(w, "  │")
		fmt.Fprint(w, strings.Repeat(" ", padding))
		cyan.Fprint(w, line)
		fmt.Fprint(w, strings.Repeat(" ", width-len(line)-padding))
		fmt.Fprintln(w, "│")
	}

	// Display Version
	versionStr := summary.Version
	verPadding := (width - len(versionStr)) / 2
	fmt.Fprint(w, "  │")
	fmt.Fprint(w, strings.Repeat(" ", verPadding))
	cyan.Fprint(w, versionStr)
	fmt.Fprint(w, strings.Repeat(" ", width-len(versionStr)-verPadding))
	fmt.Fprintln(w, "│")

	fmt.Fprintln(w, "  │                                                                    │")

	subtitle := "KUBERNETES CLUSTER HEALTH ANALYZER"
	subPadding := (width - len(subtitle)) / 2
	fmt.Fprint(w, "  │")
	fmt.Fprint(w, strings.Repeat(" ", subPadding))
	white.Fprint(w, subtitle)
	fmt.Fprint(w, strings.Repeat(" ", width-len(subtitle)-subPadding))
	fmt.Fprintln(w, "│")

	fmt.Fprintln(w, "  ╘════════════════════════════════════════════════════════════════════╛")
	fmt.Fprintf(w, "  Scan Date: %s\n", summary.ScanTime.Format("2006-01-02 15:04:05"))
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
