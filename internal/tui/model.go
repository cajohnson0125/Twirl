package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	// Side panels scale to this fraction of terminal width.
	leftRatio  = 0.18
	rightRatio = 0.18

	// Upper bound for side panel content width.
	maxPanelW = 32

	// Width thresholds for progressive panel collapse.
	// Below showBoth, only the left panel is shown.
	// Below showLeft, only the center panel is shown.
	showBoth = 70
	showLeft = 50
)

// dims holds all computed layout dimensions derived from terminal size.
// Zero width for a side panel means that panel is not rendered.
type dims struct {
	leftW       int // content width of agents panel (0 = hidden)
	rightW      int // content width of status panel (0 = hidden)
	centerW     int // content width of output panel (always >= 1)
	panelInnerH int // content height of all panels (always >= 1)
	vpH         int // viewport height inside the center panel
}

// computeDims derives all panel dimensions from terminal width/height.
//
// Layout is progressive — side panels collapse as the terminal narrows:
//
//	width >= 70:  three panels (left + center + right)
//	width >= 50:  two panels   (left + center)
//	width <  50:  one panel    (center only, full width)
//
// This ensures the layout works at any terminal size, including tiling
// window manager splits. All values are clamped to >= 0.
func computeDims(w, h int) dims {
	var leftW, rightW int

	switch {
	case w >= showBoth:
		leftW = clamp(int(float64(w)*leftRatio), 1, maxPanelW)
		rightW = clamp(int(float64(w)*rightRatio), 1, maxPanelW)
	case w >= showLeft:
		leftW = clamp(int(float64(w)*leftRatio), 1, maxPanelW)
		rightW = 0
	default:
		leftW = 0
		rightW = 0
	}

	// Account for side panel borders (2 chars each for rendered panels).
	sideBorders := 0
	if leftW > 0 {
		sideBorders += 2
	}
	if rightW > 0 {
		sideBorders += 2
	}

	// Center content = total - side panels - side borders - center borders.
	centerW := w - leftW - rightW - sideBorders - 2
	if centerW < 1 {
		centerW = 1
	}

	// Panels occupy all vertical space except: input(1) + footer(1) + borders(2).
	panelInnerH := h - 4
	if panelInnerH < 1 {
		panelInnerH = 1
	}

	// Viewport sits inside the center panel below: title(1) + blank line(1).
	vpH := panelInnerH - 2
	if vpH < 1 {
		vpH = 1
	}

	return dims{leftW, rightW, centerW, panelInnerH, vpH}
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
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
		m.input.Width = clamp(m.width-4, 1, m.width)

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

	// Build the panel row — only include panels with width > 0.
	var panels []string
	if m.d.leftW > 0 {
		panels = append(panels, m.viewAgents())
	}
	panels = append(panels, m.viewOutput())
	if m.d.rightW > 0 {
		panels = append(panels, m.viewStatus())
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, panels...)

	footer := styleFooter.Width(m.width).Render(
		"ctrl+c quit  •  enter send  •  ↑↓ scroll",
	)

	return row + "\n" + m.viewInput() + "\n" + footer
}

func (m model) viewAgents() string {
	var sb strings.Builder
	sb.WriteString(stylePanelTitle.Render("AGENTS") + "\n\n")

	for _, a := range m.agents {
		var line string
		switch a.status {
		case statusActive:
			line = fmt.Sprintf("%s %s", m.spin.View(), styleActive.Render(a.name))
		case statusDone:
			line = fmt.Sprintf("%s %s", styleDone.Render("✓"), styleDone.Render(a.name))
		default:
			line = fmt.Sprintf("%s %s", styleIdle.Render("○"), styleIdle.Render(a.name))
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
	sb.WriteString(stylePanelTitle.Render("STATUS") + "\n\n")
	sb.WriteString(styleLabel.Render("Phase") + "\n")
	sb.WriteString(styleValue.Render(m.phase) + "\n\n")
	sb.WriteString(styleLabel.Render("Step") + "\n")
	sb.WriteString(styleValue.Render(fmt.Sprintf("%d / %d", m.step, m.totalSteps)) + "\n")

	return stylePanelBorder.
		Width(m.d.rightW).
		Height(m.d.panelInnerH).
		Render(sb.String())
}

func (m model) viewInput() string {
	return fmt.Sprintf("%s %s", stylePrompt.Render(">"), m.input.View())
}

// Run starts the TUI program.
func Run() error {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
