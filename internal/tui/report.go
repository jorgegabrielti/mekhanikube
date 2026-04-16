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
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

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
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

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

// pdfSanitize translates a UTF-8 string to the CP1252 encoding expected by
// fpdf's built-in fonts (Helvetica, Courier, Times). Characters outside the
// CP1252 range are replaced with their closest ASCII equivalent so that
// multi-byte glyphs like →, —, • never produce garbled output.
func pdfSanitize(tr func(string) string, s string) string {
	// Replace common symbols not in CP1252 with readable ASCII equivalents
	// before passing through the fpdf translator (which handles the rest).
	replacer := strings.NewReplacer(
		"\u2192", "->", // → rightwards arrow
		"\u2190", "<-", // ← leftwards arrow
		"\u2014", "--", // — em dash
		"\u2013", "-", // – en dash
		"\u2022", "*", // • bullet
		"\u2018", "'", // ' left single quote
		"\u2019", "'", // ' right single quote
		"\u201c", "\"", // " left double quote
		"\u201d", "\"", // " right double quote
		"\u2026", "...", // … ellipsis
	)
	return tr(replacer.Replace(s))
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
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	ps := func(s string) string { return pdfSanitize(tr, s) }

	// Title
	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(0, 10, "NautiKube - Full Report", "", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(0, 5, fmt.Sprintf("Generated: %s   |   Version: %s",
		time.Now().Format("2006-01-02 15:04:05"), version.Version), "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 5, ps(fmt.Sprintf("Context: %s   |   Namespace: %s   |   Issues: %d",
		contextName, namespace, len(problems))), "", 1, "C", false, 0, "")
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

	// Summary table rows — ISSUE column wraps via MultiCell for long text.
	const (
		colSev  = 25.0
		colRes  = 28.0
		colName = 45.0
		colScr  = 15.0
		colIss  = 67.0
		rowH    = 5.0 // base line height
	)
	pdf.SetFont("Courier", "", 8)
	for i, p := range problems {
		if i%2 == 0 {
			pdf.SetFillColor(245, 245, 245)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		fill := true
		issueText := ps(p.Issue)

		// Calculate how tall the ISSUE MultiCell will be.
		x0, y0 := pdf.GetXY()
		leftM, _, _, _ := pdf.GetMargins()
		pdf.SetLeftMargin(x0 + colSev + colRes + colName + colScr)
		pdf.SetXY(x0+colSev+colRes+colName+colScr, y0)
		// Dry-run: write to measure, then rewind.
		pdf.MultiCell(colIss, rowH, issueText, "", "L", false)
		issueH := pdf.GetY() - y0
		if issueH < rowH {
			issueH = rowH
		}
		pdf.SetLeftMargin(leftM)

		// Rewind to row start and draw fixed-width columns at the computed height.
		pdf.SetXY(x0, y0)
		pdf.CellFormat(colSev, issueH, ps(string(p.Severity)), "1", 0, "L", fill, 0, "")
		pdf.CellFormat(colRes, issueH, ps(truncate(p.Resource, 14)), "1", 0, "L", fill, 0, "")
		pdf.CellFormat(colName, issueH, ps(truncate(p.Name, 24)), "1", 0, "L", fill, 0, "")
		pdf.CellFormat(colScr, issueH, strconv.Itoa(p.Score), "1", 0, "C", fill, 0, "")

		// Draw the ISSUE cell: background rect + MultiCell text.
		issX := x0 + colSev + colRes + colName + colScr
		pdf.Rect(issX, y0, colIss, issueH, "FD")
		pdf.SetLeftMargin(issX)
		pdf.SetXY(issX, y0)
		pdf.MultiCell(colIss, rowH, issueText, "", "L", false)
		pdf.SetLeftMargin(leftM)

		// Advance to next row.
		pdf.SetXY(x0, y0+issueH)
	}
	pdf.Ln(6)

	// Detail section
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, 8, "Issue Details", "B", 1, "L", false, 0, "")
	pdf.Ln(2)

	for i, p := range problems {
		pdf.SetFont("Helvetica", "B", 10)
		pdf.SetFillColor(220, 220, 220)
		pdf.CellFormat(0, 7, ps(fmt.Sprintf("  #%d -- %s -- %s", i+1, p.Resource, truncate(p.Name, 40))), "LRB", 1, "L", true, 0, "")
		pdf.SetFillColor(255, 255, 255)
		pdfDetailField(pdf, ps, "Severity", ps(string(p.Severity)))
		pdfDetailField(pdf, ps, "Namespace", ps(p.Namespace))
		pdfDetailField(pdf, ps, "Score", strconv.Itoa(p.Score))
		pdfDetailField(pdf, ps, "Issue", ps(p.Issue))
		if p.OffendingProperty != "" {
			pdfDetailField(pdf, ps, "Offending Property", ps(p.OffendingProperty))
		}
		if p.Explanation != "" {
			pdf.SetFont("Helvetica", "B", 9)
			pdf.CellFormat(0, 5, "  Explanation:", "", 1, "L", false, 0, "")
			pdf.SetFont("Helvetica", "", 9)
			leftM, _, _, _ := pdf.GetMargins()
			pdf.SetLeftMargin(leftM + 6)
			pdf.MultiCell(0, 5, ps(p.Explanation), "", "L", false)
			pdf.SetLeftMargin(leftM)
			pdf.SetX(leftM)
		}
		if len(p.Remediation) > 0 {
			pdf.SetFont("Helvetica", "B", 9)
			pdf.CellFormat(0, 5, "  Remediation:", "", 1, "L", false, 0, "")
			pdf.SetFont("Courier", "", 8)
			leftM, _, _, _ := pdf.GetMargins()
			pdf.SetLeftMargin(leftM + 6)
			for j, cmd := range p.Remediation {
				pdf.MultiCell(0, 5, ps(fmt.Sprintf("[%d] %s", j+1, cmd)), "", "L", false)
			}
			pdf.SetLeftMargin(leftM)
			pdf.SetX(leftM)
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
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	ps := func(s string) string { return pdfSanitize(tr, s) }

	// Title
	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(0, 10, "NautiKube - Issue Report", "", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(0, 5, ps(fmt.Sprintf("Generated: %s   |   Version: %s   |   Context: %s",
		time.Now().Format("2006-01-02 15:04:05"), version.Version, contextName)), "", 1, "C", false, 0, "")
	pdf.Ln(6)

	// Detail block
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(0, 7, ps(fmt.Sprintf("  %s -- %s", p.Resource, p.Name)), "LRB", 1, "L", true, 0, "")
	pdf.SetFillColor(255, 255, 255)
	pdfDetailField(pdf, ps, "Severity", ps(string(p.Severity)))
	pdfDetailField(pdf, ps, "Namespace", ps(p.Namespace))
	pdfDetailField(pdf, ps, "Score", strconv.Itoa(p.Score))
	pdfDetailField(pdf, ps, "Issue", ps(p.Issue))
	if p.OffendingProperty != "" {
		pdfDetailField(pdf, ps, "Offending Property", ps(p.OffendingProperty))
	}
	if p.Explanation != "" {
		pdf.SetFont("Helvetica", "B", 9)
		pdf.CellFormat(0, 5, "  Explanation:", "", 1, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 9)
		leftM, _, _, _ := pdf.GetMargins()
		pdf.SetLeftMargin(leftM + 6)
		pdf.MultiCell(0, 5, ps(p.Explanation), "", "L", false)
		pdf.SetLeftMargin(leftM)
		pdf.SetX(leftM)
	}
	if len(p.Remediation) > 0 {
		pdf.SetFont("Helvetica", "B", 9)
		pdf.CellFormat(0, 5, "  Remediation:", "", 1, "L", false, 0, "")
		pdf.SetFont("Courier", "", 8)
		leftM, _, _, _ := pdf.GetMargins()
		pdf.SetLeftMargin(leftM + 6)
		for j, cmd := range p.Remediation {
			pdf.MultiCell(0, 5, ps(fmt.Sprintf("[%d] %s", j+1, cmd)), "", "L", false)
		}
		pdf.SetLeftMargin(leftM)
		pdf.SetX(leftM)
	}
	if len(p.Details) > 0 {
		pdf.SetFont("Helvetica", "B", 9)
		pdf.CellFormat(0, 5, "  Details:", "", 1, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 9)
		leftM, _, _, _ := pdf.GetMargins()
		pdf.SetLeftMargin(leftM + 6)
		for _, d := range p.Details {
			pdf.MultiCell(0, 5, ps("* "+d), "", "L", false)
		}
		pdf.SetLeftMargin(leftM)
		pdf.SetX(leftM)
	}

	return abs, pdf.OutputFileAndClose(abs)
}

// pdfDetailField writes a bold key + normal value row in a PDF detail section.
// Long values wrap within the value column using a temporary left-margin shift.
// ps is the sanitizer function that converts UTF-8 to CP1252.
func pdfDetailField(pdf *fpdf.Fpdf, ps func(string) string, key, val string) {
	const keyColW = 42.0
	leftM, _, _, _ := pdf.GetMargins()
	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(keyColW, 5, ps("  "+key+":"), "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	// Shift left margin so MultiCell wraps within the value column, not at page edge.
	pdf.SetLeftMargin(leftM + keyColW)
	pdf.MultiCell(0, 5, val, "", "L", false)
	// Restore left margin AND X cursor; fpdf leaves X at lMargin after MultiCell,
	// which at this point is leftM+keyColW, so the next element would start offset.
	pdf.SetLeftMargin(leftM)
	pdf.SetX(leftM)
}
