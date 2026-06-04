package tui

import (
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"github.com/cajohnson0125/Twirl/internal/engine"
)

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

func countActive(agents []agent) int {
	n := 0
	for _, a := range agents {
		if a.status == statusActive {
			n++
		}
	}
	return n
}

// isGateState returns true if the engine is waiting for a gate
// approval and user text input should be suppressed.
func isGateState(s engine.State) bool {
	return s == engine.StateCoordinatorGate ||
		s == engine.StateSpecialistGate
}

// stateDisplay returns a human-friendly label for an engine state.
func stateDisplay(s engine.State) string {
	switch s {
	case engine.StateCoordinator:
		return "Coordinator"
	case engine.StateCoordinatorGate:
		return "Awaiting Approval"
	case engine.StateSpecialistRoom:
		return "Specialist"
	case engine.StateSpecialistGate:
		return "Awaiting Approval"
	case engine.StateFiling:
		return "Filing"
	default:
		return string(s)
	}
}

func (m model) handleKey(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "ctrl+c":
		return func() tea.Msg { return tea.QuitMsg{} }
	case "enter":
		if val := m.input.Value(); val != "" {
			return m.startStreaming(val)
		}
	case "y":
		if isGateState(m.engineState) {
			m.eng.SendEvent(engine.GateResponse{
				Approved: true,
				GateID:   string(m.engineState),
			})
			return nil
		}
	case "n":
		if isGateState(m.engineState) {
			m.eng.SendEvent(engine.GateResponse{
				Approved: false,
				GateID:   string(m.engineState),
			})
			return nil
		}
	case "up":
		m.output.ScrollUp(1)
	case "down":
		m.output.ScrollDown(1)
	}
	return nil
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

func (m *model) handleStreamChunk(msg engine.StreamChunk) {
	if msg.Content != "" {
		m.logs = append(m.logs,
			styleAI.Render("AI: ")+msg.Content,
		)
	}
	if msg.Done {
		m.logs = append(m.logs,
			styleDivider.Render(
				strings.Repeat("─", max(20, m.d.vpContentW)),
			),
		)
	}
	m.syncOutput()
	m.output.GotoBottom()
}

func (m *model) handleStatusUpdate(msg engine.StatusUpdate) {
	m.phase = msg.Phase
	for i := range m.agents {
		if m.agents[i].name == msg.Agent {
			m.agents[i].status = statusActive
		} else {
			m.agents[i].status = statusIdle
		}
	}
}

func (m *model) handleError(msg engine.ErrorMsg) {
	m.logs = append(m.logs,
		styleError.Render("Error: "+msg.Message),
	)
	m.syncOutput()
	m.output.GotoBottom()
}

func (m *model) handleShowGate(msg engine.ShowGate) {
	m.logs = append(m.logs,
		styleAccent.Render("Gate: "+msg.Message),
	)
	m.syncOutput()
	m.output.GotoBottom()
}

func (m *model) handleShowDiff(msg engine.ShowDiff) {
	m.logs = append(m.logs,
		styleAI.Render("Diff: "+msg.Title)+"\n"+msg.Content,
	)
	m.syncOutput()
	m.output.GotoBottom()
}

func (m model) startStreaming(
	prompt string,
) tea.Cmd {
	m.logs = append(m.logs,
		styleUser.Render("You: ")+prompt,
	)
	m.syncOutput()
	m.output.GotoBottom()
	m.input.Reset()
	m.eng.SendEvent(engine.UserInput{Text: prompt})
	return nil
}
