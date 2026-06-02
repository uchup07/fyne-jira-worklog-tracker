// export/csv.go
package export

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/uchup07/fyne-jira-worklog-tracker/jira"
)

// WriteCSV writes worklog groups to a CSV file at path.
func WriteCSV(path string, groups []jira.WorklogGroup) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create CSV: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	w.Write([]string{"Work Reference", "Issue Key", "Summary", "Author", "Hours", "Date", "Comment"})

	for _, group := range groups {
		for _, item := range group.Items {
			w.Write([]string{
				group.WorkReference,
				item.IssueKey,
				item.IssueSummary,
				item.Author.DisplayName,
				fmt.Sprintf("%.2f", float64(item.TimeSpentSeconds)/3600),
				item.Started.Format("2006-01-02"),
				item.Comment,
			})
		}
	}
	return w.Error()
}
