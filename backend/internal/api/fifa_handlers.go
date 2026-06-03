package api

import (
	"encoding/json"
	"net/http"
	"time"

	fifasim "github.com/chengyang/sim1.0/backend/internal/fifa/sim"
)

// FIFAListTeams godoc
// @Summary      List all FIFA World Cup 2026 teams
// @Tags         fifa
// @Produce      json
// @Success      200  {object}  FIFAListTeamsResponse
// @Router       /fifa/teams [get]
func (h *Handlers) FIFAListTeams(w http.ResponseWriter, r *http.Request) {
	teams := h.fifaCfg.SortedTeams()
	opts := make([]FIFATeamOption, len(teams))
	for i, t := range teams {
		opts[i] = FIFATeamOption{ID: t.ID, Name: t.Name, Group: t.Group, Elo: t.Elo, Host: t.Host}
	}
	writeJSON(w, http.StatusOK, FIFAListTeamsResponse{Teams: opts})
}

// FIFASimulate godoc
// @Summary      Simulate a FIFA World Cup match
// @Description  Uses a Poisson goals model with Dixon-Coles correction. ExtraTime=true adds AET and penalties if the match is drawn.
// @Tags         fifa
// @Accept       json
// @Produce      json
// @Param        body  body      FIFASimulateRequest  true  "Home/away team ids, optional seed, extra_time flag"
// @Success      200   {object}  FIFASimulateResponse
// @Failure      400   {object}  ErrorResponse
// @Router       /fifa/simulate [post]
func (h *Handlers) FIFASimulate(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var req FIFASimulateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Team1ID == "" || req.Team2ID == "" {
		writeError(w, http.StatusBadRequest, "team1_id and team2_id required")
		return
	}
	if req.Team1ID == req.Team2ID {
		writeError(w, http.StatusBadRequest, "teams must differ")
		return
	}

	team1, err := h.fifaCfg.Team(req.Team1ID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	team2, err := h.fifaCfg.Team(req.Team2ID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	seed := time.Now().UnixNano()
	if req.Seed != nil {
		seed = *req.Seed
	}

	result := fifasim.SimulateMatch(team1, team2, req.ExtraTime, req.HostTeamID, seed)

	resp := FIFASimulateResponse{
		Seed:          seed,
		Team1ID:       result.Team1ID,
		Team2ID:       result.Team2ID,
		Team1Name:     team1.Name,
		Team2Name:     team2.Name,
		Team1Goals:    result.Team1Goals,
		Team2Goals:    result.Team2Goals,
		Team1GoalsAET: result.Team1GoalsAET,
		Team2GoalsAET: result.Team2GoalsAET,
		Winner:        result.Winner,
		Period:        result.Period,
	}
	if result.Penalties != nil {
		kicks := make([]FIFAPenaltyKick, len(result.Penalties.Kicks))
		for i, k := range result.Penalties.Kicks {
			kicks[i] = FIFAPenaltyKick{Team: k.Team, KickNo: k.KickNo, Scored: k.Scored}
		}
		resp.Penalties = &FIFAPenaltyResult{
			Team1Scored: result.Penalties.Team1Scored,
			Team2Scored: result.Penalties.Team2Scored,
			Kicks:       kicks,
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

// FIFASimulateGroupStage godoc
// @Summary      Simulate the FIFA World Cup 2026 group stage
// @Description  Runs all 72 group-stage matches, returns standings for all 12 groups ordered by points. Top 2 per group advance; best 8 third-place teams also advance.
// @Tags         fifa
// @Accept       json
// @Produce      json
// @Param        body  body      FIFASimulateGroupStageRequest  false  "Optional seed"
// @Success      200   {object}  FIFASimulateGroupStageResponse
// @Router       /fifa/simulate/group-stage [post]
func (h *Handlers) FIFASimulateGroupStage(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var req FIFASimulateGroupStageRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	seed := time.Now().UnixNano()
	if req.Seed != nil {
		seed = *req.Seed
	}

	result := fifasim.SimulateGroupStage(h.fifaCfg, seed)

	groups := make([]FIFAGroupStanding, len(result.Groups))
	for i, g := range result.Groups {
		teams := make([]FIFATeamStanding, len(g.Teams))
		for j, t := range g.Teams {
			teams[j] = FIFATeamStanding{
				TeamID:  t.TeamID,
				Name:    t.Name,
				MP:      t.MP,
				W:       t.W,
				D:       t.D,
				L:       t.L,
				GF:      t.GF,
				GA:      t.GA,
				GD:      t.GD,
				Pts:     t.Pts,
				Advance: t.Advance,
			}
		}
		groups[i] = FIFAGroupStanding{Group: g.Group, Teams: teams}
	}

	writeJSON(w, http.StatusOK, FIFASimulateGroupStageResponse{Seed: seed, Groups: groups})
}
