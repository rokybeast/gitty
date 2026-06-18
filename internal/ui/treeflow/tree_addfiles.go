package treeflow

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	afTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#88c0d0")). // nord frost blue
			PaddingLeft(2)

	afStagedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a3be8c")) // nord green

	afUnstagedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#bf616a")) // nord red

	afUntrackedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#d08770")) // nord orange

	afDeletedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#bf616a")). // nord red
			Strikethrough(true)

	afCursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#88c0d0")). // nord frost blue
			Bold(true)

	afHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4c566a")). // nord muted gray
			PaddingLeft(2)

	afDirStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#81a1c1")). // nord frost
			Bold(true)

	afHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#eceff4")). // nord snow
			Bold(true).
			PaddingLeft(2)
)

// a single dirty/untracked file entry
type dirtyFile struct {
	path      string
	name      string
	status    string
	staged    bool
	deleted   bool
	untracked bool
	dir       string
}

// ts only shows dirty, untracked, and deleted files
type AddFilesModel struct {
	files  []dirtyFile
	cursor int
	width  int
	height int
}

// make a new addfiles model by scanning git status
func NewAddFiles(width, height int) AddFilesModel {
	m := AddFilesModel{
		width:  width,
		height: height,
	}
	m.refresh()
	return m
}

func (m AddFilesModel) Init() tea.Cmd {
	return nil
}

func (m AddFilesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc", "q":
			return m, func() tea.Msg { return BackMsg{} }
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.files)-1 {
				m.cursor++
			}
		case "a":
			// staging for sel file
			if len(m.files) > 0 && m.cursor < len(m.files) {
				f := m.files[m.cursor]
				afToggleStaging(f)
				m.refresh()
			}
		case "A":
			// stage all at once (for 'em bulk committers)
			afStageAll()
			m.refresh()
		case "g":
			m.cursor = 0
		case "G":
			if len(m.files) > 0 {
				m.cursor = len(m.files) - 1
			}
		}
	}
	return m, nil
}

func (m AddFilesModel) View() string {
	var view strings.Builder

	title := afTitleStyle.Render("add files (a: stage/unstage, A: stage all, esc/q: back)")
	view.WriteString("\n" + title + "\n\n")

	if len(m.files) == 0 {
		view.WriteString(afHintStyle.Render("  nothing to stage, working tree is clean.") + "\n")
		view.WriteString("\n" + afHintStyle.Render("esc/q: back"))
		return view.String()
	}

	// count staged vs unstaged for the header
	staged, unstaged := 0, 0
	for _, f := range m.files {
		if f.staged {
			staged++
		} else {
			unstaged++
		}
	}
	header := fmt.Sprintf("  %d file%s changed (%d staged, %d unstaged)",
		len(m.files), afPlural(len(m.files)), staged, unstaged)
	view.WriteString(afHeaderStyle.Render(header) + "\n\n")

	maxVisible := m.height - 8
	if maxVisible < 1 {
		maxVisible = 1
	}

	start := m.cursor - maxVisible/2
	if start < 0 {
		start = 0
	}
	end := start + maxVisible
	if end > len(m.files) {
		end = len(m.files)
		start = end - maxVisible
		if start < 0 {
			start = 0
		}
	}

	// track which dirs we've printed headers for
	lastDir := ""

	for i := start; i < end; i++ {
		f := m.files[i]

		// print a directory header if we moved into a new dir
		if f.dir != lastDir {
			dirLabel := f.dir
			if dirLabel == "" {
				dirLabel = "."
			}
			view.WriteString("  " + afDirStyle.Render("󰉋 "+dirLabel+"/") + "\n")
			lastDir = f.dir
		}

		// check if this is the last file in its directory group
		isLastInDir := true
		if i+1 < len(m.files) && m.files[i+1].dir == f.dir {
			isLastInDir = false
		}

		// tree connector
		connector := "├─ "
		if isLastInDir {
			connector = "└─ "
		}

		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		// pick icon and style based on file state
		icon, style := afFileStyle(f)

		// staging indicator
		stageTag := afUnstagedStyle.Render("  ") // nf-cod-circle
		if f.staged {
			stageTag = afStagedStyle.Render("  ") //nf-cod-circle-fill
		}

		// the status code
		statusTag := afHintStyle.Render(fmt.Sprintf(" [%s]", strings.TrimSpace(f.status)))

		styledConnector := afDirStyle.Render(connector)

		if i == m.cursor {
			line := afCursorStyle.Render(fmt.Sprintf("%s%s%s %s", cursor, connector, icon, f.name))
			view.WriteString(line + stageTag + statusTag + "\n")
		} else {
			line := fmt.Sprintf("%s%s%s %s", cursor, styledConnector, icon, style.Render(f.name))
			view.WriteString(line + stageTag + statusTag + "\n")
		}
	}

	if len(m.files) > maxVisible {
		pos := fmt.Sprintf(" [%d/%d]", m.cursor+1, len(m.files))
		view.WriteString("\n" + afHintStyle.Render("esc/q: back | a: toggle | A: stage all"+pos))
	} else {
		view.WriteString("\n" + afHintStyle.Render("esc/q: back | a: toggle | A: stage all"))
	}

	return view.String()
}

// refresh the file list from git status
func (m *AddFilesModel) refresh() {
	m.files = afGetDirtyFiles()
	if m.cursor >= len(m.files) {
		m.cursor = len(m.files) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

// grab EVERY dirty/untracked files from git status --porcelain
func afGetDirtyFiles() []dirtyFile {
	cmd := exec.Command("git", "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var files []dirtyFile
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if len(line) < 4 {
			continue
		}

		status := line[:2]
		rawPath := line[3:]

		// handle renames: "old -> new"
		parts := strings.SplitN(rawPath, " -> ", 2)
		path := parts[len(parts)-1]
		path = strings.Trim(path, "\"")

		f := dirtyFile{
			path:   path,
			name:   filepath.Base(path),
			status: status,
			dir:    filepath.Dir(path),
		}

		// figure out staging state from the status codes
		// first char = index (staged), second char = worktree
		indexCode := status[0]
		workCode := status[1]

		// untracked files
		if status == "??" {
			f.untracked = true
			f.staged = false
		} else if indexCode != ' ' && indexCode != '?' {
			f.staged = true
		}

		// deleted files (either staged delete or worktree delete)
		if indexCode == 'D' || workCode == 'D' {
			f.deleted = true
		}

		files = append(files, f)
	}

	// sort: dirs first (via path), then alphabetical
	sort.Slice(files, func(i, j int) bool {
		if files[i].dir != files[j].dir {
			return files[i].dir < files[j].dir
		}
		return files[i].name < files[j].name
	})

	return files
}

// toggle staging for a single file
func afToggleStaging(f dirtyFile) {
	if f.staged {
		cmd := exec.Command("git", "reset", "HEAD", "--", f.path)
		_ = cmd.Run()
	} else {
		cmd := exec.Command("git", "add", "--", f.path)
		_ = cmd.Run()
	}
}

// stage all dirty files
func afStageAll() {
	cmd := exec.Command("git", "add", "-A")
	_ = cmd.Run()
}

// pick the right icon and style for a file based on its state
func afFileStyle(f dirtyFile) (string, lipgloss.Style) {
	if f.deleted {
		return "󰮘", afDeletedStyle // nf-md-file_remove (with strike style)
	}
	if f.untracked {
		return "󰝒", afUntrackedStyle // nf-md-file_plus
	}
	if f.staged {
		return "󰄬", afStagedStyle // nf-md-check
	}
	return "󱇧", afUnstagedStyle // nf-md-file_edit
}

// "s" for plural
func afPlural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
