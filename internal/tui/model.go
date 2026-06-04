package tui

import (
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/cajohnson0125/Twirl/internal/engine"
)

func newModel(
	eng *engine.Engine,
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
		eng:         eng,
		cs:          cs,
		input:       ti,
		spin:        sp,
		phase:       "Idle",
		engineState: engine.StateCoordinator,
		agents: []agent{
			{name: "Coordinator", status: statusIdle},
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

// Run starts the TUI program with the given cursor config.
func Run(
	eng *engine.Engine,
	cursorShape string,
	cursorBlink bool,
) error {
	m, err := newModel(eng, cursorShape, cursorBlink)
	if err != nil {
		return err
	}
	p := tea.NewProgram(m)
	_, err = p.Run()
	return err
}

// waitForEngineMsg returns a Cmd that blocks until the
// engine sends a RenderMsg, then delivers it as a tea.Msg.
// When the engine channel closes it returns QuitMsg.
func waitForEngineMsg(e *engine.Engine) tea.Cmd {
	if e == nil {
		return nil
	}
	return func() tea.Msg {
		msg, ok := <-e.ReceiveMsg()
		if !ok {
			return tea.QuitMsg{}
		}
		return msg
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spin.Tick, waitForEngineMsg(m.eng))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		cmds = append(cmds, m.handleKey(msg))
	case tea.WindowSizeMsg:
		m.handleResize(msg)
	case engine.StreamChunk:
		m.handleStreamChunk(msg)
		cmds = append(cmds, waitForEngineMsg(m.eng))
	case engine.StatusUpdate:
		m.handleStatusUpdate(msg)
		cmds = append(cmds, waitForEngineMsg(m.eng))
	case engine.ErrorMsg:
		m.handleError(msg)
		cmds = append(cmds, waitForEngineMsg(m.eng))
	case engine.ShowGate:
		m.handleShowGate(msg)
		cmds = append(cmds, waitForEngineMsg(m.eng))
	case engine.ShowDiff:
		m.handleShowDiff(msg)
		cmds = append(cmds, waitForEngineMsg(m.eng))
	case engine.StateChangeMsg:
		m.phase = string(msg.NewState)
		m.engineState = msg.NewState
		cmds = append(cmds, waitForEngineMsg(m.eng))
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
