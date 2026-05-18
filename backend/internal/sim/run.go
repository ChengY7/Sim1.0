package sim

const MaxPossSafety = 300

// Result is a completed (or safety-stopped) game.
type Result struct {
	State  *State  `json:"state"`
	Events []Event `json:"events"`
}

// RunUntilFinal steps until the clock ends or MaxPossSafety is hit.
func (e *Engine) RunUntilFinal() Result {
	state := e.NewGame()
	var events []Event

	for state.Status != "final" && state.Possession < MaxPossSafety {
		events = append(events, e.Step(state))
	}

	return Result{State: state, Events: events}
}
