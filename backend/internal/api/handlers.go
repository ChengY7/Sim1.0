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

type simulateRequest struct {
	HomeTeamID string `json:"home_team_id"`
	AwayTeamID string `json:"away_team_id"`
	Seed       *int64 `json:"seed,omitempty"`
}

type teamOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (h *Handlers) ListTeams(w http.ResponseWriter, r *http.Request) {
	opts := make([]teamOption, 0, len(h.cfg.Teams))
	for _, t := range h.cfg.Teams {
		opts = append(opts, teamOption{ID: t.ID, Name: t.Name})
	}
	sort.Slice(opts, func(i, j int) bool { return opts[i].Name < opts[j].Name })
	writeJSON(w, http.StatusOK, map[string]any{"teams": opts})
}

func (h *Handlers) Simulate(w http.ResponseWriter, r *http.Request) {
	var req simulateRequest
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
	writeJSON(w, http.StatusOK, map[string]any{
		"seed":   seed,
		"state":  result.State,
		"events": result.Events,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
