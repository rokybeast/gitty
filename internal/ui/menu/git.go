package menu

import (
	"fmt"

	"gitty/internal/git"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type GitChoiceMsg struct {
	ID string
}

var gitTitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#88c0d0")). // nord frost blue
	PaddingLeft(2)

// nord-themed list delegate
func nordListDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.Styles.NormalTitle = d.Styles.NormalTitle.Foreground(lipgloss.Color("#eceff4"))
	d.Styles.NormalDesc = d.Styles.NormalDesc.Foreground(lipgloss.Color("#4c566a"))
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#88c0d0")).
		BorderLeftForeground(lipgloss.Color("#88c0d0"))
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#81a1c1")).
		BorderLeftForeground(lipgloss.Color("#88c0d0"))
	d.Styles.DimmedTitle = d.Styles.DimmedTitle.Foreground(lipgloss.Color("#4c566a"))
	d.Styles.DimmedDesc = d.Styles.DimmedDesc.Foreground(lipgloss.Color("#4c566a"))
	d.Styles.FilterMatch = d.Styles.FilterMatch.Foreground(lipgloss.Color("#a3be8c"))
	return d
}

type GitModel struct {
	list   list.Model
	choice string
	width  int
	height int
}

// the main git menu
func NewGit(width, height int) GitModel {
	items := []list.Item{
		item{id: IDAddFiles, title: "󰝒 Add Files", desc: "stage or unstage files for commit"},
		item{id: IDCommit, title: "󰜘 Commit", desc: "stage and write commits"},
		item{id: IDPush, title: " Push Commits", desc: "push the commits to different remotes"},
		item{id: IDTree, title: "󰙅 Project Tree", desc: "view and manage tracked files"},
		item{id: IDHistory, title: "󰋚 Commit History", desc: "browse the commit log with a nice graph"},
		item{id: IDOtherTools, title: "󱈧 Other Git Tools", desc: "merge, rebase, reset, restore, fetch, pull, status and more..."},
		item{id: IDAbout, title: "󰋼 About gitty", desc: "info about gitty"},
		item{id: IDQuit, title: "󰈆 Quit", desc: "exit gitty :("},
	}

	branch := git.CurrentBranch()
	repoName := git.RepoName()
	l := list.New(items, nordListDelegate(), width, height)
	l.Title = fmt.Sprintf("gitty - v0.3.0 (unstable; not yet released) [󰘬 %s/%s]", repoName, branch)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(true)
	l.Styles.Title = gitTitleStyle

	return GitModel{list: l, width: width, height: height}
}

// no init command needed
func (m GitModel) Init() tea.Cmd {
	return nil
}

// handle input and window resizes
func (m GitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			selected, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = selected.id
				return m, func() tea.Msg {
					return GitChoiceMsg{ID: selected.id}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// render the list
func (m GitModel) View() string {
	return m.list.View()
}
