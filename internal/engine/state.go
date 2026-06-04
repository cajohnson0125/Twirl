package engine

import "fmt"

// State represents the current operational state of the engine.
// Legal transitions are enforced by ValidTransitions.
type State string

const (
	StateCoordinator     State = "coordinator"
	StateCoordinatorGate State = "coordinator_gate"
	StateSpecialistRoom  State = "specialist_room"
	StateSpecialistGate  State = "specialist_gate"
	StateFiling          State = "filing"
)

// ValidTransitions defines which state transitions are allowed.
// Any transition not in this map is illegal and will be rejected.
var ValidTransitions = map[State][]State{
	StateCoordinator: {
		StateCoordinatorGate,
		StateFiling,
	},
	StateCoordinatorGate: {
		StateCoordinator,
		StateSpecialistRoom,
	},
	StateSpecialistRoom: {
		StateSpecialistGate,
		StateCoordinator,
	},
	StateSpecialistGate: {
		StateCoordinator,
		StateSpecialistRoom,
	},
	StateFiling: {
		StateCoordinator,
	},
}

// CanTransitionTo returns true if the engine can legally move
// from the current state to the target state.
func CanTransitionTo(current, target State) bool {
	allowed, ok := ValidTransitions[current]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == target {
			return true
		}
	}
	return false
}

// ValidateTransition returns an error if the transition is illegal.
func ValidateTransition(current, target State) error {
	if CanTransitionTo(current, target) {
		return nil
	}
	return fmt.Errorf(
		"invalid state transition: %s → %s", current, target)
}
