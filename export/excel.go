// export/excel.go
package export

import (
	"fmt"

	"github.com/uchup07/fyne-jira-worklog-tracker/jira"
	"github.com/xuri/excelize/v2"
)

// WriteExcel writes worklog groups to an Excel (.xlsx) file at path.
func WriteExcel(path string, groups []jira.WorklogGroup) error {
	f := excelize.NewFile()
	sheet := "Worklogs"
	f.SetSheetName("Sheet1", sheet)

	bold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	headers := []string{"Work Reference", "Issue Key", "Summary", "Author", "Hours", "Date", "Comment"}
	for col, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, bold)
	}

	row := 2
	for _, group := range groups {
		for _, item := range group.Items {
			values := []any{
				group.WorkReference,
				item.IssueKey,
				item.IssueSummary,
				item.Author.DisplayName,
				fmt.Sprintf("%.2f", float64(item.TimeSpentSeconds)/3600),
				item.Started.Format("2006-01-02"),
				item.Comment,
			}
			for col, v := range values {
				cell, _ := excelize.CoordinatesToCellName(col+1, row)
				f.SetCellValue(sheet, cell, v)
			}
			row++
		}
	}

	for col := 1; col <= 7; col++ {
		name, _ := excelize.ColumnNumberToName(col)
		f.SetColWidth(sheet, name, name, 18)
	}

	return f.SaveAs(path)
}
