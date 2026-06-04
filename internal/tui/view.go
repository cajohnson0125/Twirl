package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"

	"github.com/cajohnson0125/Twirl/internal/engine"
)

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

	stateLabel := stateDisplay(m.engineState)
	if isGateState(m.engineState) {
		stateLabel = "⏳ " + stateLabel
	} else if m.engineState == engine.StateFiling {
		stateLabel = m.spin.View() + " filing"
	}

	activeStyle := styleActive.Bold(true).Render(
		"▸ " + active)
	phase := styleLabel.Render("State:") + " " +
		styleValue.Render(stateLabel)
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

func (m model) viewViewport() string {
	content := stylePanelTitle.Render("OUTPUT") +
		"\n" + m.output.View()
	return stylePanelBorder.
		Width(m.d.w).
		Height(m.d.vpContentH + 3).
		Render(content)
}

func (m model) viewInputBar() string {
	if isGateState(m.engineState) {
		return styleInputBar.
			Width(m.d.w).
			Render(styleMuted.Render("  Waiting for gate response..."))
	}
	inputContent := m.input.View()
	return styleInputBar.
		Width(m.d.w).
		Render(inputContent)
}

func (m model) viewFooter() string {
	var parts []string
	if isGateState(m.engineState) {
		parts = append(parts, "y approve", "n reject")
	} else {
		parts = append(parts, "ctrl+c quit")
		if m.width > 30 {
			parts = append(parts, "enter send")
		}
	}
	if m.width > 50 && !isGateState(m.engineState) {
		parts = append(parts, "↑↓ scroll")
	}
	return styleFooter.Width(m.width).Render(
		strings.Join(parts, "  •  "),
	)
}
