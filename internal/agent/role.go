package agent

// Role identifies a specialist agent.
type Role string

const (
	Brainstorm  Role = "brainstorm"
	Research    Role = "research"
	Report      Role = "report"
	Plan        Role = "plan"
	PlanReview  Role = "plan_review"
	Execution   Role = "execution"
	CodeReview  Role = "code_review"
	Triage      Role = "triage"
	Assessment  Role = "assessment"
	Scribe      Role = "scribe"
)
