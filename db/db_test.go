// db/db_test.go
package db_test

import (
	"testing"

	"github.com/uchup07/fyne-jira-worklog-tracker/db"
)

// openTestDB returns an in-memory database — isolated per test call.
func openTestDB(t *testing.T) *db.Repository {
	t.Helper()
	repo := db.Open(":memory:")
	t.Cleanup(func() { repo.Close() })
	return repo
}

// --- Config repo ---

func TestConfigGetEmptyDB(t *testing.T) {
	repo := openTestDB(t)
	cfg, err := repo.Config.Get()
	if err != nil {
		t.Fatalf("Get on empty DB: %v", err)
	}
	if cfg != nil {
		t.Error("expected nil on empty DB")
	}
}

func TestConfigSaveAndGet(t *testing.T) {
	repo := openTestDB(t)
	want := &db.AppConfig{
		JiraDomain:     "myorg.atlassian.net",
		Email:          "me@example.com",
		WorkRefFieldID: "customfield_10001",
	}
	if err := repo.Config.Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := repo.Config.Get()
	if err != nil {
		t.Fatalf("Get after Save: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil after save")
	}
	if got.JiraDomain != want.JiraDomain {
		t.Errorf("JiraDomain: got %q, want %q", got.JiraDomain, want.JiraDomain)
	}
	if got.Email != want.Email {
		t.Errorf("Email: got %q, want %q", got.Email, want.Email)
	}
	if got.WorkRefFieldID != want.WorkRefFieldID {
		t.Errorf("WorkRefFieldID: got %q, want %q", got.WorkRefFieldID, want.WorkRefFieldID)
	}
	// ApiToken is NOT stored in DB — should always be empty from Get()
	if got.ApiToken != "" {
		t.Error("ApiToken should not be persisted in DB")
	}
}

func TestConfigSaveUpserts(t *testing.T) {
	repo := openTestDB(t)
	repo.Config.Save(&db.AppConfig{JiraDomain: "old.atlassian.net", Email: "old@e.com"})
	repo.Config.Save(&db.AppConfig{JiraDomain: "new.atlassian.net", Email: "new@e.com"})

	cfg, _ := repo.Config.Get()
	if cfg.JiraDomain != "new.atlassian.net" {
		t.Errorf("upsert failed: got %q", cfg.JiraDomain)
	}
}

// --- User repo ---

func TestUserUpsertAndList(t *testing.T) {
	repo := openTestDB(t)
	users := []db.JiraUser{
		{ID: "u1", DisplayName: "Alice", EmailAddress: "alice@e.com", Active: true},
		{ID: "u2", DisplayName: "Bob", EmailAddress: "bob@e.com", Active: true},
	}
	if err := repo.Users.Upsert(users); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	got, err := repo.Users.ListActive()
	if err != nil {
		t.Fatalf("ListActive: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 users, got %d", len(got))
	}
}

// --- Team repo ---

func TestTeamCreateAndList(t *testing.T) {
	repo := openTestDB(t)

	id, err := repo.Teams.CreateTeam("Backend")
	if err != nil {
		t.Fatalf("CreateTeam: %v", err)
	}
	if id == 0 {
		t.Error("expected non-zero ID")
	}

	teams, err := repo.Teams.ListTeams()
	if err != nil {
		t.Fatalf("ListTeams: %v", err)
	}
	if len(teams) != 1 || teams[0].Name != "Backend" {
		t.Errorf("unexpected teams: %+v", teams)
	}
}

func TestTeamMemberAddAndList(t *testing.T) {
	repo := openTestDB(t)
	teamID, _ := repo.Teams.CreateTeam("Frontend")

	err := repo.Teams.AddMember(db.TeamMember{
		TeamID:   teamID,
		UserID:   "user-abc",
		JoinDate: "2026-01-01",
	})
	if err != nil {
		t.Fatalf("AddMember: %v", err)
	}

	members, err := repo.Teams.ListMembers(teamID)
	if err != nil {
		t.Fatalf("ListMembers: %v", err)
	}
	if len(members) != 1 || members[0].UserID != "user-abc" {
		t.Errorf("unexpected members: %+v", members)
	}
}

// --- Holiday repo ---

func TestHolidayAddAndList(t *testing.T) {
	repo := openTestDB(t)

	err := repo.Holidays.Add(db.PublicHoliday{Date: "2026-08-17", Name: "Independence Day"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	holidays, err := repo.Holidays.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(holidays) != 1 || holidays[0].Name != "Independence Day" {
		t.Errorf("unexpected holidays: %+v", holidays)
	}
}
