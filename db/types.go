// db/types.go
package db

import "time"

// AppConfig holds Jira connection settings.
// ApiToken is NOT stored in the DB — it lives in fyne.Preferences (OS keychain).
type AppConfig struct {
	JiraDomain      string
	Email           string
	ApiToken        string // populated by caller from fyne.Preferences
	WorkRefFieldID  string
	VerticalFieldID string
	CompanyFieldID  string
	UatEndFieldID   string
}

// Team is a named group of Jira users.
type Team struct {
	ID   int
	Name string
}

// TeamMember links a Jira user to a Team with optional date range.
type TeamMember struct {
	ID        int
	TeamID    int
	UserID    string // Jira accountId
	JoinDate  string // "2006-01-02" or ""
	LeaveDate string // "2006-01-02" or ""
}

// PublicHoliday is a non-working day used in timesheet calculations.
type PublicHoliday struct {
	ID   int
	Date string // "2006-01-02"
	Name string
}

// JiraUser is a cached Jira user record.
type JiraUser struct {
	ID           string // Jira accountId
	DisplayName  string
	EmailAddress string
	AvatarURL    string
	Active       bool
	SyncedAt     time.Time
}
