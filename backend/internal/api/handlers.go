package api

import (
	"encoding/json"
	"net/http"
	"sort"
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
	opts := make([]TeamOption, 0, len(h.cfg.Teams))
	for _, t := range h.cfg.Teams {
		opts = append(opts, TeamOption{ID: t.ID, Name: t.Name})
	}
	sort.Slice(opts, func(i, j int) bool { return opts[i].Name < opts[j].Name })
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
		Seed:   seed,
		State:  result.State,
		Events: result.Events,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{Error: msg})
}
