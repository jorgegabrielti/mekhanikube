package output

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
)

// CSVFormatter writes scan results in CSV format.
type CSVFormatter struct{}

// Format writes the scan results to the provided writer in CSV format.
func (f *CSVFormatter) Format(w io.Writer, problems []diagnosis.Problem, summary diagnosis.Summary) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write Header
	header := []string{"Timestamp", "Version", "Resource", "Namespace", "Name", "Severity", "Score", "Issue", "Cause", "OffendingProperty", "MutativeFix", "Remediation"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	timestamp := summary.ScanTime.Format(time.RFC3339)
	version := summary.Version

	// Write Data Rows
	for _, p := range problems {
		explanation := p.Explanation
		if len(p.Details) > 0 {
			explanation += "\n" + strings.Join(p.Details, "\n")
		}

		// Join remediation commands with a newline or semicolon so it fits in a single cell
		remediation := strings.Join(p.Remediation, " ; ")

		row := []string{
			timestamp,
			version,
			p.Resource,
			p.Namespace,
			p.Name,
			string(p.Severity),
			fmt.Sprintf("%d", p.Score),
			p.Issue,
			explanation,
			p.OffendingProperty,
			p.MutativeFix,
			remediation,
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}
