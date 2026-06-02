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
	"github.com/cajohnson0125/Twirl/internal/config"
	"github.com/cajohnson0125/Twirl/internal/pubsub"
	"github.com/cajohnson0125/Twirl/internal/state"
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

// --- Message types ---

// busEvent wraps a pubsub.Event for the Bubbletea event loop.
type busEvent struct{ event pubsub.Event }

// engineDoneMsg signals the engine has finished.
type engineDoneMsg struct{ err error }

// --- Agent tracking ---

type agentStatus int

const (
	agentIdle agentStatus = iota
	agentActive
	agentDone
)

type agentInfo struct {
	name   string
	status agentStatus
}

// --- HITL gate state ---

type gateState struct {
	prompt  string
	options []string
	active  bool
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

	agents    map[string]agentInfo
	agentList []string // ordered for display
	logs      []string
	phase     string
	cs        cursorStyle

	// Engine integration.
	bus     *pubsub.Bus
	hitlOut chan<- state.HITLResponse
	engine  engineController

	// Bus subscriptions (created once in newModel).
	busCh   <-chan pubsub.Event
	busDone chan struct{}

	// Engine completion.
	engineDone <-chan error

	// HITL gate state.
	gate gateState

	// Streaming state (tokens from EventStream).
	responseBuf strings.Builder
	streaming   bool

	// Markdown renderer.
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

// engineController abstracts engine control for the TUI.
// The real implementation cancels the engine context.
type engineController interface {
	Cancel()
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

// Opts holds the dependencies the TUI needs from the
// orchestration layer.
type Opts struct {
	CursorShape string
	CursorBlink bool
	LLM         config.LLM
	Bus         *pubsub.Bus
	HITLOut     chan<- state.HITLResponse
	Engine      engineController
	EngineDone  <-chan error
}

func newModel(opts Opts) (model, error) {
	shape := parseCursorShape(opts.CursorShape)
	cs := cursorStyle{shape: shape, blink: opts.CursorBlink}

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
		Blink: opts.CursorBlink,
	}
	ti.SetStyles(styles)

	sp := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(
			lipgloss.NewStyle().Foreground(clrAccent),
		),
	)

	// Build initial agent tracking from the 10 roles.
	agentNames := []string{
		"Brainstorm", "Research", "Report",
		"Plan", "Plan Review", "Execution",
		"Code Review", "Triage", "Assessment", "Scribe",
	}
	agents := make(map[string]agentInfo, len(agentNames))
	for _, n := range agentNames {
		agents[n] = agentInfo{name: n, status: agentIdle}
	}

	m := model{
		cs:        cs,
		input:     ti,
		spin:      sp,
		phase:     "Idle",
		agents:    agents,
		agentList: agentNames,
		logs:      []string{"Twirl ready"},
		bus:       opts.Bus,
		hitlOut:   opts.HITLOut,
		engine:    opts.Engine,
	}

	if opts.Bus != nil {
		m.mergeBusSubscriptions(opts)
	}
	if opts.EngineDone != nil {
		m.engineDone = opts.EngineDone
	}

	if !opts.LLM.IsZero() {
		m.logs = append(m.logs,
			styleLabel.Render(
				"LLM: "+opts.LLM.Model+" @ "+opts.LLM.BaseURL,
			),
		)
	} else {
		m.logs = append(m.logs,
			styleLabel.Render(
				"No LLM configured. " +
					"Edit ~/.config/twirl/config.toml",
			),
		)
	}

	return m, nil
}

func (m model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.spin.Tick}
	if m.busCh != nil {
		cmds = append(cmds, m.waitForBusEvent())
	}
	if m.engineDone != nil {
		cmds = append(cmds, m.waitForEngineDone())
	}
	return tea.Batch(cmds...)
}

// mergeBusSubscriptions creates a single merged channel
// from all 5 event-type subscriptions. Called once in
// newModel to avoid leaking channels on every event.
func (m *model) mergeBusSubscriptions(opts Opts) {
	types := []pubsub.EventType{
		pubsub.EventStream,
		pubsub.EventAgentStarted,
		pubsub.EventAgentDone,
		pubsub.EventGate,
		pubsub.EventError,
	}

	merged := make(chan pubsub.Event, opts.Bus.BufferCap())
	done := make(chan struct{})

	for _, typ := range types {
		ch := opts.Bus.Subscribe(typ)
		go func(c <-chan pubsub.Event) {
			for {
				select {
				case <-done:
					return
				case ev, ok := <-c:
					if !ok {
						return
					}
					select {
					case merged <- ev:
					default:
					}
				}
			}
		}(ch)
	}

	m.busCh = merged
	m.busDone = done
}

// waitForBusEvent returns a Cmd that blocks until the next
// event arrives on the merged bus channel.
func (m model) waitForBusEvent() tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-m.busCh
		if !ok {
			return nil
		}
		return busEvent{ev}
	}
}

// waitForEngineDone returns a Cmd that blocks until the
// engine signals completion.
func (m model) waitForEngineDone() tea.Cmd {
	return func() tea.Msg {
		err, ok := <-m.engineDone
		if !ok {
			return engineDoneMsg{nil}
		}
		return engineDoneMsg{err}
	}
}

func (m model) Update(
	msg tea.Msg,
) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m.handleKey(msg)

	case busEvent:
		cmd := m.handleBusEvent(msg.event)
		cmds = append(cmds, cmd)

	case engineDoneMsg:
		m.handleEngineDone(msg.err)

	case tea.WindowSizeMsg:
		m.handleResize(msg)
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

func (m model) handleKey(
	msg tea.KeyPressMsg,
) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		if m.gate.active {
			// Ignore ctrl+c during HITL gate.
			return m, nil
		}
		if m.engine != nil {
			m.engine.Cancel()
		}
		return m, func() tea.Msg { return tea.QuitMsg{} }

	case "enter":
		if m.gate.active {
			return m.handleGateInput()
		}
		if m.streaming {
			return m, nil
		}
		return m, nil

	case "up":
		m.output.ScrollUp(1)
		return m, nil
	case "down":
		m.output.ScrollDown(1)
		return m, nil
	}
	return m, nil
}

// handleGateInput processes user input when a HITL gate
// is active. The user types a number (1-based) and presses
// enter to select an option.
func (m model) handleGateInput() (
	tea.Model,
	tea.Cmd,
) {
	val := m.input.Value()
	if val == "" || m.hitlOut == nil {
		return m, nil
	}

	// Determine choice: accept number or exact text.
	choice := val
	var idx int
	if fmt.Sscanf(val, "%d", &idx); idx >= 1 && idx <= len(m.gate.options) {
		choice = m.gate.options[idx-1]
	}

	select {
	case m.hitlOut <- state.HITLResponse{
		ID:     m.gate.prompt,
		Choice: choice,
		Input:  val,
	}:
	default:
		m.logs = append(m.logs,
			styleError.Render(
				"Gate response dropped (channel full)"))
	}

	m.logs = append(m.logs,
		styleUser.Render("You: ")+val,
	)
	m.gate.active = false
	m.input.Reset()
	m.input.Placeholder = "Type a message..."
	m.syncOutput()
	m.output.GotoBottom()
	return m, nil
}

// handleBusEvent dispatches engine events to the
// appropriate handler.
func (m *model) handleBusEvent(
	ev pubsub.Event,
) tea.Cmd {
	switch ev.Type {
	case pubsub.EventStream:
		m.responseBuf.WriteString(ev.Token)
		m.streaming = true
		m.phase = "Streaming ▸"
		m.syncOutput()
		m.output.GotoBottom()

	case pubsub.EventAgentStarted:
		name := roleDisplayName(ev.Role)
		if info, ok := m.agents[name]; ok {
			info.status = agentActive
			m.agents[name] = info
		}
		m.phase = name + " ▸"
		m.responseBuf.Reset()
		m.streaming = true
		m.logs = append(m.logs,
			styleLabel.Render(
				"▸ "+name+" started",
			),
		)
		m.syncOutput()
		m.output.GotoBottom()

	case pubsub.EventAgentDone:
		name := roleDisplayName(ev.Role)
		if info, ok := m.agents[name]; ok {
			info.status = agentDone
			m.agents[name] = info
		}
		if m.streaming {
			raw := m.responseBuf.String()
			if raw != "" {
				idx := len(m.logs)
				m.logs = append(m.logs,
					m.renderMarkdown(raw),
				)
				m.aiRaw = append(m.aiRaw, aiEntry{
					logIndex: idx,
					rawText:  raw,
				})
			}
			m.responseBuf.Reset()
			m.streaming = false

			m.logs = append(m.logs,
				styleDivider.Render(
					strings.Repeat(
						"─",
						max(20, m.d.vpContentW),
					),
				),
			)
		}
		m.syncOutput()
		m.output.GotoBottom()

	case pubsub.EventGate:
		m.gate = gateState{
			prompt:  ev.Prompt,
			options: ev.Options,
			active:  true,
		}
		m.streaming = false
		m.responseBuf.Reset()

		var lines []string
		lines = append(lines,
			styleAccent.Render(
				"⚡ HITL Gate: "+ev.Prompt,
			),
		)
		for i, opt := range ev.Options {
			lines = append(lines,
				fmt.Sprintf("  %d. %s", i+1, opt),
			)
		}
		lines = append(lines,
			styleMuted.Render(
				"  Type a number and press Enter.",
			),
		)
		m.logs = append(m.logs, lines...)
		m.input.Reset()
		m.input.Placeholder = "Select option..."
		m.input.Focus()
		m.syncOutput()
		m.output.GotoBottom()

	case pubsub.EventError:
		m.logs = append(m.logs,
			styleError.Render("Error: "+ev.Err),
		)
		m.streaming = false
		m.phase = "Error"
		m.syncOutput()
		m.output.GotoBottom()
	}

	// Continue listening for events.
	if m.busCh != nil {
		return m.waitForBusEvent()
	}
	return nil
}

func (m *model) handleEngineDone(err error) {
	if err != nil {
		m.logs = append(m.logs,
			styleError.Render(
				"Engine error: "+err.Error(),
			),
		)
		m.phase = "Error"
	} else {
		m.logs = append(m.logs,
			styleActive.Render("✓ Workflow complete"),
		)
		m.phase = "Done"
	}
	m.streaming = false
	m.syncOutput()
	m.output.GotoBottom()
}

func (m *model) handleResize(msg tea.WindowSizeMsg) {
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

	if m.streaming && m.responseBuf.Len() > 0 {
		lines = append(lines,
			wrapLine(
				styleAI.Render("AI: ")+
					m.responseBuf.String(),
				w,
			)...,
		)
	}

	m.output.SetContent(strings.Join(lines, "\n"))
}

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

// roleDisplayName converts a role string to a display name.
func roleDisplayName(role string) string {
	switch role {
	case "brainstorm":
		return "Brainstorm"
	case "research":
		return "Research"
	case "report":
		return "Report"
	case "plan":
		return "Plan"
	case "plan_review":
		return "Plan Review"
	case "execution":
		return "Execution"
	case "code_review":
		return "Code Review"
	case "triage":
		return "Triage"
	case "assessment":
		return "Assessment"
	case "scribe":
		return "Scribe"
	case "hitl_gate":
		return "HITL Gate"
	default:
		return role
	}
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
	var activeNames []string
	for _, name := range m.agentList {
		if m.agents[name].status == agentActive {
			activeNames = append(activeNames, name)
		}
	}

	activeCount := len(activeNames)
	var activeLabel string
	switch {
	case activeCount == 0:
		activeLabel = "Idle"
	case activeCount == 1:
		activeLabel = activeNames[0]
	case activeCount <= 3:
		activeLabel = strings.Join(activeNames, ", ")
	default:
		activeLabel = fmt.Sprintf("%d agents", activeCount)
	}

	activeStyle := styleActive.Bold(true).Render(
		"▸ " + activeLabel)

	phase := styleLabel.Render("Phase:") + " " +
		styleValue.Render(m.phase)
	count := styleLabel.Render("Active:") + " " +
		styleValue.Render(
			fmt.Sprintf("%d/%d",
				activeCount, len(m.agentList)),
		)

	content := activeStyle + "  " + phase + "  " + count
	return stylePanelBorder.
		Width(m.d.w).
		Render(content)
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

// Run starts the TUI program with the given options.
func Run(opts Opts) error {
	m, err := newModel(opts)
	if err != nil {
		return err
	}
	p := tea.NewProgram(m)
	_, err = p.Run()
	if m.busDone != nil {
		close(m.busDone)
	}
	return err
}
