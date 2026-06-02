// state/filter.go
package state

import (
	"time"

	"fyne.io/fyne/v2/data/binding"
)

// FilterState holds all user-selected filter criteria for a worklog search.
// Fields are binding types so bound widgets auto-refresh when values change.
type FilterState struct {
	StartDate        binding.String      // "2006-01-02"
	EndDate          binding.String      // "2006-01-02"
	SelectedAuthors  binding.UntypedList // []string — Jira accountId values
	SelectedProjects binding.UntypedList // []string — Jira project keys
	SelectedTeamIDs  binding.UntypedList // []int — internal DB team IDs
	DateMode         binding.String      // "day" | "week" | "month" | "between"
}

// NewFilterState creates FilterState with sensible defaults (today's date, day mode).
func NewFilterState() *FilterState {
	today := time.Now().Format("2006-01-02")
	s := &FilterState{
		StartDate:        binding.NewString(),
		EndDate:          binding.NewString(),
		SelectedAuthors:  binding.NewUntypedList(),
		SelectedProjects: binding.NewUntypedList(),
		SelectedTeamIDs:  binding.NewUntypedList(),
		DateMode:         binding.NewString(),
	}
	_ = s.StartDate.Set(today)
	_ = s.EndDate.Set(today)
	_ = s.DateMode.Set("day")
	return s
}
