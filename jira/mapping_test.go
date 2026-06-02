// jira/mapping_test.go
package jira_test

import (
	"testing"

	"github.com/uchup07/fyne-jira-worklog-tracker/jira"
)

func TestAggregateByIrqGroups(t *testing.T) {
	rows := []jira.MappingRow{
		{JSWKey: "JSW-1", IRQKey: "IRQ-1", IRQSummary: "Feature A", Vertical: "Core", Company: "Acme", TotalSeconds: 3600},
		{JSWKey: "JSW-2", IRQKey: "IRQ-1", IRQSummary: "Feature A", Vertical: "Core", Company: "Acme", TotalSeconds: 7200},
		{JSWKey: "JSW-3", IRQKey: "IRQ-2", IRQSummary: "Feature B", Vertical: "Data", Company: "Beta", TotalSeconds: 1800},
	}

	groups := jira.AggregateByIRQ(rows)

	if len(groups) != 2 {
		t.Fatalf("expected 2 IRQ groups, got %d", len(groups))
	}

	g1 := findIRQGroup(groups, "IRQ-1")
	if g1 == nil {
		t.Fatal("IRQ-1 group not found")
	}
	if len(g1.JSWTasks) != 2 {
		t.Errorf("IRQ-1 tasks: got %d, want 2", len(g1.JSWTasks))
	}
	if g1.TotalSeconds != 10800 {
		t.Errorf("IRQ-1 total: got %d, want 10800", g1.TotalSeconds)
	}
}

func TestAggregateByIrqDeduplicatesTasks(t *testing.T) {
	rows := []jira.MappingRow{
		{JSWKey: "JSW-1", IRQKey: "IRQ-1", TotalSeconds: 3600},
		{JSWKey: "JSW-1", IRQKey: "IRQ-1", TotalSeconds: 3600}, // duplicate
	}
	groups := jira.AggregateByIRQ(rows)
	if len(groups) != 1 || len(groups[0].JSWTasks) != 1 {
		t.Errorf("duplicate JSW task not de-duplicated: %+v", groups)
	}
}

func TestVerticalChartData(t *testing.T) {
	rows := []jira.MappingRow{
		{Vertical: "Core", TotalSeconds: 7200},
		{Vertical: "Core", TotalSeconds: 3600},
		{Vertical: "Data", TotalSeconds: 1800},
	}
	cd := jira.VerticalChartData(rows)
	if len(cd.Labels) != 2 {
		t.Fatalf("expected 2 verticals, got %d", len(cd.Labels))
	}
	if cd.Labels[0] != "Core" {
		t.Errorf("expected Core first (sorted desc), got %q", cd.Labels[0])
	}
	if cd.Values[0] != 3.0 {
		t.Errorf("Core hours: got %f, want 3.0", cd.Values[0])
	}
}

func findIRQGroup(groups []jira.IrqGroup, irqKey string) *jira.IrqGroup {
	for i := range groups {
		if groups[i].IRQKey == irqKey {
			return &groups[i]
		}
	}
	return nil
}
