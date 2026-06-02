// i18n/i18n.go
package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed locales/*.json
var localesFS embed.FS

// I18n holds the active language strings. Pass *I18n to every screen and widget.
type I18n struct {
	lang    string
	strings map[string]string
}

// New loads the locale file for lang ("en" or "id").
// Falls back to English if the locale file is missing.
func New(lang string) *I18n {
	i := &I18n{}
	i.load(lang)
	return i
}

func (i *I18n) load(lang string) {
	data, err := localesFS.ReadFile(fmt.Sprintf("locales/%s.json", lang))
	if err != nil {
		data, _ = localesFS.ReadFile("locales/en.json")
		lang = "en"
	}
	var m map[string]string
	json.Unmarshal(data, &m)
	i.lang = lang
	i.strings = m
}

// Lang returns the currently active language code ("en" or "id").
func (i *I18n) Lang() string { return i.lang }

// T looks up key and substitutes {param} placeholders from params.
// Returns the raw key if not found — makes missing strings obvious during development.
//
// Example:
//
//	tr.T("dashboard.progress.searching", map[string]any{"found": 42})
//	// → "Searching... 42 issues found"
func (i *I18n) T(key string, params ...map[string]any) string {
	s, ok := i.strings[key]
	if !ok {
		return key
	}
	if len(params) > 0 {
		for k, v := range params[0] {
			s = strings.ReplaceAll(s, fmt.Sprintf("{%s}", k), fmt.Sprintf("%v", v))
		}
	}
	return s
}

// SetLang reloads strings for the given language code.
// Call window.Content().Refresh() afterwards to repaint all labels.
func (i *I18n) SetLang(lang string) {
	i.load(lang)
}
