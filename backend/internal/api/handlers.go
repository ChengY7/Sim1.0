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
// @Router       /nba/seasons [get]
func (h *Handlers) ListSeasons(w http.ResponseWriter, r *http.Request) {
	seasons := make([]SeasonInfo, len(h.cfg.AvailableSeasons))
	for i, s := range h.cfg.AvailableSeasons {
		info := SeasonInfo{Season: s}
		if records, ok := h.cfg.SeasonStandings[s]; ok {
			info.Standings = make([]TeamRecord, len(records))
			for j, rec := range records {
				info.Standings[j] = TeamRecord{ID: rec.ID, W: rec.W, L: rec.L}
			}
		}
		seasons[i] = info
	}
	writeJSON(w, http.StatusOK, ListSeasonsResponse{
		Seasons:       seasons,
		DefaultSeason: h.cfg.DefaultSeason,
	})
}

// ListTeams godoc
// @Summary      List all teams
// @Description  Returns team ids and names from teams.json
// @Tags         teams
// @Produce      json
// @Success      200  {object}  ListTeamsResponse
// @Router       /nba/teams [get]
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
// @Router       /nba/simulate [post]
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
// @Router       /nba/simulate/season [post]
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
// @Router       /nba/simulate/playin [post]
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
// @Router       /nba/simulate/draft-lottery [post]
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

// SimulatePlayoffs godoc
// @Summary      Simulate the NBA playoffs
// @Description  Simulates all playoff series from the current bracket state to the champion. Each conference seeds 1-8; matchups follow standard NBA seeding (1v8, 4v5, 3v6, 2v7). Home court follows HHAAAHA schedule. Finals home court goes to the team with the better regular-season record (coin flip on a tie). Partially completed series are resumed from their current wins.
// @Tags         simulate
// @Accept       json
// @Produce      json
// @Param        body  body      SimulatePlayoffsRequest  true  "16 playoff teams (8 east, 8 west) with seeds and records; optional partial bracket state"
// @Success      200   {object}  SimulatePlayoffsResponse
// @Failure      400   {object}  ErrorResponse
// @Router       /nba/simulate/playoffs [post]
func (h *Handlers) SimulatePlayoffs(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 16384)
	var req SimulatePlayoffsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if len(req.East) != 8 {
		writeError(w, http.StatusBadRequest, "east must contain exactly 8 teams")
		return
	}
	if len(req.West) != 8 {
		writeError(w, http.StatusBadRequest, "west must contain exactly 8 teams")
		return
	}
	for label, teams := range map[string][]PlayoffTeamInput{"east": req.East, "west": req.West} {
		seen := make(map[int]bool)
		for _, t := range teams {
			if t.TeamID == "" {
				writeError(w, http.StatusBadRequest, fmt.Sprintf("%s: team_id must not be empty", label))
				return
			}
			if t.Seed < 1 || t.Seed > 8 {
				writeError(w, http.StatusBadRequest, fmt.Sprintf("%s team %s: seed must be 1-8", label, t.TeamID))
				return
			}
			if seen[t.Seed] {
				writeError(w, http.StatusBadRequest, fmt.Sprintf("%s: duplicate seed %d", label, t.Seed))
				return
			}
			seen[t.Seed] = true
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

	input := sim.PlayoffInput{
		East: toSimPlayoffTeams(req.East),
		West: toSimPlayoffTeams(req.West),
	}
	if req.Bracket != nil {
		input.Bracket = toSimBracketState(req.Bracket)
	}

	result := sim.SimulatePlayoffs(cfg, input, seed)

	writeJSON(w, http.StatusOK, SimulatePlayoffsResponse{
		Seed:     seed,
		East:     mapConferencePlayoffResult(result.East),
		West:     mapConferencePlayoffResult(result.West),
		Finals:   mapPlayoffSeries(result.Finals),
		Champion: result.Champion,
	})
}

func toSimPlayoffTeams(in []PlayoffTeamInput) [8]sim.PlayoffTeam {
	var out [8]sim.PlayoffTeam
	for i, t := range in {
		out[i] = sim.PlayoffTeam{TeamID: t.TeamID, Seed: t.Seed, SeasonW: t.SeasonW, SeasonL: t.SeasonL}
	}
	return out
}

func toSimSeriesState(in PlayoffSeriesInput) sim.PlayoffSeriesState {
	return sim.PlayoffSeriesState{
		HomeTeam: in.HomeTeam,
		AwayTeam: in.AwayTeam,
		HomeWins: in.HomeWins,
		AwayWins: in.AwayWins,
	}
}

func toSimBracketState(in *PlayoffBracketInput) sim.PlayoffBracketState {
	var out sim.PlayoffBracketState
	for i := range in.EastR1 {
		out.EastR1[i] = toSimSeriesState(in.EastR1[i])
	}
	for i := range in.EastR2 {
		out.EastR2[i] = toSimSeriesState(in.EastR2[i])
	}
	out.EastR3 = toSimSeriesState(in.EastR3)
	for i := range in.WestR1 {
		out.WestR1[i] = toSimSeriesState(in.WestR1[i])
	}
	for i := range in.WestR2 {
		out.WestR2[i] = toSimSeriesState(in.WestR2[i])
	}
	out.WestR3 = toSimSeriesState(in.WestR3)
	out.Finals = toSimSeriesState(in.Finals)
	return out
}

func mapPlayoffGame(g sim.PlayoffGameResult) PlayoffGame {
	return PlayoffGame{
		GameNum:   g.GameNum,
		Home:      g.Home,
		Away:      g.Away,
		HomeScore: g.HomeScore,
		AwayScore: g.AwayScore,
		Winner:    g.Winner,
	}
}

func mapPlayoffSeries(s sim.PlayoffSeriesResult) PlayoffSeries {
	games := make([]PlayoffGame, len(s.Games))
	for i, g := range s.Games {
		games[i] = mapPlayoffGame(g)
	}
	return PlayoffSeries{
		HomeTeam: s.HomeTeam,
		AwayTeam: s.AwayTeam,
		HomeWins: s.HomeWins,
		AwayWins: s.AwayWins,
		Winner:   s.Winner,
		Games:    games,
	}
}

func mapConferencePlayoffResult(c sim.ConferencePlayoffResult) ConferencePlayoffBracket {
	var r1 [4]PlayoffSeries
	for i, s := range c.R1 {
		r1[i] = mapPlayoffSeries(s)
	}
	var r2 [2]PlayoffSeries
	for i, s := range c.R2 {
		r2[i] = mapPlayoffSeries(s)
	}
	return ConferencePlayoffBracket{
		R1:       r1,
		R2:       r2,
		R3:       mapPlayoffSeries(c.R3),
		Champion: c.Champion,
	}
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
