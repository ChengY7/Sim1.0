package api

// SimulateRequest is the body for POST /simulate.
type SimulateRequest struct {
	HomeTeamID string `json:"home_team_id" example:"LAL"`
	AwayTeamID string `json:"away_team_id" example:"BOS"`
	Seed       *int64 `json:"seed,omitempty" example:"42"`
}

// SimulateResponse is returned after a full game simulation.
type SimulateResponse struct {
	Seed   int64       `json:"seed" example:"42"`
	State  GameState   `json:"state"`
	Events []GameEvent `json:"events"`
}

// GameState is the final state of a simulated game.
type GameState struct {
	HomeID     string `json:"home_id" example:"LAL"`
	AwayID     string `json:"away_id" example:"BOS"`
	HomeName   string `json:"home_name" example:"Lakers"`
	AwayName   string `json:"away_name" example:"Celtics"`
	HomeScore  int    `json:"home_score" example:"105"`
	AwayScore  int    `json:"away_score" example:"98"`
	Offense    string `json:"offense" example:"home"`
	Possession int    `json:"possession" example:"201"`
	Quarter    int    `json:"quarter" example:"4"`
	ClockSec   int    `json:"clock_sec" example:"0"`
	Status     string `json:"status" example:"final"`
}

// GameEvent describes a single possession.
type GameEvent struct {
	Possession int    `json:"possession" example:"1"`
	Team       string `json:"team" example:"home"`
	Type       string `json:"type" example:"make_2pt"`
	Points     int    `json:"points" example:"2"`
	Text       string `json:"text" example:"Lakers make_2pt (2 pts)"`
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
