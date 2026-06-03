package orchestrator

// HITLRequest is sent when a specialist needs human input.
type HITLRequest struct {
	ID      string
	Prompt  string
	Options []string
}

// HITLResponse is the user's answer to a HITLRequest.
type HITLResponse struct {
	ID     string
	Choice string
	Input  string
}
