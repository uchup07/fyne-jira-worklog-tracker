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

const mappingPageSize = 20

var mappingCols = []struct {
	header string
	width  float32
}{
	{"IRQ Key", 120},
	{"JSW Key", 120},
	{"Task Name", 180},
	{"Vertical", 140},
	{"Company", 140},
	{"Hours", 80},
}

// MappingTable renders MappingReport rows in a virtualised table with pagination.
type MappingTable struct {
	rs          *state.ReportState
	allRows     []jira.MappingRow // full dataset, updated on every search
	visibleRows []jira.MappingRow // slice of allRows for the current page
	table       *widget.Table
	pageLabel   *widget.Label
	prevBtn     *widget.Button
	nextBtn     *widget.Button
	page        int
	canvas      fyne.CanvasObject
}

// NewMappingTable creates a table bound to rs.MappingReport.
func NewMappingTable(rs *state.ReportState) *MappingTable {
	t := &MappingTable{rs: rs}

	t.table = widget.NewTable(
		func() (int, int) { return len(t.visibleRows), len(mappingCols) },
		func() fyne.CanvasObject {
			lbl := widget.NewLabel("")
			lbl.Wrapping = fyne.TextWrapWord
			return lbl
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			if id.Row >= len(t.visibleRows) {
				return
			}
			row := t.visibleRows[id.Row]
			lbl := cell.(*widget.Label)
			switch id.Col {
			case 0:
				lbl.Alignment = fyne.TextAlignLeading
				lbl.SetText(row.IRQKey)
			case 1:
				lbl.Alignment = fyne.TextAlignLeading
				lbl.SetText(row.JSWKey)
			case 2:
				lbl.Alignment = fyne.TextAlignLeading
				lbl.SetText(row.TaskName)
			case 3:
				lbl.Alignment = fyne.TextAlignLeading
				lbl.SetText(row.Vertical)
			case 4:
				lbl.Alignment = fyne.TextAlignLeading
				lbl.SetText(row.Company)
			case 5:
				lbl.Alignment = fyne.TextAlignTrailing
				lbl.SetText(fmt.Sprintf("%.1fh", float64(row.TotalSeconds)/3600))
			}
		},
	)

	// Native sticky header row (Fyne 2.4+).
	t.table.ShowHeaderRow = true
	t.table.CreateHeader = func() fyne.CanvasObject {
		return widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	}
	t.table.UpdateHeader = func(id widget.TableCellID, cell fyne.CanvasObject) {
		if id.Col >= 0 && id.Col < len(mappingCols) {
			cell.(*widget.Label).SetText(mappingCols[id.Col].header)
		}
	}

	for i, col := range mappingCols {
		t.table.SetColumnWidth(i, col.width)
	}

	for i := 0; i < mappingPageSize; i++ {
		t.table.SetRowHeight(i, worklogRowHeight)
	}

	// Pagination controls.
	t.pageLabel = widget.NewLabel("")
	t.pageLabel.Alignment = fyne.TextAlignCenter
	t.prevBtn = widget.NewButton("< Prev", func() {
		if t.page > 0 {
			t.goToPage(t.page - 1)
		}
	})
	t.nextBtn = widget.NewButton("Next >", func() {
		if t.page < t.pageCount()-1 {
			t.goToPage(t.page + 1)
		}
	})
	t.updatePaginationM()

	// MappingReport.Set is called inside fyne.Do, so the listener fires on the main goroutine.
	rs.MappingReport.AddListener(binding.NewDataListener(func() {
		val, err := rs.MappingReport.Get()
		if err != nil || val == nil {
			t.allRows = nil
			t.goToPage(0)
			return
		}
		report, ok := val.(*jira.MappingReport)
		if !ok {
			return
		}
		t.allRows = report.Rows
		t.goToPage(0)
	}))

	pagination := container.NewPadded(
		container.NewBorder(nil, nil, t.prevBtn, t.nextBtn, t.pageLabel),
	)
	t.canvas = container.NewBorder(nil, pagination, nil, nil, t.table)
	return t
}

// Canvas returns the Fyne canvas object.
func (t *MappingTable) Canvas() fyne.CanvasObject { return t.canvas }

func (t *MappingTable) pageCount() int {
	n := len(t.allRows)
	if n == 0 {
		return 1
	}
	pages := n / mappingPageSize
	if n%mappingPageSize != 0 {
		pages++
	}
	return pages
}

func (t *MappingTable) goToPage(p int) {
	t.page = p
	start := p * mappingPageSize
	end := start + mappingPageSize
	if end > len(t.allRows) {
		end = len(t.allRows)
	}
	if start > len(t.allRows) {
		start = len(t.allRows)
	}
	t.visibleRows = t.allRows[start:end]
	t.table.ScrollTo(widget.TableCellID{Row: 0, Col: 0})
	t.table.Refresh()
	t.updatePaginationM()
}

func (t *MappingTable) updatePaginationM() {
	pages := t.pageCount()
	t.pageLabel.SetText(fmt.Sprintf("Page %d of %d (%d items)", t.page+1, pages, len(t.allRows)))
	if t.page > 0 {
		t.prevBtn.Enable()
	} else {
		t.prevBtn.Disable()
	}
	if t.page < pages-1 {
		t.nextBtn.Enable()
	} else {
		t.nextBtn.Disable()
	}
}
