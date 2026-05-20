package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/chengyang/sim1.0/backend/internal/config"
	"github.com/chengyang/sim1.0/backend/internal/sim"
)

type Handlers struct {
	cfg *config.Bundle
}

func NewHandlers(cfg *config.Bundle) *Handlers {
	return &Handlers{cfg: cfg}
}

// ListTeams godoc
// @Summary      List all teams
// @Description  Returns team ids and names from teams.json
// @Tags         teams
// @Produce      json
// @Success      200  {object}  ListTeamsResponse
// @Router       /teams [get]
func (h *Handlers) ListTeams(w http.ResponseWriter, r *http.Request) {
	teams := h.cfg.SortedTeams()
	opts := make([]TeamOption, len(teams))
	for i, t := range teams {
		opts[i] = TeamOption{ID: t.ID, Name: t.Name}
	}
	writeJSON(w, http.StatusOK, ListTeamsResponse{Teams: opts})
}

// Simulate godoc
// @Summary      Simulate a full game
// @Description  Runs possessions until the game clock ends. Same seed produces the same game.
// @Tags         simulate
// @Accept       json
// @Produce      json
// @Param        body  body      SimulateRequest  true  "Home/away team ids and optional seed"
// @Success      200   {object}  SimulateResponse
// @Failure      400   {object}  ErrorResponse
// @Router       /simulate [post]
func (h *Handlers) Simulate(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var req SimulateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.HomeTeamID == "" || req.AwayTeamID == "" {
		writeError(w, http.StatusBadRequest, "home_team_id and away_team_id required")
		return
	}
	if req.HomeTeamID == req.AwayTeamID {
		writeError(w, http.StatusBadRequest, "teams must differ")
		return
	}

	seed := time.Now().UnixNano()
	if req.Seed != nil {
		seed = *req.Seed
	}

	engine, err := sim.NewEngine(h.cfg, req.HomeTeamID, req.AwayTeamID, seed)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result := engine.RunUntilFinal()
	writeJSON(w, http.StatusOK, SimulateResponse{
		Seed:      seed,
		Truncated: result.Truncated,
		State:     toGameState(result.State, h.cfg.Game.Quarters),
		Events:    toGameEvents(result.Events),
	})
}

func toGameState(s *sim.State, quarters int) GameState {
	return GameState{
		HomeID:     s.HomeID,
		AwayID:     s.AwayID,
		HomeName:   s.HomeName,
		AwayName:   s.AwayName,
		HomeScore:  s.HomeScore,
		AwayScore:  s.AwayScore,
		Offense:    offenseField(s),
		Possession: s.Possession,
		Quarter:    s.Quarter,
		Period:     sim.PeriodLabel(s.Quarter, quarters),
		ClockSec:   s.ClockSec,
		Status:     s.Status,
	}
}

func offenseField(s *sim.State) string {
	if s.Status == "final" {
		return ""
	}
	return string(s.Offense)
}

func toGameEvents(events []sim.Event) []GameEvent {
	out := make([]GameEvent, len(events))
	for i, e := range events {
		out[i] = GameEvent{
			Possession: e.Possession,
			Period:     e.Period,
			ClockSec:   e.ClockSec,
			Team:       string(e.Team),
			Type:       e.Type,
			Points:     e.Points,
			Text:       e.Text,
			HomeScore:  e.HomeScore,
			AwayScore:  e.AwayScore,
		}
	}
	return out
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{Error: msg})
}
