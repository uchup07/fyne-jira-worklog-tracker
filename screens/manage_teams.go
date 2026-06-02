// screens/manage_teams.go
package screens

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// ManageTeams is the team/member CRUD screen.
// Full implementation comes in Plan 4.
type ManageTeams struct {
	canvas fyne.CanvasObject
}

func NewManageTeams() *ManageTeams {
	m := &ManageTeams{}
	m.canvas = widget.NewLabel("Manage Teams — full implementation in Plan 4")
	return m
}

func (m *ManageTeams) Canvas() fyne.CanvasObject { return m.canvas }
