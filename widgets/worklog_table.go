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

// WorklogTable renders WorklogGroup data in a virtualised table.
type WorklogTable struct {
	groups binding.UntypedList // []jira.WorklogGroup
	canvas fyne.CanvasObject
}

// NewWorklogTable creates a table bound to ws.Groups.
func NewWorklogTable(ws *state.WorklogState) *WorklogTable {
	t := &WorklogTable{groups: ws.Groups}

	headers := make([]fyne.CanvasObject, len(worklogCols))
	for i, col := range worklogCols {
		headers[i] = widget.NewLabelWithStyle(col.header, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	}
	headerRow := container.NewHBox(headers...)

	table := widget.NewTable(
		func() (int, int) {
			return t.groups.Length(), len(worklogCols)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			item, err := t.groups.GetValue(id.Row)
			if err != nil {
				return
			}
			group, ok := item.(jira.WorklogGroup)
			if !ok {
				return
			}
			label := cell.(*widget.Label)
			switch id.Col {
			case 0:
				label.SetText(group.WorkReference)
			case 1:
				label.SetText(fmt.Sprintf("%.1fh", float64(group.TotalSeconds)/3600))
			case 2:
				label.SetText(fmt.Sprintf("%d", len(group.Items)))
			case 3:
				label.SetText(uniqueAuthors(group.Items))
			case 4:
				label.SetText(lastEntryDate(group.Items))
			}
		},
	)

	for i, col := range worklogCols {
		table.SetColumnWidth(i, col.width)
	}

	// Refresh table whenever the groups binding changes
	ws.Groups.AddListener(binding.NewDataListener(func() {
		table.Refresh()
	}))

	t.canvas = container.NewBorder(headerRow, nil, nil, nil, table)
	return t
}

// Canvas returns the Fyne canvas object.
func (t *WorklogTable) Canvas() fyne.CanvasObject { return t.canvas }

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
