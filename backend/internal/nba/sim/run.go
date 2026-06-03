package sim

// Result is a completed (or safety-stopped) game.
type Result struct {
	State     *State  `json:"state"`
	Events    []Event `json:"events"`
	Truncated bool    `json:"truncated"` // true if safety limit was hit before the clock ended
}

// RunUntilFinal steps until the clock ends or a safety limit is hit.
// The limit scales with config: regulation possessions + room for 20 OT periods.
func (e *Engine) RunUntilFinal() Result {
	g := e.cfg.Game
	otPerPeriod := int(float64(g.OTSeconds)/g.SecondsPerPossession()*2) + 1
	safetyLimit := int(g.ExpectedTotalPossessions()) + 20*otPerPeriod

	state := e.NewGame()
	var events []Event

	for state.Status != "final" && state.Possession < safetyLimit {
		events = append(events, e.Step(state))
	}

	return Result{State: state, Events: events, Truncated: state.Status != "final"}
}
