package ui

import (
	"os"

	"gitty/internal/git"
	"gitty/internal/ui/menu"

	tea "github.com/charmbracelet/bubbletea"
)

type state int

const (
	stateNoGit state = iota
	stateGit
)

type Model struct {
	state    state
	noGit    menu.NoGitModel
	quitting bool
}

// make a new fresh model and detect the git repo to pick the first state
func New() Model {
	cwd, _ := os.Getwd()
	s := stateNoGit
	if git.IsRepo(cwd) {
		s = stateGit
	}

	return Model{
		state: s,
		noGit: menu.NewNoGit(),
	}
}

func (m Model) Init() tea.Cmd {
	switch m.state {
	case stateNoGit:
		return m.noGit.Init()
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case menu.ChoiceMsg:
		return m.handleNoGitOption(msg)
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
	}

	switch m.state {
	case stateNoGit:
		updated, cmd := m.noGit.Update(msg)
		m.noGit = updated.(menu.NoGitModel)
		return m, cmd
	}

	return m, nil
}

// renders the active sub-model
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	switch m.state {
	case stateNoGit:
		return m.noGit.View()
	case stateGit:
		return "wow! git repo, main menu coming soon\n"
	}

	return ""
}

func (m Model) handleNoGitOption(msg menu.ChoiceMsg) (tea.Model, tea.Cmd) {
	switch msg.Choice {
	case "Quit":
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}
