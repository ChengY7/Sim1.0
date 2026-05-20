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

	// Streak — count consecutive identical results from the end.
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

	// Scoring averages (1 decimal place)
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

// SimulateSeason simulates every non-excluded game in the schedule and returns
// standings sorted by wins (desc), then losses (asc).
// Each game gets a deterministic seed derived from the master seed, so the
// same master seed always produces the same season.
func SimulateSeason(cfg *config.Bundle, seed int64) []TeamSeasonStat {
	records := make(map[string]*teamRecord, len(cfg.Teams))
	for id, t := range cfg.Teams {
		records[id] = &teamRecord{id: id, name: t.Name}
	}

	masterRng := rand.New(rand.NewSource(seed))

	for _, game := range cfg.Schedule.Games {
		if game.Excluded {
			continue
		}

		// Skip games whose teams are not in the bundle (safety guard).
		if _, ok := cfg.Teams[game.Home]; !ok {
			continue
		}
		if _, ok := cfg.Teams[game.Away]; !ok {
			continue
		}

		gameSeed := masterRng.Int63()
		engine, err := NewEngine(cfg, game.Home, game.Away, gameSeed)
		if err != nil {
			continue
		}
		result := engine.RunUntilFinal()

		homeScore := result.State.HomeScore
		awayScore := result.State.AwayScore
		homeWon := homeScore > awayScore

		records[game.Home].record(homeWon, true, homeScore, awayScore)
		records[game.Away].record(!homeWon, false, awayScore, homeScore)
	}

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

	return stats
}
