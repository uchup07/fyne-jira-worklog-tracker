// state/worklog.go
package state

import "fyne.io/fyne/v2/data/binding"

// WorklogState holds the results and loading state for the Dashboard screen.
type WorklogState struct {
	// Groups holds []jira.WorklogGroup — stored as UntypedList so any widget
	// can listen for changes without importing the jira package directly.
	Groups      binding.UntypedList
	SearchStart binding.String // date of the last completed search
	SearchEnd   binding.String
	ActiveTab   binding.String // "workreference" | "diagram" | "timesheet"
	IsLoading   binding.Bool
}

// NewWorklogState creates WorklogState with default tab selection.
func NewWorklogState() *WorklogState {
	s := &WorklogState{
		Groups:      binding.NewUntypedList(),
		SearchStart: binding.NewString(),
		SearchEnd:   binding.NewString(),
		ActiveTab:   binding.NewString(),
		IsLoading:   binding.NewBool(),
	}
	_ = s.ActiveTab.Set("workreference")
	return s
}
