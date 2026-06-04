package tui

import (
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"

	"github.com/cajohnson0125/Twirl/internal/engine"
)

// dims holds computed layout dimensions for the stacked
// three-panel TUI layout:
//
//	┌─ info bar (1 content row) ──────────────────────┐
//	│ ▸ Agent │ State │ Active                         │
//	└──────────────────────────────────────────────────┘
//	┌─ viewport (fills remaining space) ──────────────┐
//	│  output scrolls here                             │
//	└──────────────────────────────────────────────────┘
//	┌─ input bar (1 content row) ─────────────────────┐
//	│ > type here...                                   │
//	└──────────────────────────────────────────────────┘
//	  ctrl+c quit  •  enter send
type dims struct {
	w          int
	vpContentW int
	vpContentH int
	infoH      int
	inputH     int
	footerH    int
}

type agentStatus int

const (
	statusIdle agentStatus = iota
	statusActive
)

type agent struct {
	name   string
	status agentStatus
}

type cursorStyle struct {
	shape tea.CursorShape
	blink bool
}

// aiEntry maps a log index to its raw markdown text.
type aiEntry struct {
	logIndex int
	rawText  string
}

// model is the Bubbletea TUI model.
type model struct {
	eng *engine.Engine

	width  int
	height int
	ready  bool
	d      dims

	output viewport.Model
	input  textinput.Model
	spin   spinner.Model

	agents      []agent
	logs        []string
	phase       string
	engineState engine.State
	cs          cursorStyle

	mdRenderer *glamour.TermRenderer
	aiRaw      []aiEntry
}
