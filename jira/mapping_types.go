// jira/mapping_types.go
package jira

import "time"

// MappingRow is one row in the mapping report — one JSW task linked to one IRQ.
type MappingRow struct {
	JSWKey        string
	JSWSummary    string
	JSWProjectKey string
	TaskName      string // parsed category from issue summary
	IRQKey        string
	IRQSummary    string
	IRQStatus     string
	Vertical      string
	Company       string
	Application   string
	Module        string
	Division      string
	UATEndDate    string
	TotalSeconds  int
	WorklogMonths map[string]int // "2026-01" -> seconds
}

// MappingReport is the complete output of the mapping report pipeline.
type MappingReport struct {
	Rows            []MappingRow
	TotalCRSeconds  int
	UniqueIRQCount  int
	UniqueVerticals int
	UniqueCompanies int
	GeneratedAt     time.Time
}

// IrqGroup aggregates MappingRows by IRQ key — used by the collapsible report view.
type IrqGroup struct {
	IRQKey       string
	IRQSummary   string
	IRQStatus    string
	Company      string
	Application  string
	Module       string
	Division     string
	Vertical     string
	UATEndDate   string
	JSWTasks     []IrqJSWTask
	TotalSeconds int
}

// IrqJSWTask is a de-duplicated JSW task within an IrqGroup.
type IrqJSWTask struct {
	Key          string
	Summary      string
	ProjectKey   string
	TaskName     string
	TotalSeconds int
}

// ChartData holds labels + values for pie and bar charts.
type ChartData struct {
	Labels []string
	Values []float64
}
