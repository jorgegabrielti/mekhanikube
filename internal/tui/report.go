package tui

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
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

// reportFilePath computes an absolute timestamped file path for the given prefix and extension.
func reportFilePath(prefix, ext string) (string, error) {
	ts := time.Now().Format("20060102-150405")
	return filepath.Abs(fmt.Sprintf("%s-%s.%s", prefix, ts, ext))
}

// writeReport saves content to a timestamped .txt file in the current directory.
func writeReport(prefix, content string) (string, error) {
	abs, err := reportFilePath(prefix, "txt")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(abs, []byte(content), 0o600); err != nil {
		return "", err
	}
	return abs, nil
}

// saveCSVReport writes all problems to a timestamped CSV file.
func saveCSVReport(problems []diagnosis.Problem, contextName, namespace string) (string, error) {
	abs, err := reportFilePath("nautikube-report", "csv")
	if err != nil {
		return "", err
	}
	f, err := os.OpenFile(abs, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return "", err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	_ = w.Write([]string{"Severity", "Resource", "Namespace", "Name", "Score", "Issue", "OffendingProperty", "Explanation", "Remediation"})
	for _, p := range problems {
		_ = w.Write([]string{
			string(p.Severity),
			p.Resource,
			p.Namespace,
			p.Name,
			strconv.Itoa(p.Score),
			p.Issue,
			p.OffendingProperty,
			p.Explanation,
			strings.Join(p.Remediation, " | "),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return "", err
	}
	return abs, nil
}

// saveCSVIssueReport writes a single problem to a timestamped CSV file.
func saveCSVIssueReport(p diagnosis.Problem, contextName string) (string, error) {
	abs, err := reportFilePath("nautikube-issue", "csv")
	if err != nil {
		return "", err
	}
	f, err := os.OpenFile(abs, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return "", err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	_ = w.Write([]string{"Severity", "Resource", "Namespace", "Name", "Score", "Issue", "OffendingProperty", "Explanation", "Remediation"})
	_ = w.Write([]string{
		string(p.Severity),
		p.Resource,
		p.Namespace,
		p.Name,
		strconv.Itoa(p.Score),
		p.Issue,
		p.OffendingProperty,
		p.Explanation,
		strings.Join(p.Remediation, " | "),
	})
	w.Flush()
	return abs, w.Error()
}

// savePDFReport generates a PDF containing all problems and saves it to disk.
func savePDFReport(problems []diagnosis.Problem, contextName, namespace string) (string, error) {
	if namespace == "" {
		namespace = "all namespaces"
	}
	abs, err := reportFilePath("nautikube-report", "pdf")
	if err != nil {
		return "", err
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 15)
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Title
	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(0, 10, "NautiKube - Full Report", "", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(0, 5, fmt.Sprintf("Generated: %s   |   Version: %s",
		time.Now().Format("2006-01-02 15:04:05"), version.Version), "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 5, fmt.Sprintf("Context: %s   |   Namespace: %s   |   Issues: %d",
		contextName, namespace, len(problems)), "", 1, "C", false, 0, "")
	pdf.Ln(4)

	if len(problems) == 0 {
		pdf.SetFont("Helvetica", "", 11)
		pdf.CellFormat(0, 10, "No issues found. Cluster looks healthy!", "", 1, "C", false, 0, "")
		return abs, pdf.OutputFileAndClose(abs)
	}

	// Summary table header
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetFillColor(40, 40, 40)
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(25, 7, "SEVERITY", "1", 0, "C", true, 0, "")
	pdf.CellFormat(28, 7, "RESOURCE", "1", 0, "C", true, 0, "")
	pdf.CellFormat(45, 7, "NAME", "1", 0, "C", true, 0, "")
	pdf.CellFormat(15, 7, "SCORE", "1", 0, "C", true, 0, "")
	pdf.CellFormat(67, 7, "ISSUE", "1", 1, "C", true, 0, "")
	pdf.SetTextColor(0, 0, 0)

	// Summary table rows
	pdf.SetFont("Courier", "", 8)
	for i, p := range problems {
		if i%2 == 0 {
			pdf.SetFillColor(245, 245, 245)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		pdf.CellFormat(25, 6, string(p.Severity), "1", 0, "L", true, 0, "")
		pdf.CellFormat(28, 6, truncate(p.Resource, 17), "1", 0, "L", true, 0, "")
		pdf.CellFormat(45, 6, truncate(p.Name, 28), "1", 0, "L", true, 0, "")
		pdf.CellFormat(15, 6, strconv.Itoa(p.Score), "1", 0, "C", true, 0, "")
		pdf.CellFormat(67, 6, truncate(p.Issue, 42), "1", 1, "L", true, 0, "")
	}
	pdf.Ln(6)

	// Detail section
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, 8, "Issue Details", "B", 1, "L", false, 0, "")
	pdf.Ln(2)

	for i, p := range problems {
		pdf.SetFont("Helvetica", "B", 10)
		pdf.SetFillColor(220, 220, 220)
		pdf.CellFormat(0, 7, fmt.Sprintf("  #%d — %s — %s", i+1, p.Resource, truncate(p.Name, 40)), "LRB", 1, "L", true, 0, "")
		pdf.SetFillColor(255, 255, 255)
		pdfDetailField(pdf, "Severity", string(p.Severity))
		pdfDetailField(pdf, "Namespace", p.Namespace)
		pdfDetailField(pdf, "Score", strconv.Itoa(p.Score))
		pdfDetailField(pdf, "Issue", p.Issue)
		if p.OffendingProperty != "" {
			pdfDetailField(pdf, "Offending Property", p.OffendingProperty)
		}
		if p.Explanation != "" {
			pdf.SetFont("Helvetica", "B", 9)
			pdf.CellFormat(0, 5, "  Explanation:", "", 1, "L", false, 0, "")
			pdf.SetFont("Helvetica", "", 9)
			pdf.MultiCell(0, 5, "    "+p.Explanation, "", "L", false)
		}
		if len(p.Remediation) > 0 {
			pdf.SetFont("Helvetica", "B", 9)
			pdf.CellFormat(0, 5, "  Remediation:", "", 1, "L", false, 0, "")
			pdf.SetFont("Courier", "", 8)
			for j, cmd := range p.Remediation {
				pdf.MultiCell(0, 5, fmt.Sprintf("    [%d] %s", j+1, cmd), "", "L", false)
			}
		}
		pdf.Ln(3)
	}

	return abs, pdf.OutputFileAndClose(abs)
}

// savePDFIssueReport generates a single-issue PDF and saves it to disk.
func savePDFIssueReport(p diagnosis.Problem, contextName string) (string, error) {
	abs, err := reportFilePath("nautikube-issue", "pdf")
	if err != nil {
		return "", err
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 15)
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Title
	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(0, 10, "NautiKube - Issue Report", "", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(0, 5, fmt.Sprintf("Generated: %s   |   Version: %s   |   Context: %s",
		time.Now().Format("2006-01-02 15:04:05"), version.Version, contextName), "", 1, "C", false, 0, "")
	pdf.Ln(6)

	// Detail block
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(0, 7, fmt.Sprintf("  %s — %s", p.Resource, p.Name), "LRB", 1, "L", true, 0, "")
	pdf.SetFillColor(255, 255, 255)
	pdfDetailField(pdf, "Severity", string(p.Severity))
	pdfDetailField(pdf, "Namespace", p.Namespace)
	pdfDetailField(pdf, "Score", strconv.Itoa(p.Score))
	pdfDetailField(pdf, "Issue", p.Issue)
	if p.OffendingProperty != "" {
		pdfDetailField(pdf, "Offending Property", p.OffendingProperty)
	}
	if p.Explanation != "" {
		pdf.SetFont("Helvetica", "B", 9)
		pdf.CellFormat(0, 5, "  Explanation:", "", 1, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 9)
		pdf.MultiCell(0, 5, "    "+p.Explanation, "", "L", false)
	}
	if len(p.Remediation) > 0 {
		pdf.SetFont("Helvetica", "B", 9)
		pdf.CellFormat(0, 5, "  Remediation:", "", 1, "L", false, 0, "")
		pdf.SetFont("Courier", "", 8)
		for j, cmd := range p.Remediation {
			pdf.MultiCell(0, 5, fmt.Sprintf("    [%d] %s", j+1, cmd), "", "L", false)
		}
	}
	if p.MutativeFix != "" {
		pdf.SetFont("Helvetica", "B", 9)
		pdf.CellFormat(0, 5, "  Mutative Fix:", "", 1, "L", false, 0, "")
		pdf.SetFont("Courier", "", 8)
		pdf.MultiCell(0, 5, "    "+p.MutativeFix, "", "L", false)
	}
	if len(p.Details) > 0 {
		pdf.SetFont("Helvetica", "B", 9)
		pdf.CellFormat(0, 5, "  Details:", "", 1, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 9)
		for _, d := range p.Details {
			pdf.MultiCell(0, 5, "    • "+d, "", "L", false)
		}
	}

	return abs, pdf.OutputFileAndClose(abs)
}

// pdfDetailField writes a bold key + normal value row in a PDF detail section.
func pdfDetailField(pdf *fpdf.Fpdf, key, val string) {
	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(42, 5, "  "+key+":", "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(138, 5, val, "", 1, "L", false, 0, "")
}
