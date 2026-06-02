// export/pdf.go
package export

import (
	"fmt"

	"github.com/jung-kurt/gofpdf"
	"github.com/uchup07/fyne-jira-worklog-tracker/jira"
)

// WritePDF writes a worklog summary report to a PDF file at path.
func WritePDF(path string, groups []jira.WorklogGroup, startDate, endDate string) error {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Jira Worklog Report")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 8, fmt.Sprintf("Period: %s to %s", startDate, endDate))
	pdf.Ln(12)

	// Table header
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(37, 99, 235)
	pdf.SetTextColor(255, 255, 255)
	colWidths := []float64{50, 30, 60, 40, 25, 30, 55}
	headers := []string{"Work Ref", "Issue", "Summary", "Author", "Hours", "Date", "Comment"}
	for i, h := range headers {
		pdf.CellFormat(colWidths[i], 8, h, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// Table rows
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(0, 0, 0)
	fill := false
	for _, group := range groups {
		for _, item := range group.Items {
			if fill {
				pdf.SetFillColor(235, 241, 255)
			} else {
				pdf.SetFillColor(255, 255, 255)
			}
			values := []string{
				group.WorkReference,
				item.IssueKey,
				truncate(item.IssueSummary, 28),
				item.Author.DisplayName,
				fmt.Sprintf("%.1f", float64(item.TimeSpentSeconds)/3600),
				item.Started.Format("2006-01-02"),
				truncate(item.Comment, 28),
			}
			for i, v := range values {
				pdf.CellFormat(colWidths[i], 7, v, "1", 0, "L", fill, 0, "")
			}
			pdf.Ln(-1)
			fill = !fill
		}
	}

	// Summary footer
	pdf.Ln(4)
	pdf.SetFont("Arial", "B", 10)
	totalHours := 0.0
	for _, g := range groups {
		totalHours += float64(g.TotalSeconds) / 3600
	}
	pdf.Cell(0, 8, fmt.Sprintf("Total: %.1f hours across %d work references", totalHours, len(groups)))

	return pdf.OutputFileAndClose(path)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
