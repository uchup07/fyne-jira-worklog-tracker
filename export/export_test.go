// export/export_test.go
package export_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/uchup07/fyne-jira-worklog-tracker/export"
	"github.com/uchup07/fyne-jira-worklog-tracker/jira"
)

var sampleGroups = []jira.WorklogGroup{
	{
		WorkReference: "CR-001",
		TotalSeconds:  10800,
		Items: []jira.WorklogItem{
			{IssueKey: "JSW-1", IssueSummary: "Task A", Author: jira.User{DisplayName: "Alice"}, TimeSpentSeconds: 7200, Started: time.Date(2026, 1, 15, 9, 0, 0, 0, time.UTC)},
			{IssueKey: "JSW-2", IssueSummary: "Task B", Author: jira.User{DisplayName: "Bob"}, TimeSpentSeconds: 3600, Started: time.Date(2026, 1, 15, 14, 0, 0, 0, time.UTC)},
		},
	},
}

func TestWriteCSV(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")
	if err := export.WriteCSV(path, sampleGroups); err != nil {
		t.Fatalf("WriteCSV: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read CSV: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty CSV file")
	}
	content := string(data)
	if !strings.Contains(content, "CR-001") {
		t.Error("CSV missing work reference 'CR-001'")
	}
	if !strings.Contains(content, "Alice") {
		t.Error("CSV missing author 'Alice'")
	}
	if !strings.Contains(content, "JSW-1") {
		t.Error("CSV missing issue key 'JSW-1'")
	}
}

func TestWriteExcel(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.xlsx")
	if err := export.WriteExcel(path, sampleGroups); err != nil {
		t.Fatalf("WriteExcel: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat xlsx: %v", err)
	}
	if info.Size() == 0 {
		t.Error("expected non-empty xlsx file")
	}
}

func TestWritePDF(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.pdf")
	if err := export.WritePDF(path, sampleGroups, "2026-01-01", "2026-01-31"); err != nil {
		t.Fatalf("WritePDF: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat pdf: %v", err)
	}
	if info.Size() == 0 {
		t.Error("expected non-empty PDF file")
	}
}
