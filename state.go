package sigmon

// State ...
type State struct {
	s Signal
}

func newState(s Signal) *State {
	return &State{
		s: s,
	}
}

// Signal ...
func (s *State) Signal() Signal {
	return s.s
}
