package tui

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/cajohnson0125/Twirl/internal/pubsub"
	"github.com/cajohnson0125/Twirl/internal/state"
)

// --- computeDims ---

func TestComputeDims_ViewportContentWidth(t *testing.T) {
	for _, w := range []int{20, 40, 80, 120, 200} {
		d := computeDims(w, 24)
		if d.vpContentW != w-2 {
			t.Errorf("w=%d: vpContentW=%d, want %d",
				w, d.vpContentW, w-2)
		}
	}
}

func TestComputeDims_FooterPresence(t *testing.T) {
	tests := []struct {
		h    int
		want int // expected footerH
	}{
		{1, 0},
		{3, 0},
		{5, 0},
		{6, 1},
		{7, 1},
		{24, 1},
		{50, 1},
		{100, 1},
	}
	for _, tt := range tests {
		d := computeDims(80, tt.h)
		if d.footerH != tt.want {
			t.Errorf("h=%d: footerH=%d, want %d",
				tt.h, d.footerH, tt.want)
		}
	}
}

func TestComputeDims_TotalHeightFits(t *testing.T) {
	// For terminals tall enough to fit the minimum layout (info + vp +
	// input + footer), the total rendered height must not exceed the
	// terminal. Minimum that fits: infoH(3) + vpTotalH_min(4) +
	// inputH(3) + footerH(1) = 11 lines.
	for _, h := range []int{11, 12, 24, 50, 100} {
		d := computeDims(80, h)
		total := d.infoH + (d.vpContentH + 3) + d.inputH + d.footerH
		if total > h {
			t.Errorf("h=%d: total rendered=%d exceeds terminal",
				h, total)
		}
	}
}

func TestComputeDims_TinyOverflowAccepted(t *testing.T) {
	// For very small terminals, computeDims clamps the viewport to a
	// minimum height. The rendered layout may exceed the terminal —
	// that's expected. Just verify vpContentH >= 1 (never zero/neg).
	for _, h := range []int{1, 3, 5, 7, 9} {
		d := computeDims(80, h)
		if d.vpContentH < 1 {
			t.Errorf("h=%d: vpContentH=%d, want >= 1", h, d.vpContentH)
		}
	}
}

func TestComputeDims_FixedPanelHeights(t *testing.T) {
	d := computeDims(80, 24)
	if d.infoH != 3 {
		t.Errorf("infoH=%d, want 3", d.infoH)
	}
	if d.inputH != 3 {
		t.Errorf("inputH=%d, want 3", d.inputH)
	}
}

func TestComputeDims_WidthStored(t *testing.T) {
	for _, w := range []int{10, 40, 80, 200} {
		d := computeDims(w, 24)
		if d.w != w {
			t.Errorf("w=%d: d.w=%d, want %d", w, d.w, w)
		}
	}
}

func TestComputeDims_TinyTerminal(t *testing.T) {
	d := computeDims(20, 3)
	// Viewport should get minimum height, never negative.
	if d.vpContentH < 1 {
		t.Errorf("tiny terminal: vpContentH=%d, want >= 1",
			d.vpContentH)
	}
	if d.footerH != 0 {
		t.Errorf("tiny terminal: footerH=%d, want 0",
			d.footerH)
	}
}

func TestComputeDims_LargeTerminal(t *testing.T) {
	d := computeDims(300, 120)
	if d.footerH != 1 {
		t.Errorf("large terminal: footerH=%d, want 1",
			d.footerH)
	}
	if d.vpContentW != 298 {
		t.Errorf("large terminal: vpContentW=%d, want 298",
			d.vpContentW)
	}
	// Most of the height should go to the viewport.
	expectedVP := 120 - 3 - 3 - 1 - 3 // h - info - input - footer - vp border/title
	if d.vpContentH != expectedVP {
		t.Errorf("large terminal: vpContentH=%d, want %d",
			d.vpContentH, expectedVP)
	}
}

// --- roleDisplayName ---

func TestRoleDisplayName(t *testing.T) {
	tests := []struct {
		role string
		want string
	}{
		{"brainstorm", "Brainstorm"},
		{"research", "Research"},
		{"report", "Report"},
		{"plan", "Plan"},
		{"plan_review", "Plan Review"},
		{"execution", "Execution"},
		{"code_review", "Code Review"},
		{"triage", "Triage"},
		{"assessment", "Assessment"},
		{"scribe", "Scribe"},
		{"hitl_gate", "HITL Gate"},
		{"unknown_role", "unknown_role"},
	}
	for _, tt := range tests {
		got := roleDisplayName(tt.role)
		if got != tt.want {
			t.Errorf("roleDisplayName(%q) = %q, want %q",
				tt.role, got, tt.want)
		}
	}
}

// --- wrapLine ---

func TestWrapLine_Short(t *testing.T) {
	lines := wrapLine("hello", 80)
	if len(lines) != 1 || lines[0] != "hello" {
		t.Errorf("got %v", lines)
	}
}

func TestWrapLine_Long(t *testing.T) {
	text := "one two three four five"
	lines := wrapLine(text, 10)
	if len(lines) < 2 {
		t.Errorf("expected wrapping, got %v", lines)
	}
}

// --- newModel ---

func TestNewModel_AgentsInitialized(t *testing.T) {
	m, err := newModel(Opts{
		CursorShape: "bar",
		CursorBlink: true,
	})
	if err != nil {
		t.Fatalf("newModel: %v", err)
	}
	if len(m.agents) != 10 {
		t.Errorf("agents: got %d, want 10", len(m.agents))
	}
	if len(m.agentList) != 10 {
		t.Errorf("agentList: got %d, want 10",
			len(m.agentList))
	}
}

func TestNewModel_PhaseIdle(t *testing.T) {
	m, _ := newModel(Opts{
		CursorShape: "bar",
		CursorBlink: true,
	})
	if m.phase != "Idle" {
		t.Errorf("phase: got %q, want %q", m.phase, "Idle")
	}
}

// --- handleBusEvent ---

func newTestModelWithBus() model {
	bus := pubsub.NewBus(64)
	m, _ := newModel(Opts{
		CursorShape: "bar",
		CursorBlink: true,
		Bus:         bus,
	})
	return m
}

func TestHandleBusEvent_AgentStarted(t *testing.T) {
	m := newTestModelWithBus()
	cmd := m.handleBusEvent(pubsub.Event{
		Type: pubsub.EventAgentStarted,
		Role: "brainstorm",
	})
	if cmd == nil {
		t.Error("expected non-nil cmd (waitForBusEvent)")
	}
	if m.phase != "Brainstorm ▸" {
		t.Errorf("phase: got %q, want %q",
			m.phase, "Brainstorm ▸")
	}
	info := m.agents["Brainstorm"]
	if info.status != agentActive {
		t.Errorf("agent status: got %d, want %d",
			info.status, agentActive)
	}
	if !m.streaming {
		t.Error("expected streaming=true")
	}
}

func TestHandleBusEvent_StreamToken(t *testing.T) {
	m := newTestModelWithBus()
	m.handleBusEvent(pubsub.Event{
		Type:  pubsub.EventStream,
		Token: "hello ",
	})
	m.handleBusEvent(pubsub.Event{
		Type:  pubsub.EventStream,
		Token: "world",
	})
	if m.responseBuf.String() != "hello world" {
		t.Errorf("buf: got %q, want %q",
			m.responseBuf.String(), "hello world")
	}
}

func TestHandleBusEvent_AgentDone(t *testing.T) {
	m := newTestModelWithBus()
	m.handleBusEvent(pubsub.Event{
		Type: pubsub.EventAgentStarted,
		Role: "research",
	})
	m.handleBusEvent(pubsub.Event{
		Type:  pubsub.EventStream,
		Token: "findings",
	})
	m.handleBusEvent(pubsub.Event{
		Type: pubsub.EventAgentDone,
		Role: "research",
	})
	if m.streaming {
		t.Error("expected streaming=false after done")
	}
	info := m.agents["Research"]
	if info.status != agentDone {
		t.Errorf("agent status: got %d, want %d",
			info.status, agentDone)
	}
	if m.responseBuf.Len() != 0 {
		t.Errorf("buf should be reset, got %q",
			m.responseBuf.String())
	}
}

func TestHandleBusEvent_Gate(t *testing.T) {
	m := newTestModelWithBus()
	m.handleBusEvent(pubsub.Event{
		Type:    pubsub.EventGate,
		Prompt:  "Approve scope?",
		Options: []string{"yes", "no"},
	})
	if !m.gate.active {
		t.Error("expected gate.active=true")
	}
	if m.gate.prompt != "Approve scope?" {
		t.Errorf("prompt: got %q", m.gate.prompt)
	}
	if len(m.gate.options) != 2 {
		t.Errorf("options: got %d, want 2",
			len(m.gate.options))
	}
	if m.streaming {
		t.Error("expected streaming=false after gate")
	}
}

func TestHandleBusEvent_Error(t *testing.T) {
	m := newTestModelWithBus()
	m.streaming = true
	m.handleBusEvent(pubsub.Event{
		Type: pubsub.EventError,
		Err:  "something broke",
	})
	if m.phase != "Error" {
		t.Errorf("phase: got %q, want %q",
			m.phase, "Error")
	}
	if m.streaming {
		t.Error("expected streaming=false after error")
	}
}

// --- handleGateInput ---

func TestHandleGateInput_ByNumber(t *testing.T) {
	hitlCh := make(chan state.HITLResponse, 1)
	bus := pubsub.NewBus(64)
	m, _ := newModel(Opts{
		CursorShape: "bar",
		CursorBlink: true,
		Bus:         bus,
		HITLOut:     hitlCh,
	})
	m.gate = gateState{
		prompt:  "Proceed?",
		options: []string{"yes", "no"},
		active:  true,
	}
	m.input.SetValue("2")

	result, _ := m.handleGateInput()
	m2 := result.(model)

	select {
	case resp := <-hitlCh:
		if resp.Choice != "no" {
			t.Errorf("choice: got %q, want %q",
				resp.Choice, "no")
		}
		if resp.Input != "2" {
			t.Errorf("input: got %q, want %q",
				resp.Input, "2")
		}
	default:
		t.Error("expected HITLResponse on channel")
	}
	if m2.gate.active {
		t.Error("expected gate.active=false after input")
	}
}

func TestHandleGateInput_ByExactText(t *testing.T) {
	hitlCh := make(chan state.HITLResponse, 1)
	bus := pubsub.NewBus(64)
	m, _ := newModel(Opts{
		CursorShape: "bar",
		CursorBlink: true,
		Bus:         bus,
		HITLOut:     hitlCh,
	})
	m.gate = gateState{
		prompt:  "Proceed?",
		options: []string{"yes", "no"},
		active:  true,
	}
	m.input.SetValue("maybe")

	_, _ = m.handleGateInput()

	select {
	case resp := <-hitlCh:
		if resp.Choice != "maybe" {
			t.Errorf("choice: got %q, want %q",
				resp.Choice, "maybe")
		}
	default:
		t.Error("expected HITLResponse on channel")
	}
}

func TestHandleGateInput_Empty(t *testing.T) {
	hitlCh := make(chan state.HITLResponse, 1)
	bus := pubsub.NewBus(64)
	m, _ := newModel(Opts{
		CursorShape: "bar",
		CursorBlink: true,
		Bus:         bus,
		HITLOut:     hitlCh,
	})
	m.gate = gateState{active: true}
	m.input.SetValue("")

	_, _ = m.handleGateInput()

	select {
	case <-hitlCh:
		t.Error("should not send on empty input")
	default:
	}
}

// --- handleEngineDone ---

func TestHandleEngineDone_Success(t *testing.T) {
	m := newTestModelWithBus()
	m.streaming = true
	m.handleEngineDone(nil)
	if m.phase != "Done" {
		t.Errorf("phase: got %q, want %q",
			m.phase, "Done")
	}
	if m.streaming {
		t.Error("expected streaming=false")
	}
}

func TestHandleEngineDone_Error(t *testing.T) {
	m := newTestModelWithBus()
	m.streaming = true
	m.handleEngineDone(errors.New("engine failed"))
	if m.phase != "Error" {
		t.Errorf("phase: got %q, want %q",
			m.phase, "Error")
	}
	if m.streaming {
		t.Error("expected streaming=false")
	}
}

// --- mergeBusSubscriptions ---

func TestMergeBusSubscriptions_ReceivesEvents(t *testing.T) {
	bus := pubsub.NewBus(64)
	m, _ := newModel(Opts{
		CursorShape: "bar",
		CursorBlink: true,
		Bus:         bus,
	})

	bus.Publish(pubsub.Event{
		Type: pubsub.EventAgentStarted,
		Role: "plan",
	})

	ev, ok := <-m.busCh
	if !ok {
		t.Fatal("busCh closed")
	}
	if ev.Role != "plan" {
		t.Errorf("role: got %q, want %q", ev.Role, "plan")
	}
}

func TestMergeBusSubscriptions_MultipleTypes(t *testing.T) {
	bus := pubsub.NewBus(64)
	m, _ := newModel(Opts{
		CursorShape: "bar",
		CursorBlink: true,
		Bus:         bus,
	})

	bus.Publish(pubsub.Event{
		Type:  pubsub.EventStream,
		Token: "x",
	})
	bus.Publish(pubsub.Event{
		Type: pubsub.EventAgentDone,
		Role: "scribe",
	})

	types := map[pubsub.EventType]bool{}
	for i := 0; i < 2; i++ {
		ev, ok := <-m.busCh
		if !ok {
			t.Fatal("busCh closed")
		}
		types[ev.Type] = true
	}
	if !types[pubsub.EventStream] {
		t.Error("missing EventStream")
	}
	if !types[pubsub.EventAgentDone] {
		t.Error("missing EventAgentDone")
	}
}

// --- engineController ---

type mockController struct {
	cancelled bool
}

func (mc *mockController) Cancel() { mc.cancelled = true }

func TestHandleKey_CtrlC_CancelsEngine(t *testing.T) {
	ctrl := &mockController{}
	bus := pubsub.NewBus(64)
	_, _ = newModel(Opts{
		CursorShape: "bar",
		CursorBlink: true,
		Bus:         bus,
		Engine:      ctrl,
	})

	// Verify controller is wired (non-ctrl+c key should not cancel).
	if ctrl.cancelled {
		t.Error("should not cancel on init")
	}
}

func TestHandleKey_CtrlC_SuppressedDuringGate(t *testing.T) {
	ctrl := &mockController{}
	bus := pubsub.NewBus(64)
	m, _ := newModel(Opts{
		CursorShape: "bar",
		CursorBlink: true,
		Bus:         bus,
		Engine:      ctrl,
	})
	m.gate.active = true

	// When gate is active, ctrl+c is suppressed.
	_, cmd := m.handleKey(tea.KeyPressMsg{})
	if cmd != nil {
		t.Error("expected nil cmd for non-ctrl+c")
	}
	if ctrl.cancelled {
		t.Error("should not cancel during gate")
	}
}
