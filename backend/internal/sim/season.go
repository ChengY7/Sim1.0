package sim

import (
	"fmt"
	"math"
	"math/rand"
	"sort"

	"github.com/chengyang/sim1.0/backend/internal/config"
)

// TeamSeasonStat is the end-of-season summary for one team.
type TeamSeasonStat struct {
	TeamID     string  `json:"team_id"`
	TeamName   string  `json:"team_name"`
	Conference string  `json:"conference"`
	Division   string  `json:"division"`
	W          int     `json:"w"`
	L          int     `json:"l"`
	ConfRecord string  `json:"conf_record"`
	DivRecord  string  `json:"div_record"`
	Streak     string  `json:"streak"`
	Last10     string  `json:"last_10"`
	HomeRecord string  `json:"home_record"`
	AwayRecord string  `json:"away_record"`
	PPG        float64 `json:"ppg"`
	OPPG       float64 `json:"oppg"`
	Diff       float64 `json:"diff"`
}

// teamRecord accumulates raw stats during the simulation.
type teamRecord struct {
	id         string
	name       string
	conference string
	division   string
	results    []bool // chronological W/L; true = win
	homeW      int
	homeL      int
	awayW      int
	awayL      int
	confW      int
	confL      int
	divW       int
	divL       int
	totalPts   int
	totalOpp   int
	netPtDiff  float64         // capped at ±10 per game, accumulated
	h2h        map[string]int  // wins against each opponent ID
}

func (tr *teamRecord) record(win, home bool, pts, opp int, against *teamRecord) {
	tr.results = append(tr.results, win)
	if home {
		if win {
			tr.homeW++
		} else {
			tr.homeL++
		}
	} else {
		if win {
			tr.awayW++
		} else {
			tr.awayL++
		}
	}
	tr.totalPts += pts
	tr.totalOpp += opp

	if win {
		tr.h2h[against.id]++
	}

	diff := pts - opp
	if diff > 10 {
		diff = 10
	} else if diff < -10 {
		diff = -10
	}
	tr.netPtDiff += float64(diff)

	if tr.conference == against.conference {
		if win {
			tr.confW++
		} else {
			tr.confL++
		}
		if tr.division == against.division {
			if win {
				tr.divW++
			} else {
				tr.divL++
			}
		}
	}
}

func (tr *teamRecord) toStat() TeamSeasonStat {
	n := len(tr.results)
	stat := TeamSeasonStat{
		TeamID:     tr.id,
		TeamName:   tr.name,
		Conference: tr.conference,
		Division:   tr.division,
		HomeRecord: fmt.Sprintf("%d-%d", tr.homeW, tr.homeL),
		AwayRecord: fmt.Sprintf("%d-%d", tr.awayW, tr.awayL),
		ConfRecord: fmt.Sprintf("%d-%d", tr.confW, tr.confL),
		DivRecord:  fmt.Sprintf("%d-%d", tr.divW, tr.divL),
	}

	for _, w := range tr.results {
		if w {
			stat.W++
		} else {
			stat.L++
		}
	}

	// Streak
	if n > 0 {
		last := tr.results[n-1]
		count := 0
		for i := n - 1; i >= 0 && tr.results[i] == last; i-- {
			count++
		}
		if last {
			stat.Streak = fmt.Sprintf("W%d", count)
		} else {
			stat.Streak = fmt.Sprintf("L%d", count)
		}
	}

	// Last 10
	start := n - 10
	if start < 0 {
		start = 0
	}
	w10 := 0
	for _, w := range tr.results[start:] {
		if w {
			w10++
		}
	}
	played := n - start
	stat.Last10 = fmt.Sprintf("%d-%d", w10, played-w10)

	// Scoring averages
	if n > 0 {
		stat.PPG = round1(float64(tr.totalPts) / float64(n))
		stat.OPPG = round1(float64(tr.totalOpp) / float64(n))
		stat.Diff = round1(stat.PPG - stat.OPPG)
	}

	return stat
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}

func winCount(tr *teamRecord) int {
	n := 0
	for _, w := range tr.results {
		if w {
			n++
		}
	}
	return n
}

// pct returns win percentage, returning 0 if no games played.
func pct(w, l int) float64 {
	total := w + l
	if total == 0 {
		return 0
	}
	return float64(w) / float64(total)
}

// sortWithTiebreaker sorts stats in-place using the NBA tiebreaker chain:
// 1. Wins (desc)
// 2. H2H record (pairwise, among tied teams)
// 3. Division record (if same division)
// 4. Conference record
// 5. Capped net point differential
func sortWithTiebreaker(stats []TeamSeasonStat, records map[string]*teamRecord) {
	sort.SliceStable(stats, func(i, j int) bool {
		a, b := stats[i], stats[j]
		if a.W != b.W {
			return a.W > b.W
		}
		ra, rb := records[a.TeamID], records[b.TeamID]

		// 1. H2H
		aH2H := ra.h2h[b.TeamID]
		bH2H := rb.h2h[a.TeamID]
		if aH2H != bH2H {
			return aH2H > bH2H
		}

		// 2. Division record (same division only)
		if ra.division == rb.division {
			pa, pb := pct(ra.divW, ra.divL), pct(rb.divW, rb.divL)
			if pa != pb {
				return pa > pb
			}
		}

		// 3. Conference record
		pa, pb := pct(ra.confW, ra.confL), pct(rb.confW, rb.confL)
		if pa != pb {
			return pa > pb
		}

		// 4. Capped point differential
		return ra.netPtDiff > rb.netPtDiff
	})
}

// SimulateSeason simulates the full 2025-26 regular season including the NBA Cup.
// Returns standings split by conference, sorted with NBA tiebreakers applied.
func SimulateSeason(cfg *config.Bundle, seed int64) SeasonResult {
	records := make(map[string]*teamRecord, len(cfg.Teams))
	for id, t := range cfg.Teams {
		records[id] = &teamRecord{
			id:         id,
			name:       t.Name,
			conference: t.Conference,
			division:   t.Division,
			h2h:        make(map[string]int),
		}
	}

	cupStats := make(map[string]*cupTeamStats, 30)
	for _, g := range cfg.CupGroups {
		for _, id := range g.Teams {
			cupStats[id] = newCupStats(id, g.Name)
		}
	}

	masterRng := rand.New(rand.NewSource(seed))

	const groupStageEnd = "2025-11-28"
	groupSnapped := false

	for _, game := range cfg.Schedule.Games {
		if game.Excluded {
			continue
		}
		homeRec, homeOK := records[game.Home]
		awayRec, awayOK := records[game.Away]
		if !homeOK || !awayOK {
			continue
		}

		if !groupSnapped && game.Date > groupStageEnd {
			for id, rec := range records {
				if cs, ok := cupStats[id]; ok {
					cs.rsWins = winCount(rec)
				}
			}
			groupSnapped = true
		}

		eng, err := NewEngine(cfg, game.Home, game.Away, masterRng.Int63())
		if err != nil {
			continue
		}
		res := eng.RunUntilFinal()
		homeWon := res.State.HomeScore > res.State.AwayScore

		homeRec.record(homeWon, true, res.State.HomeScore, res.State.AwayScore, awayRec)
		awayRec.record(!homeWon, false, res.State.AwayScore, res.State.HomeScore, homeRec)

		if game.NbaCup {
			if cs, ok := cupStats[game.Home]; ok {
				cs.addResult(homeWon, res.State.HomeScore, res.State.AwayScore, game.Away)
			}
			if cs, ok := cupStats[game.Away]; ok {
				cs.addResult(!homeWon, res.State.AwayScore, res.State.HomeScore, game.Home)
			}
		}
	}

	if !groupSnapped {
		for id, rec := range records {
			if cs, ok := cupStats[id]; ok {
				cs.rsWins = winCount(rec)
			}
		}
	}

	cup := runCup(cfg, cupStats, records, masterRng)
	generateAndSimFlexGames(cfg, records, masterRng)

	var east, west []TeamSeasonStat
	for _, tr := range records {
		stat := tr.toStat()
		if tr.conference == "east" {
			east = append(east, stat)
		} else {
			west = append(west, stat)
		}
	}

	sortWithTiebreaker(east, records)
	sortWithTiebreaker(west, records)

	return SeasonResult{East: east, West: west, Cup: cup}
}
