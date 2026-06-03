package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
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
	vpContentW int
	vpContentH int
	infoH      int
	inputH     int
	footerH    int
}

func computeDims(w, h int) dims {
	border := 2
	infoH := border + 1
	inputH := border + 1
	footerH := 0
	if h > 5 {
		footerH = 1
	}

	vpTotalH := h - infoH - inputH - footerH
	if vpTotalH < border+2 {
		vpTotalH = border + 2
	}

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

	// Markdown renderer for AI responses.
	mdRenderer *glamour.TermRenderer

	// Raw AI texts for re-rendering on resize.
	aiRaw []aiEntry
}

// aiEntry maps a log index to its raw markdown text.
type aiEntry struct {
	logIndex int
	rawText  string
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

// newRenderer creates a Glamour markdown renderer for the
// given content width and dark/light theme.
func newRenderer(width int) (*glamour.TermRenderer, error) {
	style := "dark"
	if !isDarkTheme {
		style = "light"
	}
	return glamour.NewTermRenderer(
		glamour.WithStandardStyle(style),
		glamour.WithWordWrap(width),
	)
}

func newModel(
	cursorShape string,
	cursorBlink bool,
) (model, error) {
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

	m := model{
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

	return m, nil
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
				return m.startStreaming(val)
			}

		case "up":
			m.output.ScrollUp(1)
			return m, nil
		case "down":
			m.output.ScrollDown(1)
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.d = computeDims(m.width, m.height)
		m.input.SetWidth(max(1, m.d.vpContentW-4))

		if r, err := newRenderer(
			max(20, m.d.vpContentW-6),
		); err == nil {
			m.mdRenderer = r
			m.rerenderAI()
		}

		if !m.ready {
			m.output = viewport.New(
				viewport.WithWidth(m.d.vpContentW),
				viewport.WithHeight(m.d.vpContentH),
			)
			m.syncOutput()
			m.ready = true
		} else {
			m.output.SetWidth(m.d.vpContentW)
			m.output.SetHeight(m.d.vpContentH)
			m.syncOutput()
		}
	}

	var spinCmd tea.Cmd
	m.spin, spinCmd = m.spin.Update(msg)
	cmds = append(cmds, spinCmd)

	var vpCmd tea.Cmd
	m.output, vpCmd = m.output.Update(msg)
	cmds = append(cmds, vpCmd)

	if m.ready {
		m.output.SetWidth(m.d.vpContentW)
		m.output.SetHeight(m.d.vpContentH)
	}

	var tiCmd tea.Cmd
	m.input, tiCmd = m.input.Update(msg)
	cmds = append(cmds, tiCmd)

	return m, tea.Batch(cmds...)
}

// startStreaming echoes input back as a placeholder until
// the Engine channel bus is wired in (Phase 1.3).
func (m model) startStreaming(
	prompt string,
) (tea.Model, tea.Cmd) {
	m.logs = append(m.logs,
		styleUser.Render("You: ")+prompt,
	)
	m.logs = append(m.logs,
		styleAI.Render("AI: ")+"I received your message: "+prompt,
	)
	m.logs = append(m.logs,
		styleDivider.Render(
			strings.Repeat("─", max(20, m.d.vpContentW)),
		),
	)
	m.syncOutput()
	m.output.GotoBottom()
	m.input.Reset()
	return m, nil
}

// renderMarkdown renders AI response text through the
// Glamour markdown renderer.
func (m model) renderMarkdown(text string) string {
	if m.mdRenderer == nil {
		return styleAI.Render("AI: ") + text
	}
	rendered, err := m.mdRenderer.Render(text)
	if err != nil {
		return styleAI.Render("AI: ") + text
	}
	rendered = strings.TrimSpace(rendered)
	return styleAI.Render("AI: ") + rendered
}

// rerenderAI re-renders all stored AI responses through the
// current Glamour renderer. Called on resize so text wraps
// to the new width.
func (m *model) rerenderAI() {
	if m.mdRenderer == nil {
		return
	}
	for _, entry := range m.aiRaw {
		if entry.logIndex < len(m.logs) {
			m.logs[entry.logIndex] = m.renderMarkdown(
				entry.rawText,
			)
		}
	}
}

// syncOutput rebuilds the viewport content from logs,
// including the in-flight streaming response if active.
// Lines containing ANSI escape codes (Glamour output)
// are passed through without re-wrapping.
func (m *model) syncOutput() {
	w := max(20, m.d.vpContentW)
	var lines []string
	for _, line := range m.logs {
		if strings.Contains(line, "\x1b[") {
			lines = append(lines, line)
		} else {
			lines = append(lines, wrapLine(line, w)...)
		}
	}

	m.output.SetContent(strings.Join(lines, "\n"))
}

// wrapLine breaks a string into lines of at most n visible
// characters, splitting at word boundaries when possible.
func wrapLine(s string, n int) []string {
	if lipgloss.Width(s) <= n {
		return []string{s}
	}

	var lines []string
	var cur strings.Builder
	for _, word := range strings.Split(s, " ") {
		if cur.Len() == 0 {
			cur.WriteString(word)
			continue
		}
		test := cur.String() + " " + word
		if lipgloss.Width(test) > n {
			lines = append(lines, cur.String())
			cur.Reset()
			cur.WriteString(word)
		} else {
			cur.WriteString(" ")
			cur.WriteString(word)
		}
	}
	if cur.Len() > 0 {
		lines = append(lines, cur.String())
	}
	return lines
}

// --- View ---

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

	c := m.input.Cursor()
	if c == nil {
		promptW := lipgloss.Width(m.input.Prompt)
		c = tea.NewCursor(m.input.Position()+promptW, 0)
		c.Shape = m.cs.shape
		c.Blink = m.cs.blink
	}
	c.Position.Y += m.d.infoH + (m.d.vpContentH + 3) + 1
	c.Position.X += 1
	v.Cursor = c

	return v
}

func (m model) viewInfoBar() string {
	active := ""
	for _, a := range m.agents {
		if a.status == statusActive {
			active = a.name
			break
		}
	}

	activeStyle := styleActive.Bold(true).Render(
		"▸ " + active)
	phase := styleLabel.Render("Phase:") + " " +
		styleValue.Render(m.phase)
	count := styleLabel.Render("Active:") + " " +
		styleValue.Render(
			fmt.Sprintf("%d/%d",
				countActive(m.agents), len(m.agents)),
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
		Height(m.d.vpContentH + 3).
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
func Run(
	cursorShape string,
	cursorBlink bool,
) error {
	m, err := newModel(cursorShape, cursorBlink)
	if err != nil {
		return err
	}
	p := tea.NewProgram(m)
	_, err = p.Run()
	return err
}
