// i18n/i18n_test.go
package i18n_test

import (
	"testing"

	"github.com/uchup07/fyne-jira-worklog-tracker/i18n"
)

func TestTReturnsEnglishString(t *testing.T) {
	tr := i18n.New("en")
	got := tr.T("nav.dashboard")
	if got != "Dashboard" {
		t.Errorf("got %q, want %q", got, "Dashboard")
	}
}

func TestTReturnsIndonesianString(t *testing.T) {
	tr := i18n.New("id")
	got := tr.T("nav.dashboard")
	if got != "Dasbor" {
		t.Errorf("got %q, want %q", got, "Dasbor")
	}
}

func TestTSubstitutesParams(t *testing.T) {
	tr := i18n.New("en")
	got := tr.T("dashboard.progress.searching", map[string]any{"found": 42})
	want := "Searching... 42 issues found"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTReturnsMissingKeyAsKey(t *testing.T) {
	tr := i18n.New("en")
	got := tr.T("nonexistent.key")
	if got != "nonexistent.key" {
		t.Errorf("expected key passthrough, got %q", got)
	}
}

func TestSetLangSwitches(t *testing.T) {
	tr := i18n.New("en")
	tr.SetLang("id")
	got := tr.T("nav.dashboard")
	if got != "Dasbor" {
		t.Errorf("after SetLang(id): got %q, want %q", got, "Dasbor")
	}
}

func TestFallbackToEnglishForUnknownLang(t *testing.T) {
	tr := i18n.New("fr") // French not supported — falls back to English
	got := tr.T("nav.dashboard")
	if got != "Dashboard" {
		t.Errorf("fallback: got %q, want %q", got, "Dashboard")
	}
}
