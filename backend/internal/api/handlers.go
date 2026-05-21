package api

import (
	"encoding/json"
	"fmt"
	"log"
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

// ListSeasons godoc
// @Summary      List available season rating files
// @Description  Returns season identifiers that can be passed as the "season" parameter to simulate endpoints.
// @Tags         seasons
// @Produce      json
// @Success      200  {object}  ListSeasonsResponse
// @Router       /seasons [get]
func (h *Handlers) ListSeasons(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, ListSeasonsResponse{
		Seasons:       h.cfg.AvailableSeasons,
		DefaultSeason: h.cfg.DefaultSeason,
	})
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
		opts[i] = TeamOption{ID: t.ID, Name: t.Name, Conference: t.Conference, Division: t.Division}
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

	cfg, err := h.cfgForSeason(req.Season)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	engine, err := sim.NewEngine(cfg, req.HomeTeamID, req.AwayTeamID, seed)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result := engine.RunUntilFinal()
	writeJSON(w, http.StatusOK, SimulateResponse{
		Seed:      seed,
		Truncated: result.Truncated,
		State:     toGameState(result.State, cfg.Game.Quarters),
		Events:    toGameEvents(result.Events),
	})
}

// SimulateSeason godoc
// @Summary      Simulate a full NBA regular season
// @Description  Runs every game in the 2025-26 schedule and returns standings with W, L, streak, last-10, home/away records, PPG, OPPG, and DIFF.
// @Tags         simulate
// @Accept       json
// @Produce      json
// @Param        body  body      SimulateSeasonRequest  false  "Optional seed for reproducibility"
// @Success      200   {object}  SimulateSeasonResponse
// @Router       /simulate/season [post]
func (h *Handlers) SimulateSeason(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var req SimulateSeasonRequest
	// Body is optional — ignore decode errors for empty bodies.
	_ = json.NewDecoder(r.Body).Decode(&req)

	seed := time.Now().UnixNano()
	if req.Seed != nil {
		seed = *req.Seed
	}

	cfg, err := h.cfgForSeason(req.Season)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result := sim.SimulateSeason(cfg, seed)

	writeJSON(w, http.StatusOK, SimulateSeasonResponse{
		Seed:   seed,
		Season: cfg.Schedule.Season,
		East:   mapSeasonStats(result.East),
		West:   mapSeasonStats(result.West),
		Cup:    mapCupBracket(result.Cup),
	})
}

// SimulatePlayIn godoc
// @Summary      Simulate the NBA play-in tournament
// @Description  Runs all 6 play-in games (3 per conference). Game 1: 7 hosts 8 — winner = 7 seed. Game 2: 9 hosts 10. Game 3: loser of G1 hosts winner of G2 — winner = 8 seed.
// @Tags         simulate
// @Accept       json
// @Produce      json
// @Param        body  body      SimulatePlayInRequest  true  "Play-in team IDs for each conference"
// @Success      200   {object}  SimulatePlayInResponse
// @Failure      400   {object}  ErrorResponse
// @Router       /simulate/playin [post]
func (h *Handlers) SimulatePlayIn(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var req SimulatePlayInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	east := req.East
	west := req.West
	for label, ids := range map[string][4]string{
		"east": {east.Seed7, east.Seed8, east.Seed9, east.Seed10},
		"west": {west.Seed7, west.Seed8, west.Seed9, west.Seed10},
	} {
		for i, id := range ids {
			if id == "" {
				writeError(w, http.StatusBadRequest, fmt.Sprintf("%s seed%d is empty", label, i+7))
				return
			}
		}
	}

	seed := time.Now().UnixNano()
	if req.Seed != nil {
		seed = *req.Seed
	}

	cfg, err := h.cfgForSeason(req.Season)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result := sim.SimulatePlayIn(cfg,
		sim.PlayInInput{Seed7: east.Seed7, Seed8: east.Seed8, Seed9: east.Seed9, Seed10: east.Seed10},
		sim.PlayInInput{Seed7: west.Seed7, Seed8: west.Seed8, Seed9: west.Seed9, Seed10: west.Seed10},
		seed,
	)

	writeJSON(w, http.StatusOK, SimulatePlayInResponse{
		Seed: seed,
		East: mapConferencePlayIn(result.East),
		West: mapConferencePlayIn(result.West),
	})
}

func mapPlayInGame(g sim.PlayInGameResult) PlayInGame {
	return PlayInGame{
		Home:      g.Home,
		Away:      g.Away,
		HomeScore: g.HomeScore,
		AwayScore: g.AwayScore,
		Winner:    g.Winner,
		Loser:     g.Loser,
	}
}

func mapConferencePlayIn(c sim.ConferencePlayInResult) ConferencePlayIn {
	return ConferencePlayIn{
		Game1:    mapPlayInGame(c.Game1),
		Game2:    mapPlayInGame(c.Game2),
		Game3:    mapPlayInGame(c.Game3),
		Playoff7: c.Playoff7,
		Playoff8: c.Playoff8,
	}
}

// SimulateDraftLottery godoc
// @Summary      Simulate the NBA draft lottery
// @Description  Runs the NBA draft lottery for 14 teams using official ball-combination odds. Picks 1-4 are drawn by weighted lottery; picks 5-14 go to remaining teams in seed order.
// @Tags         simulate
// @Accept       json
// @Produce      json
// @Param        body  body      SimulateDraftLotteryRequest  true  "14 team IDs in lottery-seed order (index 0 = worst record)"
// @Success      200   {object}  SimulateDraftLotteryResponse
// @Failure      400   {object}  ErrorResponse
// @Router       /simulate/draft-lottery [post]
func (h *Handlers) SimulateDraftLottery(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var req SimulateDraftLotteryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	for i, id := range req.Teams {
		if id == "" {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("teams[%d] is empty", i))
			return
		}
	}

	seed := time.Now().UnixNano()
	if req.Seed != nil {
		seed = *req.Seed
	}

	result := sim.SimulateDraftLottery(req.Teams, h.cfg.DraftLottery.Combinations, seed)

	picks := make([]DraftPick, len(result.Picks))
	for i, p := range result.Picks {
		picks[i] = DraftPick{Pick: p.Pick, TeamID: p.TeamID, Seed: p.Seed}
	}

	writeJSON(w, http.StatusOK, SimulateDraftLotteryResponse{Seed: seed, Picks: picks})
}

// cfgForSeason returns the bundle with ratings for the requested season,
// falling back to the default season if season is nil or empty.
func (h *Handlers) cfgForSeason(season *string) (*config.Bundle, error) {
	s := h.cfg.DefaultSeason
	if season != nil && *season != "" {
		s = *season
	}
	if s == "" {
		return h.cfg, nil
	}
	return h.cfg.WithSeason(s)
}

func mapSeasonStats(in []sim.TeamSeasonStat) []TeamSeasonStat {
	out := make([]TeamSeasonStat, len(in))
	for i, s := range in {
		out[i] = TeamSeasonStat{
			TeamID:     s.TeamID,
			TeamName:   s.TeamName,
			Conference: s.Conference,
			Division:   s.Division,
			W:          s.W,
			L:          s.L,
			ConfRecord: s.ConfRecord,
			DivRecord:  s.DivRecord,
			Streak:     s.Streak,
			Last10:     s.Last10,
			HomeRecord: s.HomeRecord,
			AwayRecord: s.AwayRecord,
			PPG:        s.PPG,
			OPPG:       s.OPPG,
			Diff:       s.Diff,
		}
	}
	return out
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

func mapCupGame(g sim.CupGameResult) CupGame {
	return CupGame{
		Home:      g.Home,
		Away:      g.Away,
		HomeScore: g.HomeScore,
		AwayScore: g.AwayScore,
		Winner:    g.Winner,
		Counted:   g.Counted,
	}
}

func mapConferenceCup(c sim.ConferenceCupResult) ConferenceCup {
	return ConferenceCup{
		Seeds: c.Seeds,
		QF:    [2]CupGame{mapCupGame(c.QF[0]), mapCupGame(c.QF[1])},
		SF:    mapCupGame(c.SF),
	}
}

func mapCupBracket(c sim.CupResult) CupBracket {
	return CupBracket{
		East:  mapConferenceCup(c.East),
		West:  mapConferenceCup(c.West),
		Final: mapCupGame(c.Final),
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON encode: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{Error: msg})
}
