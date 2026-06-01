package agent

// Task is what the orchestration layer sends to an agent when
// dispatching it. It carries the instruction, relevant project
// context, and the path to the output template the agent should
// produce.
type Task struct {
	// Instruction is the prompt or directive for the agent.
	Instruction string

	// Context is the relevant project context the agent needs
	// to complete its task (accumulated state, prior results,
	// user input, etc.).
	Context map[string]string

	// TemplatePath is the path to the markdown template the
	// agent should produce (relative to templates/ directory).
	// Empty if the agent doesn't produce a template-based
	// document.
	TemplatePath string
}
