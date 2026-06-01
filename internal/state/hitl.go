package state

// HITLRequest is sent to the presentation layer when the engine
// needs human input before proceeding.
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
