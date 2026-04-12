package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	"github.com/jorgegabrielti/nautikube/internal/version"
)

const reportWidth = 80

var reportSep = strings.Repeat("=", reportWidth)

// reportHeader builds a centered ASCII header with the given title.
func reportHeader(title string) string {
	titleLen := len([]rune(title))
	totalPad := reportWidth - titleLen
	if totalPad < 0 {
		totalPad = 0
	}
	leftPad := totalPad / 2
	rightPad := totalPad - leftPad
	line := strings.Repeat("=", reportWidth)
	mid := strings.Repeat(" ", leftPad) + title + strings.Repeat(" ", rightPad)
	return line + "\n" + mid + "\n" + line + "\n"
}

// saveFullReport generates a full report of all problems and writes it to a file.
// Returns the absolute path of the created file.
func saveFullReport(problems []diagnosis.Problem, contextName, namespace string) (string, error) {
	if namespace == "" {
		namespace = "all namespaces"
	}
	var sb strings.Builder

	sb.WriteString(reportHeader("NautiKube - Full Report"))
	sb.WriteString(fmt.Sprintf("Generated : %s\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("Version   : %s\n", version.Version))
	sb.WriteString(fmt.Sprintf("Context   : %s\n", contextName))
	sb.WriteString(fmt.Sprintf("Namespace : %s\n", namespace))
	sb.WriteString(fmt.Sprintf("Issues    : %d\n", len(problems)))
	sb.WriteString("\n")

	if len(problems) == 0 {
		sb.WriteString("No issues found. Cluster looks healthy!\n")
	} else {
		// Summary table
		sb.WriteString(reportSep + "\n")
		sb.WriteString(fmt.Sprintf("%-10s %-15s %-25s %-5s  %s\n",
			"SEVERITY", "RESOURCE", "NAME", "SCORE", "ISSUE"))
		sb.WriteString(reportSep + "\n")
		for _, p := range problems {
			sb.WriteString(fmt.Sprintf("%-10s %-15s %-25s %-5d  %s\n",
				string(p.Severity),
				truncate(p.Resource, 15),
				truncate(p.Name, 25),
				p.Score,
				truncate(p.Issue, 50),
			))
		}
		sb.WriteString(reportSep + "\n\n")

		// Detail per issue
		for i, p := range problems {
			sb.WriteString(fmt.Sprintf("Issue #%d\n", i+1))
			sb.WriteString(reportSep + "\n")
			sb.WriteString(issueDetail(p))
			sb.WriteString("\n")
		}
	}

	return writeReport("nautikube-report", sb.String())
}

// saveIssueReport generates a report for a single problem and writes it to a file.
// Returns the absolute path of the created file.
func saveIssueReport(p diagnosis.Problem, contextName string) (string, error) {
	var sb strings.Builder

	sb.WriteString(reportHeader("NautiKube - Issue Report"))
	sb.WriteString(fmt.Sprintf("Generated : %s\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("Version   : %s\n", version.Version))
	sb.WriteString(fmt.Sprintf("Context   : %s\n", contextName))
	sb.WriteString("\n")
	sb.WriteString(reportSep + "\n")
	sb.WriteString(issueDetail(p))

	return writeReport("nautikube-issue", sb.String())
}

// issueDetail returns a plain-text block describing a single problem.
func issueDetail(p diagnosis.Problem) string {
	var sb strings.Builder
	field := func(key, val string) {
		sb.WriteString(fmt.Sprintf("  %-20s %s\n", key+":", val))
	}
	field("Severity", string(p.Severity))
	field("Resource", p.Resource)
	field("Namespace", p.Namespace)
	field("Name", p.Name)
	field("Score", fmt.Sprintf("%d", p.Score))
	field("Issue", p.Issue)
	if p.OffendingProperty != "" {
		field("Offending Property", p.OffendingProperty)
	}
	if p.Explanation != "" {
		sb.WriteString(fmt.Sprintf("  %-20s\n", "Explanation:"))
		for _, line := range wrapText(p.Explanation, 76) {
			sb.WriteString("    " + line + "\n")
		}
	}
	if len(p.Remediation) > 0 {
		sb.WriteString(fmt.Sprintf("  %-20s\n", "Remediation:"))
		for i, cmd := range p.Remediation {
			sb.WriteString(fmt.Sprintf("    [%d] %s\n", i+1, cmd))
		}
	}
	if p.MutativeFix != "" {
		sb.WriteString(fmt.Sprintf("  %-20s\n", "Mutative Fix:"))
		sb.WriteString("    " + p.MutativeFix + "\n")
	}
	if len(p.Details) > 0 {
		sb.WriteString(fmt.Sprintf("  %-20s\n", "Details:"))
		for _, d := range p.Details {
			sb.WriteString("    • " + d + "\n")
		}
	}
	return sb.String()
}

// writeReport saves content to a timestamped .txt file in the current directory.
func writeReport(prefix, content string) (string, error) {
	ts := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s-%s.txt", prefix, ts)
	abs, err := filepath.Abs(filename)
	if err != nil {
		abs = filename
	}
	if err := os.WriteFile(abs, []byte(content), 0o600); err != nil {
		return "", err
	}
	return abs, nil
}
