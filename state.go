package sigmon

// State holds information set by a corresponding signal.
type State struct {
	s Signal
}

func newState(s Signal) *State {
	return &State{
		s: s,
	}
}

// Signal returns a representation of the system signal which spawned the
// creation of the State instance.
func (s *State) Signal() Signal {
	return s.s
}
