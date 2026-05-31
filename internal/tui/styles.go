package tui

import "github.com/charmbracelet/lipgloss"

// Semantic colors derived from the terminal's own palette.
//
// ANSI indices 0-15 map to the 16 colors the user configured in their
// terminal emulator.  Using them means the UI automatically matches
// whatever theme the user has — dark, light, or custom — without any
// hardcoded color values.
//
// Mapping (index → standard name):
//
//	0 Black    1 Red     2 Green   3 Yellow
//	4 Blue     5 Magenta 6 Cyan    7 White
//	8 BrBlack  9 BrRed  10 BrGreen 11 BrYellow
//	12 BrBlue 13 BrMagenta 14 BrCyan 15 BrWhite
var (
	// clrAccent — vibrant highlight for spinners and prompts.
	clrAccent = lipgloss.AdaptiveColor{Light: "5", Dark: "13"}

	// clrActive — green for active/running states.
	clrActive = lipgloss.AdaptiveColor{Light: "2", Dark: "10"}

	// clrDone — yellow/gold for completed states.
	clrDone = lipgloss.AdaptiveColor{Light: "3", Dark: "11"}

	// clrMuted — subdued text for idle states, labels, footer.
	clrMuted = lipgloss.AdaptiveColor{Light: "0", Dark: "8"}

	// clrText — primary readable text.
	clrText = lipgloss.AdaptiveColor{Light: "0", Dark: "15"}

	// clrBorder — panel border color.
	clrBorder = lipgloss.AdaptiveColor{Light: "7", Dark: "8"}

	// clrTitle — bold panel title color.
	clrTitle = lipgloss.AdaptiveColor{Light: "5", Dark: "13"}
)

var (
	stylePanelBorder = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(clrBorder)

	stylePanelTitle = lipgloss.NewStyle().
				Bold(true).
				Foreground(clrTitle)

	styleActive   = lipgloss.NewStyle().Foreground(clrActive)
	styleIdle     = lipgloss.NewStyle().Foreground(clrMuted)
	styleDone     = lipgloss.NewStyle().Foreground(clrDone)
	styleLabel    = lipgloss.NewStyle().Foreground(clrMuted)
	styleValue    = lipgloss.NewStyle().Bold(true).Foreground(clrText)
	styleFooter   = lipgloss.NewStyle().Foreground(clrMuted)
	stylePrompt   = lipgloss.NewStyle().Bold(true).Foreground(clrActive)

	styleInputBar = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(clrBorder)
)
