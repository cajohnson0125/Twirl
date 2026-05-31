package tui

import (
	"os"

	"charm.land/lipgloss/v2"
)

// Semantic colors derived from the terminal's own palette.
//
// ANSI indices 0-15 map to the 16 colors the user configured in their
// terminal emulator.  Using LightDark to pick standard vs bright variants
// means the UI automatically matches whatever theme the user has.
//
// Mapping (index → standard name):
//
//	0 Black    1 Red     2 Green   3 Yellow
//	4 Blue     5 Magenta 6 Cyan    7 White
//	8 BrBlack  9 BrRed  10 BrGreen 11 BrYellow
//	12 BrBlue 13 BrMagenta 14 BrCyan 15 BrWhite

var ld = lipgloss.LightDark(
	lipgloss.HasDarkBackground(os.Stdin, os.Stdout),
)

var (
	// clrAccent — vibrant highlight for prompts and accents.
	clrAccent = ld(lipgloss.Color("5"), lipgloss.Color("13"))

	// clrActive — green for active/running states.
	clrActive = ld(lipgloss.Color("2"), lipgloss.Color("10"))

	// clrMuted — subdued text for idle states, labels, footer.
	clrMuted = ld(lipgloss.Color("0"), lipgloss.Color("8"))

	// clrText — primary readable text.
	clrText = ld(lipgloss.Color("0"), lipgloss.Color("15"))

	// clrBorder — panel border color.
	clrBorder = ld(lipgloss.Color("7"), lipgloss.Color("8"))

	// clrTitle — bold panel title color.
	clrTitle = ld(lipgloss.Color("5"), lipgloss.Color("13"))
)

var (
	stylePanelBorder = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(clrBorder)

	stylePanelTitle = lipgloss.NewStyle().
				Bold(true).
				Foreground(clrTitle)

	styleActive = lipgloss.NewStyle().Foreground(clrActive)
	styleIdle   = lipgloss.NewStyle().Foreground(clrMuted)
	styleLabel  = lipgloss.NewStyle().Foreground(clrMuted)
	styleValue  = lipgloss.NewStyle().Bold(true).Foreground(clrText)
	styleFooter = lipgloss.NewStyle().Foreground(clrMuted)

	styleInputBar = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(clrBorder)
)
