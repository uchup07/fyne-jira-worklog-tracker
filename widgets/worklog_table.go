// widgets/worklog_table.go
package widgets

import (
	"fmt"
	"time"

	"github.com/uchup07/fyne-jira-worklog-tracker/jira"
	"github.com/uchup07/fyne-jira-worklog-tracker/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

const (
	worklogPageSize  = 20
	worklogRowHeight = float32(60)
)

var worklogCols = []struct {
	header string
	width  float32
}{
	{"Work Reference", 180},
	{"Total Hours", 110},
	{"Entries", 80},
	{"Authors", 200},
	{"Last Entry", 120},
}

// WorklogTable renders WorklogGroup data in a virtualised table with pagination.
type WorklogTable struct {
	groups      binding.UntypedList
	allRows     []jira.WorklogGroup // full dataset, updated on every search
	visibleRows []jira.WorklogGroup // slice of allRows for the current page
	table       *widget.Table
	pageLabel   *widget.Label
	prevBtn     *widget.Button
	nextBtn     *widget.Button
	page        int
	canvas      fyne.CanvasObject
}

// NewWorklogTable creates a table bound to ws.Groups.
func NewWorklogTable(ws *state.WorklogState) *WorklogTable {
	t := &WorklogTable{groups: ws.Groups}

	t.table = widget.NewTable(
		func() (int, int) { return len(t.visibleRows), len(worklogCols) },
		func() fyne.CanvasObject {
			lbl := widget.NewLabel("")
			lbl.Wrapping = fyne.TextWrapWord
			return lbl
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			if id.Row >= len(t.visibleRows) {
				return
			}
			group := t.visibleRows[id.Row]
			lbl := cell.(*widget.Label)
			switch id.Col {
			case 0:
				lbl.Alignment = fyne.TextAlignLeading
				lbl.SetText(group.WorkReference)
			case 1:
				lbl.Alignment = fyne.TextAlignTrailing
				lbl.SetText(fmt.Sprintf("%.1fh", float64(group.TotalSeconds)/3600))
			case 2:
				lbl.Alignment = fyne.TextAlignTrailing
				lbl.SetText(fmt.Sprintf("%d", len(group.Items)))
			case 3:
				lbl.Alignment = fyne.TextAlignLeading
				lbl.SetText(uniqueAuthors(group.Items))
			case 4:
				lbl.Alignment = fyne.TextAlignLeading
				lbl.SetText(lastEntryDate(group.Items))
			}
		},
	)

	// Native sticky header row (Fyne 2.4+).
	t.table.ShowHeaderRow = true
	t.table.CreateHeader = func() fyne.CanvasObject {
		return widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	}
	t.table.UpdateHeader = func(id widget.TableCellID, cell fyne.CanvasObject) {
		if id.Col >= 0 && id.Col < len(worklogCols) {
			cell.(*widget.Label).SetText(worklogCols[id.Col].header)
		}
	}

	for i, col := range worklogCols {
		t.table.SetColumnWidth(i, col.width)
	}

	// Pre-set a fixed row height for all rows in one page so wrapped text has
	// enough vertical space. These heights persist across Refresh() calls.
	for i := 0; i < worklogPageSize; i++ {
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
	t.updatePagination()

	// Groups.Set is always called inside fyne.Do, so the listener fires on the main goroutine.
	ws.Groups.AddListener(binding.NewDataListener(func() {
		n := ws.Groups.Length()
		rows := make([]jira.WorklogGroup, 0, n)
		for i := 0; i < n; i++ {
			val, err := ws.Groups.GetValue(i)
			if err == nil {
				if g, ok := val.(jira.WorklogGroup); ok {
					rows = append(rows, g)
				}
			}
		}
		t.allRows = rows
		t.goToPage(0)
	}))

	pagination := container.NewPadded(
		container.NewBorder(nil, nil, t.prevBtn, t.nextBtn, t.pageLabel),
	)
	t.canvas = container.NewBorder(nil, pagination, nil, nil, t.table)
	return t
}

// Canvas returns the Fyne canvas object.
func (t *WorklogTable) Canvas() fyne.CanvasObject { return t.canvas }

func (t *WorklogTable) pageCount() int {
	n := len(t.allRows)
	if n == 0 {
		return 1
	}
	pages := n / worklogPageSize
	if n%worklogPageSize != 0 {
		pages++
	}
	return pages
}

func (t *WorklogTable) goToPage(p int) {
	t.page = p
	start := p * worklogPageSize
	end := start + worklogPageSize
	if end > len(t.allRows) {
		end = len(t.allRows)
	}
	if start > len(t.allRows) {
		start = len(t.allRows)
	}
	t.visibleRows = t.allRows[start:end]
	t.table.ScrollTo(widget.TableCellID{Row: 0, Col: 0})
	t.table.Refresh()
	t.updatePagination()
}

func (t *WorklogTable) updatePagination() {
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

// uniqueAuthors returns a comma-separated list of unique author display names (max 3 shown).
func uniqueAuthors(items []jira.WorklogItem) string {
	seen := map[string]bool{}
	var names []string
	for _, item := range items {
		if !seen[item.Author.DisplayName] {
			seen[item.Author.DisplayName] = true
			names = append(names, item.Author.DisplayName)
		}
	}
	if len(names) > 3 {
		return fmt.Sprintf("%s +%d", names[0], len(names)-1)
	}
	result := ""
	for i, n := range names {
		if i > 0 {
			result += ", "
		}
		result += n
	}
	return result
}

// lastEntryDate returns the most recent Started date across all items.
func lastEntryDate(items []jira.WorklogItem) string {
	var latest time.Time
	for _, item := range items {
		if item.Started.After(latest) {
			latest = item.Started
		}
	}
	if latest.IsZero() {
		return ""
	}
	return latest.Format("2006-01-02")
}
