package tui

import (
	"fmt"
	"strings"

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
// All three panels always render — widths shrink smoothly as the
// terminal narrows. No panel dropping, no size thresholds.
type dims struct {
	leftW       int // content width of agents panel (always >= 1)
	rightW      int // content width of status panel (always >= 1)
	centerW     int // content width of output panel (always >= 1)
	panelInnerH int // content height of all panels
	vpH         int // viewport height inside the center panel
}

// computeDims derives panel dimensions from terminal width/height.
//
// All three panels always render. Side panel widths scale proportionally
// with terminal width. Center panel takes all remaining space.
// No thresholds, no panel dropping — continuous smooth scaling.
func computeDims(w, h int) dims {
	leftW := clamp(int(float64(w)*leftRatio), 1, maxPanelW)
	rightW := clamp(int(float64(w)*rightRatio), 1, maxPanelW)

	// 6 chars consumed by borders (2 per panel).
	centerW := w - leftW - rightW - 6

	// If center got squeezed, steal from side panels proportionally.
	if centerW < 1 {
		deficit := 1 - centerW
		leftW = max(1, leftW-deficit/2)
		rightW = max(1, rightW-(deficit-deficit/2))
		centerW = 1
	}

	panelInnerH := max(1, h-4)
	vpH := max(1, panelInnerH-2)

	return dims{leftW, rightW, centerW, panelInnerH, vpH}
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

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(clrAccent)

	return model{
		input:      ti,
		spin:       sp,
		step:       0,
		totalSteps: 7,
		phase:      "Initializing",
		agents: []agent{
			{name: "Orchestrator", status: statusActive},
			{name: "Brainstormer", status: statusIdle},
			{name: "Researcher", status: statusIdle},
			{name: "Planner", status: statusIdle},
			{name: "Coder", status: statusIdle},
			{name: "Reviewer", status: statusIdle},
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
	return tea.Batch(m.spin.Tick, textinput.Blink)
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
		m.input.Width = max(1, m.width-4)

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

	return panels + "\n" + m.viewInput() + "\n" + m.viewFooter()
}

func (m model) viewAgents() string {
	var sb strings.Builder
	sb.WriteString(stylePanelTitle.Render(trunc("AGENTS", m.d.leftW)) + "\n\n")

	// Indicator char + space = ~2 cols. Name gets the rest.
	nameW := max(1, m.d.leftW-2)

	for _, a := range m.agents {
		name := trunc(a.name, nameW)
		var line string
		switch a.status {
		case statusActive:
			line = fmt.Sprintf("%s %s", m.spin.View(), styleActive.Render(name))
		case statusDone:
			line = fmt.Sprintf("%s %s", styleDone.Render("✓"), styleDone.Render(name))
		default:
			line = fmt.Sprintf("%s %s", styleIdle.Render("○"), styleIdle.Render(name))
		}
		sb.WriteString(line + "\n")
	}

	return stylePanelBorder.
		Width(m.d.leftW).
		Height(m.d.panelInnerH).
		Render(sb.String())
}

func (m model) viewOutput() string {
	content := stylePanelTitle.Render("OUTPUT") + "\n\n" + m.output.View()

	return stylePanelBorder.
		Width(m.d.centerW).
		Height(m.d.panelInnerH).
		Render(content)
}

func (m model) viewStatus() string {
	var sb strings.Builder
	sb.WriteString(stylePanelTitle.Render(trunc("STATUS", m.d.rightW)) + "\n\n")
	sb.WriteString(styleLabel.Render(trunc("Phase", m.d.rightW)) + "\n")
	sb.WriteString(styleValue.Render(trunc(m.phase, m.d.rightW)) + "\n\n")
	sb.WriteString(styleLabel.Render(trunc("Step", m.d.rightW)) + "\n")
	sb.WriteString(styleValue.Render(
		trunc(fmt.Sprintf("%d / %d", m.step, m.totalSteps), m.d.rightW),
	) + "\n")

	return stylePanelBorder.
		Width(m.d.rightW).
		Height(m.d.panelInnerH).
		Render(sb.String())
}

func (m model) viewInput() string {
	return fmt.Sprintf("%s %s", stylePrompt.Render(">"), m.input.View())
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
