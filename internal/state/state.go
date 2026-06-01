package state

// EngineStatus represents the current status of a workflow run.
type EngineStatus int

const (
	StatusRunning EngineStatus = iota
	StatusPausedHITL
	StatusCompleted
	StatusFailed
)

// Event records a single event in the workflow audit trail.
type Event struct {
	Type    string
	NodeID  string
	Message string
}

// Result holds the output of a single node execution.
type Result struct {
	Status      EngineStatus
	OutputPath  string
	HITLRequest *HITLRequest
	Context     map[string]string
	Severity    int
	Action      string
	Error       string
}

// State is the internal memory of the engine. Persisted as binary
// via gob — not human-readable by design.
type State struct {
	ProjectID   string
	ActiveNodes []string
	Status      EngineStatus
	Context     map[string]string
	AuditLog    []Event
	PendingHITL *HITLRequest
}
