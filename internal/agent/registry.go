package agent

import "fmt"

// Registry maps agent roles to constructors. The engine looks up
// agents by role at dispatch time.
type Registry struct {
	agents map[Role]func() Agent
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		agents: make(map[Role]func() Agent),
	}
}

// Register adds a constructor for the given role. Panics on
// duplicate registration.
func (r *Registry) Register(role Role, fn func() Agent) {
	if _, ok := r.agents[role]; ok {
		panic("agent: duplicate registration for role: " + string(role))
	}
	r.agents[role] = fn
}

// Get creates and returns a new agent for the given role.
// Returns an error if the role is not registered.
func (r *Registry) Get(role Role) (Agent, error) {
	fn, ok := r.agents[role]
	if !ok {
		return nil, fmt.Errorf("agent: unregistered role: %s", role)
	}
	return fn(), nil
}

// Roles returns all registered roles.
func (r *Registry) Roles() []Role {
	roles := make([]Role, 0, len(r.agents))
	for r := range r.agents {
		roles = append(roles, r)
	}
	return roles
}
