package api

import "github.com/chengyang/sim1.0/backend/internal/sim"

// SimulateRequest is the body for POST /simulate.
type SimulateRequest struct {
	HomeTeamID string `json:"home_team_id" example:"LAL"`
	AwayTeamID string `json:"away_team_id" example:"BOS"`
	Seed       *int64 `json:"seed,omitempty" example:"42"`
}

// SimulateResponse is returned after a full game simulation.
type SimulateResponse struct {
	Seed   int64       `json:"seed" example:"42"`
	State  *sim.State  `json:"state"`
	Events []sim.Event `json:"events"`
}

// TeamOption is a team entry for dropdowns.
type TeamOption struct {
	ID   string `json:"id" example:"LAL"`
	Name string `json:"name" example:"Lakers"`
}

// ListTeamsResponse lists all teams from config.
type ListTeamsResponse struct {
	Teams []TeamOption `json:"teams"`
}

// ErrorResponse is returned on 4xx errors.
type ErrorResponse struct {
	Error string `json:"error" example:"unknown team id"`
}
