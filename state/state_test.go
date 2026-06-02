// state/state_test.go
package state_test

import (
	"os"
	"testing"
	"time"

	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/test"

	"github.com/uchup07/fyne-jira-worklog-tracker/state"
)

func TestMain(m *testing.M) {
	// Fyne bindings require a running app even in tests.
	test.NewApp()
	os.Exit(m.Run())
}

func TestFilterStateDefaults(t *testing.T) {
	fs := state.NewFilterState()
	today := time.Now().Format("2006-01-02")

	start, err := fs.StartDate.Get()
	if err != nil {
		t.Fatalf("StartDate.Get: %v", err)
	}
	if start != today {
		t.Errorf("StartDate default: got %q, want %q", start, today)
	}

	end, _ := fs.EndDate.Get()
	if end != today {
		t.Errorf("EndDate default: got %q, want %q", end, today)
	}

	mode, _ := fs.DateMode.Get()
	if mode != "day" {
		t.Errorf("DateMode default: got %q, want %q", mode, "day")
	}
}

func TestFilterStateSet(t *testing.T) {
	fs := state.NewFilterState()

	if err := fs.StartDate.Set("2026-01-01"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, _ := fs.StartDate.Get()
	if got != "2026-01-01" {
		t.Errorf("got %q, want %q", got, "2026-01-01")
	}
}

func TestWorklogStateDefaults(t *testing.T) {
	ws := state.NewWorklogState()

	tab, err := ws.ActiveTab.Get()
	if err != nil {
		t.Fatalf("ActiveTab.Get: %v", err)
	}
	if tab != "workreference" {
		t.Errorf("ActiveTab default: got %q, want %q", tab, "workreference")
	}

	loading, _ := ws.IsLoading.Get()
	if loading {
		t.Error("IsLoading should default to false")
	}
}

func TestBindingListenerFires(t *testing.T) {
	ws := state.NewWorklogState()

	fired := make(chan struct{}, 1)
	ws.IsLoading.AddListener(binding.NewDataListener(func() {
		select {
		case fired <- struct{}{}:
		default:
		}
	}))

	if err := ws.IsLoading.Set(true); err != nil {
		t.Fatalf("Set: %v", err)
	}

	select {
	case <-fired:
		// pass — listener was notified
	default:
		t.Error("listener was not notified after Set(true)")
	}
}

func TestReportStateDefaults(t *testing.T) {
	rs := state.NewReportState()

	val, err := rs.MappingReport.Get()
	if err != nil {
		t.Fatalf("MappingReport.Get: %v", err)
	}
	if val != nil {
		t.Error("MappingReport should default to nil")
	}
}
