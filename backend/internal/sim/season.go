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
	W          int     `json:"w"`
	L          int     `json:"l"`
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
	id       string
	name     string
	results  []bool // chronological W/L; true = win
	homeW    int
	homeL    int
	awayW    int
	awayL    int
	totalPts int
	totalOpp int
}

func (tr *teamRecord) record(win, home bool, pts, opp int) {
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
}

func (tr *teamRecord) toStat() TeamSeasonStat {
	n := len(tr.results)
	stat := TeamSeasonStat{
		TeamID:     tr.id,
		TeamName:   tr.name,
		HomeRecord: fmt.Sprintf("%d-%d", tr.homeW, tr.homeL),
		AwayRecord: fmt.Sprintf("%d-%d", tr.awayW, tr.awayL),
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

// winCount returns the number of wins in tr.results.
func winCount(tr *teamRecord) int {
	n := 0
	for _, w := range tr.results {
		if w {
			n++
		}
	}
	return n
}

// SimulateSeason simulates the full 2025-26 regular season including the NBA Cup.
// It returns standings sorted by wins (desc) and the complete cup bracket.
func SimulateSeason(cfg *config.Bundle, seed int64) SeasonResult {
	// Initialise per-team tracking.
	records := make(map[string]*teamRecord, len(cfg.Teams))
	for id, t := range cfg.Teams {
		records[id] = &teamRecord{id: id, name: t.Name}
	}

	cupStats := make(map[string]*cupTeamStats, 30)
	for _, g := range cfg.CupGroups {
		for _, id := range g.Teams {
			cupStats[id] = newCupStats(id, g.Name)
		}
	}

	masterRng := rand.New(rand.NewSource(seed))

	// "2025-11-28" is the date of the last NBA Cup group-stage game.
	// We snapshot RS wins just before the first game played after that date.
	const groupStageEnd = "2025-11-28"
	groupSnapped := false

	for _, game := range cfg.Schedule.Games {
		if game.Excluded {
			continue
		}
		if _, ok := cfg.Teams[game.Home]; !ok {
			continue
		}
		if _, ok := cfg.Teams[game.Away]; !ok {
			continue
		}

		// Snapshot RS wins before the first post-group-stage game.
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

		records[game.Home].record(homeWon, true, res.State.HomeScore, res.State.AwayScore)
		records[game.Away].record(!homeWon, false, res.State.AwayScore, res.State.HomeScore)

		// Track cup group-stage stats for bracket determination.
		if game.NbaCup {
			if cs, ok := cupStats[game.Home]; ok {
				cs.addResult(homeWon, res.State.HomeScore, res.State.AwayScore, game.Away)
			}
			if cs, ok := cupStats[game.Away]; ok {
				cs.addResult(!homeWon, res.State.AwayScore, res.State.HomeScore, game.Home)
			}
		}
	}

	// Fallback: if schedule ended before the cutoff date.
	if !groupSnapped {
		for id, rec := range records {
			if cs, ok := cupStats[id]; ok {
				cs.rsWins = winCount(rec)
			}
		}
	}

	// Simulate NBA Cup knockout bracket.
	cup := runCup(cfg, cupStats, records, masterRng)

	// Generate and simulate flex games to reach 41H / 41A per team.
	generateAndSimFlexGames(cfg, records, masterRng)

	// Build final standings.
	stats := make([]TeamSeasonStat, 0, len(records))
	for _, tr := range records {
		stats = append(stats, tr.toStat())
	}
	sort.Slice(stats, func(i, j int) bool {
		if stats[i].W != stats[j].W {
			return stats[i].W > stats[j].W
		}
		return stats[i].L < stats[j].L
	})

	return SeasonResult{Standings: stats, Cup: cup}
}
