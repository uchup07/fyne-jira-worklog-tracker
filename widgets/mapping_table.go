// widgets/mapping_table.go
package widgets

import (
	"fmt"

	"github.com/uchup07/fyne-jira-worklog-tracker/jira"
	"github.com/uchup07/fyne-jira-worklog-tracker/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

var mappingCols = []string{"IRQ Key", "JSW Key", "Task Name", "Vertical", "Company", "Hours"}

// MappingTable renders MappingReport rows in a virtualised table.
type MappingTable struct {
	rs     *state.ReportState
	canvas fyne.CanvasObject
}

// NewMappingTable creates a table bound to rs.MappingReport.
func NewMappingTable(rs *state.ReportState) *MappingTable {
	t := &MappingTable{rs: rs}

	getRows := func() []jira.MappingRow {
		val, err := rs.MappingReport.Get()
		if err != nil || val == nil {
			return nil
		}
		report, ok := val.(*jira.MappingReport)
		if !ok {
			return nil
		}
		return report.Rows
	}

	headers := make([]fyne.CanvasObject, len(mappingCols))
	for i, col := range mappingCols {
		headers[i] = widget.NewLabelWithStyle(col, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	}

	table := widget.NewTable(
		func() (int, int) { return len(getRows()), len(mappingCols) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			rows := getRows()
			if id.Row >= len(rows) {
				return
			}
			row := rows[id.Row]
			label := cell.(*widget.Label)
			switch id.Col {
			case 0:
				label.SetText(row.IRQKey)
			case 1:
				label.SetText(row.JSWKey)
			case 2:
				label.SetText(row.TaskName)
			case 3:
				label.SetText(row.Vertical)
			case 4:
				label.SetText(row.Company)
			case 5:
				label.SetText(fmt.Sprintf("%.1fh", float64(row.TotalSeconds)/3600))
			}
		},
	)

	colWidths := []float32{100, 100, 120, 120, 120, 80}
	for i, w := range colWidths {
		table.SetColumnWidth(i, w)
	}

	rs.MappingReport.AddListener(binding.NewDataListener(func() {
		table.Refresh()
	}))

	headerRow := container.NewHBox(headers...)
	t.canvas = container.NewBorder(headerRow, nil, nil, nil, table)
	return t
}

// Canvas returns the Fyne canvas object.
func (t *MappingTable) Canvas() fyne.CanvasObject { return t.canvas }
