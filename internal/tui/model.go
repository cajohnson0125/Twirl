package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

const (
	// Side panels scale to this fraction of terminal width.
	leftRatio  = 0.18
	rightRatio = 0.18

	// Upper bound for side panel content width.
	maxPanelW = 32
)

// dims holds all computed layout dimensions derived from terminal size.
// All three panels always render — both width and height shrink smoothly
// and proportionally as the terminal shrinks.
type dims struct {
	leftW       int  // content width of agents panel
	rightW      int  // content width of status panel
	centerW     int  // content width of output panel
	panelInnerH int  // content height of panels
	vpH         int  // viewport height inside center panel
	showFooter  bool // false when height is tight
	inputBarH   int  // height of the input bar (1 content + 2 border)
}

// computeDims derives all panel dimensions from terminal width/height.
//
// Width:  side panels scale at 18% ratio, center takes the rest.
// Height: footer hides below 6 rows. Input bar is a bordered panel.
//         Title takes 1 row inside each panel — no blank lines.
//         Everything scales continuously.
func computeDims(w, h int) dims {
	leftW := clamp(int(float64(w)*leftRatio), 1, maxPanelW)
	rightW := clamp(int(float64(w)*rightRatio), 1, maxPanelW)

	centerW := w - leftW - rightW - 6
	if centerW < 1 {
		deficit := 1 - centerW
		leftW = max(1, leftW-deficit/2)
		rightW = max(1, rightW-(deficit-deficit/2))
		centerW = 1
	}

	// Vertical layout (from top to bottom):
	//   panel row  = panelInnerH + 2 (borders)
	//   input bar  = 1 + 2 (content + borders), or 1 (no border at tiny heights)
	//   footer     = 1 (optional)
	//
	//   bordered input + footer: panelInnerH + 2 + 3 + 1 = panelInnerH + 6
	//   bordered input only:     panelInnerH + 2 + 3     = panelInnerH + 5
	//   plain input + footer:    panelInnerH + 2 + 1 + 1 = panelInnerH + 4
	//   plain input only:        panelInnerH + 2 + 1     = panelInnerH + 3

	borderedInput := h > 5
	showFooter := h > 7

	var panelInnerH int
	switch {
	case showFooter && borderedInput:
		panelInnerH = max(1, h-6)
	case borderedInput:
		panelInnerH = max(1, h-5)
	case showFooter:
		panelInnerH = max(1, h-4)
	default:
		panelInnerH = max(1, h-3)
	}

	inputBarH := 3 // bordered
	if !borderedInput {
		inputBarH = 1 // plain
	}

	vpH := max(1, panelInnerH-1)

	return dims{leftW, rightW, centerW, panelInnerH, vpH, showFooter, inputBarH}
}

func clamp(v, lo, hi int) int {
	return max(lo, min(v, hi))
}

// trunc clips s to n visible terminal columns, appending "…" when
// truncated. Returns "" for n <= 0.
func trunc(s string, n int) string {
	if n <= 0 {
		return ""
	}
	return runewidth.Truncate(s, n, "…")
}

// --- Agent types ---

type agentStatus int

const (
	statusIdle agentStatus = iota
	statusActive
	statusDone
)

type agent struct {
	name   string
	status agentStatus
}

// --- Model ---

type model struct {
	width  int
	height int
	ready  bool
	d      dims

	output viewport.Model
	input  textinput.Model
	spin   spinner.Model

	agents     []agent
	logs       []string
	step       int
	totalSteps int
	phase      string
}

func newModel() model {
	ti := textinput.New()
	ti.Placeholder = "Type a message..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Cursor.SetMode(cursor.CursorHide) // use terminal's native cursor

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(clrAccent)

	return model{
		input:      ti,
		spin:       sp,
		step:       0,
		totalSteps: 10,
		phase:      "Initializing",
		agents: []agent{
			{name: "Brainstorm", status: statusActive},
			{name: "Research", status: statusIdle},
			{name: "Report", status: statusIdle},
			{name: "Plan", status: statusIdle},
			{name: "Plan Review", status: statusIdle},
			{name: "Execution", status: statusIdle},
			{name: "Code Review", status: statusIdle},
			{name: "Triage", status: statusIdle},
			{name: "Assessment", status: statusIdle},
			{name: "Scribe", status: statusIdle},
		},
		logs: []string{
			"⟳ Twirl orchestrator starting...",
			"✓ Agents registered",
			"✓ State loaded",
			"→ Awaiting task",
		},
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spin.Tick, textinput.Blink, tea.ShowCursor)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			if val := m.input.Value(); val != "" {
				m.logs = append(m.logs, "> "+val)
				m.output.SetContent(strings.Join(m.logs, "\n"))
				m.output.GotoBottom()
				m.input.Reset()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.d = computeDims(m.width, m.height)
		m.input.Width = max(1, m.d.centerW-4)

		if !m.ready {
			m.output = viewport.New(m.d.centerW, m.d.vpH)
			m.output.SetContent(strings.Join(m.logs, "\n"))
			m.ready = true
		} else {
			m.output.Width = m.d.centerW
			m.output.Height = m.d.vpH
		}
	}

	var spinCmd tea.Cmd
	m.spin, spinCmd = m.spin.Update(msg)
	cmds = append(cmds, spinCmd)

	var vpCmd tea.Cmd
	m.output, vpCmd = m.output.Update(msg)
	cmds = append(cmds, vpCmd)

	var tiCmd tea.Cmd
	m.input, tiCmd = m.input.Update(msg)
	cmds = append(cmds, tiCmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "\n  Loading..."
	}

	panels := lipgloss.JoinHorizontal(lipgloss.Top,
		m.viewAgents(),
		m.viewOutput(),
		m.viewStatus(),
	)

	out := panels + "\n" + m.viewInputBar()
	if m.d.showFooter {
		out += "\n" + m.viewFooter()
	}
	return out
}

// noWrap forces a single line — no word wrapping.
var noWrap = lipgloss.NewStyle().Inline(true)

func (m model) viewAgents() string {
	var sb strings.Builder
	sb.WriteString(stylePanelTitle.Render(
		trunc("DISPATCH", m.d.leftW),
	) + "\n")

	// Title takes 1 row. Only render agents that fit.
	nameW := max(1, m.d.leftW-2)
	maxVisible := max(0, m.d.panelInnerH-1)

	for i, a := range m.agents {
		if i >= maxVisible {
			break
		}
		name := trunc(a.name, nameW)
		var line string
		switch a.status {
		case statusActive:
			line = fmt.Sprintf("%s %s",
				m.spin.View(), styleActive.Render(name))
		case statusDone:
			line = fmt.Sprintf("%s %s",
				styleDone.Render("✓"), styleDone.Render(name))
		default:
			line = fmt.Sprintf("%s %s",
				styleIdle.Render("○"), styleIdle.Render(name))
		}
		sb.WriteString(noWrap.Render(line) + "\n")
	}

	return stylePanelBorder.
		Width(m.d.leftW).
		Height(m.d.panelInnerH).
		Render(sb.String())
}

func (m model) viewOutput() string {
	content := stylePanelTitle.Render("OUTPUT") + "\n" + m.output.View()

	return stylePanelBorder.
		Width(m.d.centerW).
		Height(m.d.panelInnerH).
		Render(content)
}

func (m model) viewStatus() string {
	var sb strings.Builder
	sb.WriteString(stylePanelTitle.Render(
		trunc("STATUS", m.d.rightW),
	) + "\n")

	// Fill available rows, most important content first.
	avail := m.d.panelInnerH - 1

	// Orchestrator online indicator — always shown if possible.
	if avail > 0 {
		sb.WriteString(styleActive.Render(
			trunc("● Online", m.d.rightW),
		) + "\n")
		avail--
	}
	if avail > 0 {
		sb.WriteString(styleLabel.Render(
			trunc("Phase", m.d.rightW),
		) + "\n")
		avail--
	}
	if avail > 0 {
		sb.WriteString(styleValue.Render(
			trunc(m.phase, m.d.rightW),
		) + "\n")
		avail--
	}
	if avail > 1 {
		sb.WriteString(styleLabel.Render(
			trunc("Step", m.d.rightW),
		) + "\n")
		sb.WriteString(styleValue.Render(
			trunc(fmt.Sprintf("%d / %d", m.step, m.totalSteps), m.d.rightW),
		) + "\n")
	}

	return stylePanelBorder.
		Width(m.d.rightW).
		Height(m.d.panelInnerH).
		Render(sb.String())
}

// viewInputBar renders the input area — bordered panel at normal heights,
// plain inline at very small heights.
func (m model) viewInputBar() string {
	inputContent := stylePrompt.Render(">") + " " + m.input.View()
	if m.d.inputBarH <= 1 {
		return inputContent
	}
	return styleInputBar.
		Width(m.width - 2).
		Render(inputContent)
}

func (m model) viewFooter() string {
	var parts []string
	parts = append(parts, "ctrl+c quit")
	if m.width > 30 {
		parts = append(parts, "enter send")
	}
	if m.width > 50 {
		parts = append(parts, "↑↓ scroll")
	}
	return styleFooter.Width(m.width).Render(
		strings.Join(parts, "  •  "),
	)
}

// Run starts the TUI program.
func Run() error {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
