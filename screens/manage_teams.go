// screens/manage_teams.go
package screens

import (
	"fmt"

	"github.com/uchup07/fyne-jira-worklog-tracker/db"
	"github.com/uchup07/fyne-jira-worklog-tracker/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ManageTeams is the team and team-member CRUD screen.
type ManageTeams struct {
	repo       *db.Repository
	tr         *i18n.I18n
	window     fyne.Window
	canvas     fyne.CanvasObject
	teams      []db.Team
	members    []db.TeamMember
	selectedID int
	teamList   *widget.List
	memberTable *widget.Table
}

// NewManageTeams creates the Manage Teams screen.
func NewManageTeams(repo *db.Repository, tr *i18n.I18n, window fyne.Window) *ManageTeams {
	m := &ManageTeams{repo: repo, tr: tr, window: window}
	m.build()
	m.reload()
	return m
}

// Canvas returns the Fyne canvas object.
func (m *ManageTeams) Canvas() fyne.CanvasObject { return m.canvas }

func (m *ManageTeams) build() {
	m.teamList = widget.NewList(
		func() int { return len(m.teams) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id < len(m.teams) {
				item.(*widget.Label).SetText(m.teams[id].Name)
			}
		},
	)
	m.teamList.OnSelected = func(id widget.ListItemID) {
		if id < len(m.teams) {
			m.selectedID = m.teams[id].ID
			m.reloadMembers()
		}
	}

	addTeamBtn := widget.NewButton(m.tr.T("teams.btn.addTeam"), m.showAddTeamDialog)
	delTeamBtn := widget.NewButton(m.tr.T("teams.btn.deleteTeam"), m.deleteSelectedTeam)

	leftPanel := container.NewBorder(
		nil,
		container.NewHBox(addTeamBtn, delTeamBtn),
		nil, nil,
		m.teamList,
	)

	cols := []string{
		m.tr.T("teams.col.user"),
		m.tr.T("teams.col.joinDate"),
		m.tr.T("teams.col.leaveDate"),
		"",
	}
	colWidths := []float32{200, 100, 100, 80}

	m.memberTable = widget.NewTable(
		func() (int, int) { return len(m.members), len(cols) },
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewLabel(""))
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			if id.Row >= len(m.members) {
				return
			}
			mem := m.members[id.Row]
			box := cell.(*fyne.Container)
			switch id.Col {
			case 0:
				setBoxLabel(box, mem.UserID)
			case 1:
				setBoxLabel(box, mem.JoinDate)
			case 2:
				setBoxLabel(box, mem.LeaveDate)
			case 3:
				btn := widget.NewButton(m.tr.T("teams.btn.removeMember"), func() {
					m.repo.Teams.RemoveMember(mem.ID)
					m.reloadMembers()
				})
				box.Objects = []fyne.CanvasObject{btn}
				box.Refresh()
			}
		},
	)
	for i, w := range colWidths {
		m.memberTable.SetColumnWidth(i, w)
	}

	addMemberBtn := widget.NewButton(m.tr.T("teams.btn.addMember"), m.showAddMemberDialog)
	rightPanel := container.NewBorder(nil, addMemberBtn, nil, nil, m.memberTable)

	m.canvas = container.NewHSplit(leftPanel, rightPanel)
}

func (m *ManageTeams) reload() {
	teams, err := m.repo.Teams.ListTeams()
	if err != nil {
		return
	}
	m.teams = teams
	m.teamList.Refresh()
}

func (m *ManageTeams) reloadMembers() {
	if m.selectedID == 0 {
		m.members = nil
		m.memberTable.Refresh()
		return
	}
	members, err := m.repo.Teams.ListMembers(m.selectedID)
	if err != nil {
		return
	}
	m.members = members
	m.memberTable.Refresh()
}

func (m *ManageTeams) showAddTeamDialog() {
	entry := widget.NewEntry()
	entry.PlaceHolder = m.tr.T("teams.dialog.newTeam")
	dialog.ShowForm(m.tr.T("teams.btn.addTeam"), "Add", "Cancel",
		[]*widget.FormItem{widget.NewFormItem("Name", entry)},
		func(ok bool) {
			if !ok || entry.Text == "" {
				return
			}
			if _, err := m.repo.Teams.CreateTeam(entry.Text); err != nil {
				dialog.ShowError(fmt.Errorf("create team: %w", err), m.window)
				return
			}
			m.reload()
		}, m.window)
}

func (m *ManageTeams) deleteSelectedTeam() {
	if m.selectedID == 0 {
		return
	}
	dialog.ShowConfirm("Delete Team", "Delete this team and all its members?", func(ok bool) {
		if !ok {
			return
		}
		if err := m.repo.Teams.DeleteTeam(m.selectedID); err != nil {
			dialog.ShowError(fmt.Errorf("delete team: %w", err), m.window)
			return
		}
		m.selectedID = 0
		m.members = nil
		m.memberTable.Refresh()
		m.reload()
	}, m.window)
}

func (m *ManageTeams) showAddMemberDialog() {
	if m.selectedID == 0 {
		dialog.ShowInformation("No Team Selected", "Please select a team first.", m.window)
		return
	}
	userEntry := widget.NewEntry()
	userEntry.PlaceHolder = "Jira accountId"
	joinEntry := widget.NewEntry()
	joinEntry.PlaceHolder = "2006-01-02"
	leaveEntry := widget.NewEntry()
	leaveEntry.PlaceHolder = "2006-01-02 (optional)"

	dialog.ShowForm(m.tr.T("teams.btn.addMember"), "Add", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem(m.tr.T("teams.col.user"), userEntry),
			widget.NewFormItem(m.tr.T("teams.col.joinDate"), joinEntry),
			widget.NewFormItem(m.tr.T("teams.col.leaveDate"), leaveEntry),
		},
		func(ok bool) {
			if !ok || userEntry.Text == "" {
				return
			}
			if err := m.repo.Teams.AddMember(db.TeamMember{
				TeamID:    m.selectedID,
				UserID:    userEntry.Text,
				JoinDate:  joinEntry.Text,
				LeaveDate: leaveEntry.Text,
			}); err != nil {
				dialog.ShowError(fmt.Errorf("add member: %w", err), m.window)
				return
			}
			m.reloadMembers()
		}, m.window)
}

func setBoxLabel(box *fyne.Container, text string) {
	if len(box.Objects) == 0 {
		box.Objects = []fyne.CanvasObject{widget.NewLabel(text)}
	} else {
		box.Objects[0].(*widget.Label).SetText(text)
	}
	box.Refresh()
}
