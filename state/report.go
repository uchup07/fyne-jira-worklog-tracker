// state/report.go
package state

import "fyne.io/fyne/v2/data/binding"

// ReportState holds the results and loading state for the Report screen.
type ReportState struct {
	// MappingReport holds *jira.MappingReport — stored as Untyped to avoid
	// a circular import between state and jira packages.
	MappingReport binding.Untyped
	IsLoading     binding.Bool
}

// NewReportState creates ReportState with zero values.
func NewReportState() *ReportState {
	return &ReportState{
		MappingReport: binding.NewUntyped(),
		IsLoading:     binding.NewBool(),
	}
}
