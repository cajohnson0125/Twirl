package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Stacked layout:
//
//	┌─ info bar (1 content row) ──────────────────────┐
//	│ Brainstorm ▸ │ Idle │ 1/10 active               │
//	└──────────────────────────────────────────────────┘
//	┌─ viewport (fills remaining space) ──────────────┐
//	│                                                  │
//	│  output scrolls here                             │
//	│                                                  │
//	└──────────────────────────────────────────────────┘
//	┌─ input bar (1 content row) ─────────────────────┐
//	│ > type here...                                   │
//	└──────────────────────────────────────────────────┘
//	  ctrl+c quit  •  enter send

type dims struct {
	w          int
	vpContentW int // viewport content width
	vpContentH int // viewport content height
	infoH      int // info bar total (with border)
	inputH     int // input bar total (with border)
	footerH    int // footer height (0 or 1)
}

func computeDims(w, h int) dims {
	border := 2
	infoH := border + 1 // 1 content row + borders
	inputH := border + 1
	footerH := 0
	if h > 5 {
		footerH = 1
	}

	vpTotalH := h - infoH - inputH - footerH
	if vpTotalH < border+2 { // border + title + at least 1 vp line
		vpTotalH = border + 2
	}

	// Content area minus title row = viewport.Model height.
	// The rendered block is vpTotalH lines total:
	//   2 border + 1 title + vpContentH viewport lines.
	vpContentH := vpTotalH - border - 1

	return dims{
		w:          w,
		vpContentW: w - border,
		vpContentH: vpContentH,
		infoH:      infoH,
		inputH:     inputH,
		footerH:    footerH,
	}
}

// --- Agent types ---

type agentStatus int

const (
	statusIdle agentStatus = iota
	statusActive
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

	agents []agent
	logs   []string
	phase  string
	cs     cursorStyle
}

type cursorStyle struct {
	shape tea.CursorShape
	blink bool
}

func parseCursorShape(s string) tea.CursorShape {
	switch s {
	case "underline":
		return tea.CursorUnderline
	case "bar":
		return tea.CursorBar
	default:
		return tea.CursorBlock
	}
}

func newModel(cursorShape string, cursorBlink bool) model {
	shape := parseCursorShape(cursorShape)
	cs := cursorStyle{shape: shape, blink: cursorBlink}

	ti := textinput.New()
	ti.Prompt = "> "
	ti.Placeholder = "Type a message..."
	ti.CharLimit = 256
	ti.SetVirtualCursor(false)
	ti.Focus()

	styles := textinput.DefaultDarkStyles()
	styles.Focused.Prompt = lipgloss.NewStyle().
		Foreground(clrActive).Bold(true)
	styles.Cursor = textinput.CursorStyle{
		Shape: shape,
		Blink: cursorBlink,
	}
	ti.SetStyles(styles)

	sp := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(
			lipgloss.NewStyle().Foreground(clrAccent),
		),
	)

	return model{
		cs:    cs,
		input: ti,
		spin:  sp,
		phase: "Idle",
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
		logs: []string{"Twirl ready"},
	}
}

func (m model) Init() tea.Cmd {
	return m.spin.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, func() tea.Msg { return tea.QuitMsg{} }
		case "enter":
			if val := m.input.Value(); val != "" {
				m.logs = append(m.logs, "> "+val)
				m.output.SetContent(
					strings.Join(m.logs, "\n"),
				)
				m.output.GotoBottom()
				m.input.Reset()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.d = computeDims(m.width, m.height)
		m.input.SetWidth(max(1, m.d.vpContentW-4))

		if !m.ready {
			m.output = viewport.New(
				viewport.WithWidth(m.d.vpContentW),
				viewport.WithHeight(m.d.vpContentH),
			)
			m.output.SetContent(
				strings.Join(m.logs, "\n"),
			)
			m.ready = true
		} else {
			m.output.SetWidth(m.d.vpContentW)
			m.output.SetHeight(m.d.vpContentH)
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

func (m model) View() tea.View {
	if !m.ready {
		return tea.NewView("\n  Loading...")
	}

	var b strings.Builder
	b.WriteString(m.viewInfoBar())
	b.WriteString("\n")
	b.WriteString(m.viewViewport())
	b.WriteString("\n")
	b.WriteString(m.viewInputBar())
	if m.d.footerH > 0 {
		b.WriteString("\n" + m.viewFooter())
	}

	v := tea.NewView(b.String())
	v.AltScreen = true

	// Position the native cursor at the textinput.
	c := m.input.Cursor()
	if c == nil {
		promptW := lipgloss.Width(m.input.Prompt)
		c = tea.NewCursor(m.input.Position()+promptW, 0)
		c.Shape = m.cs.shape
		c.Blink = m.cs.blink
	}
	// Info bar (3 lines) + viewport block (vpContentH + 3)
	// + input top border (1 line) = input content row.
	c.Position.Y += m.d.infoH + (m.d.vpContentH + 3) + 1
	c.Position.X += 1
	v.Cursor = c

	return v
}

// viewInfoBar renders a single-row status bar showing the
// active agent, phase, and agent count.
func (m model) viewInfoBar() string {
	active := ""
	for _, a := range m.agents {
		if a.status == statusActive {
			active = a.name
			break
		}
	}

	activeStyle := styleActive.Bold(true).Render("▸ " + active)
	phase := styleLabel.Render("Phase:") + " " +
		styleValue.Render(m.phase)
	count := styleLabel.Render("Active:") + " " +
		styleValue.Render(
			fmt.Sprintf("%d/%d", countActive(m.agents), len(m.agents)),
		)

	content := activeStyle + "  " + phase + "  " + count
	return stylePanelBorder.
		Width(m.d.w).
		Render(content)
}

func countActive(agents []agent) int {
	n := 0
	for _, a := range agents {
		if a.status == statusActive {
			n++
		}
	}
	return n
}

func (m model) viewViewport() string {
	content := stylePanelTitle.Render("OUTPUT") +
		"\n" + m.output.View()
	return stylePanelBorder.
		Width(m.d.w).
		Height(m.d.vpContentH + 3). // border + title + content + border
		Render(content)
}

func (m model) viewInputBar() string {
	inputContent := m.input.View()
	return styleInputBar.
		Width(m.d.w).
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

// Run starts the TUI program with the given cursor config.
func Run(cursorShape string, cursorBlink bool) error {
	p := tea.NewProgram(newModel(cursorShape, cursorBlink))
	_, err := p.Run()
	return err
}
