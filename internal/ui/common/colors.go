package common

import "github.com/charmbracelet/lipgloss"

// nord colors
var (
	// frost colors
	ColorFrostBlue      = lipgloss.Color("#88c0d0")
	ColorFrostLightBlue = lipgloss.Color("#81a1c1")

	// snow colors
	ColorSnow     = lipgloss.Color("#eceff4")
	ColorSnowDark = lipgloss.Color("#d8dee9")

	// muted colors
	ColorMutedGray     = lipgloss.Color("#4c566a")
	ColorMutedGrayDark = lipgloss.Color("#525252")

	// status colors
	ColorGreen  = lipgloss.Color("#a3be8c") // success
	ColorRed    = lipgloss.Color("#bf616a") // error
	ColorOrange = lipgloss.Color("#d08770") // warning

	// diff colors
	ColorBgDiffAdd    = lipgloss.Color("#2e3b32")
	ColorBgDiffDelete = lipgloss.Color("#3b2e2e")
	ColorBgDiffLine   = lipgloss.Color("#3b4252")
)
