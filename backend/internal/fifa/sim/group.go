package sim

import (
	"math/rand"
	"sort"

	"github.com/chengyang/sim1.0/backend/internal/fifa/config"
)

var groupOrder = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L"}

type TeamStanding struct {
	TeamID  string
	Name    string
	Group   string
	MP      int
	W       int
	D       int
	L       int
	GF      int
	GA      int
	GD      int
	Pts     int
	Advance bool
}

type GroupResult struct {
	Group string
	Teams []TeamStanding
}

type GroupStageResult struct {
	Groups []GroupResult
}

// h2hRecord tracks goals scored by one team against another.
type h2hRecord struct{ GF, GA int }

func SimulateGroupStage(bundle *config.Bundle, seed int64) GroupStageResult {
	rng := rand.New(rand.NewSource(seed))

	// Initialise standings.
	standings := make(map[string]*TeamStanding, len(bundle.Teams))
	for _, t := range bundle.Teams {
		standings[t.ID] = &TeamStanding{TeamID: t.ID, Name: t.Name, Group: t.Group}
	}

	// h2h[a][b] = goals scored by a in matches against b.
	h2h := make(map[string]map[string]h2hRecord, len(bundle.Teams))
	for id := range bundle.Teams {
		h2h[id] = make(map[string]h2hRecord)
	}

	// Simulate every group-stage match.
	// Each match gets seed+i+1 so the full result is reproducible from a single seed.
	for i, m := range bundle.Matches {
		t1 := bundle.Teams[m.Team1]
		t2 := bundle.Teams[m.Team2]

		hostID := ""
		if t1.Host {
			hostID = t1.ID
		} else if t2.Host {
			hostID = t2.ID
		}

		res := SimulateMatch(t1, t2, false, hostID, seed+int64(i+1))
		g1, g2 := res.Team1Goals, res.Team2Goals

		s1, s2 := standings[t1.ID], standings[t2.ID]
		s1.MP++
		s2.MP++
		s1.GF += g1
		s1.GA += g2
		s2.GF += g2
		s2.GA += g1
		s1.GD = s1.GF - s1.GA
		s2.GD = s2.GF - s2.GA

		switch {
		case g1 > g2:
			s1.W++
			s1.Pts += 3
			s2.L++
		case g2 > g1:
			s2.W++
			s2.Pts += 3
			s1.L++
		default:
			s1.D++
			s1.Pts++
			s2.D++
			s2.Pts++
		}

		r1 := h2h[t1.ID][t2.ID]
		r1.GF += g1
		r1.GA += g2
		h2h[t1.ID][t2.ID] = r1

		r2 := h2h[t2.ID][t1.ID]
		r2.GF += g2
		r2.GA += g1
		h2h[t2.ID][t1.ID] = r2
	}

	// Pre-assign random tiebreaker floats so sorting is stable and reproducible.
	tieRng := make(map[string]float64, len(standings))
	for id := range standings {
		tieRng[id] = rng.Float64()
	}

	// Sort each group and collect 3rd-place finishers.
	byGroup := make(map[string][]*TeamStanding)
	for _, s := range standings {
		byGroup[s.Group] = append(byGroup[s.Group], s)
	}

	results := make([]GroupResult, 0, 12)
	thirdPlace := make([]*TeamStanding, 0, 12)

	for _, g := range groupOrder {
		teams := byGroup[g]
		sortGroup(teams, h2h, tieRng)

		// Top 2 automatically advance.
		teams[0].Advance = true
		teams[1].Advance = true
		thirdPlace = append(thirdPlace, teams[2])

		out := make([]TeamStanding, len(teams))
		for i, t := range teams {
			out[i] = *t
		}
		results = append(results, GroupResult{Group: g, Teams: out})
	}

	// Rank the 12 third-place teams: Pts → GD → GF → random draw.
	sort.SliceStable(thirdPlace, func(i, j int) bool {
		a, b := thirdPlace[i], thirdPlace[j]
		if a.Pts != b.Pts {
			return a.Pts > b.Pts
		}
		if a.GD != b.GD {
			return a.GD > b.GD
		}
		if a.GF != b.GF {
			return a.GF > b.GF
		}
		return tieRng[a.TeamID] < tieRng[b.TeamID]
	})

	// Top 8 third-place teams advance; mark them in the results.
	for i := 0; i < 8; i++ {
		id := thirdPlace[i].TeamID
		for ri := range results {
			for ti := range results[ri].Teams {
				if results[ri].Teams[ti].TeamID == id {
					results[ri].Teams[ti].Advance = true
				}
			}
		}
	}

	return GroupStageResult{Groups: results}
}

// sortGroup ranks four teams using:
// 1. Points  2. GD  3. GF  4. H2H record  5. coin flip
func sortGroup(teams []*TeamStanding, h2h map[string]map[string]h2hRecord, tieRng map[string]float64) {
	sort.SliceStable(teams, func(i, j int) bool {
		a, b := teams[i], teams[j]
		if a.Pts != b.Pts {
			return a.Pts > b.Pts
		}
		if a.GD != b.GD {
			return a.GD > b.GD
		}
		if a.GF != b.GF {
			return a.GF > b.GF
		}
		// Head-to-head (pairwise)
		aVsB := h2h[a.TeamID][b.TeamID]
		bVsA := h2h[b.TeamID][a.TeamID]
		aPts := h2hPoints(aVsB)
		bPts := h2hPoints(bVsA)
		if aPts != bPts {
			return aPts > bPts
		}
		aGD := aVsB.GF - aVsB.GA
		bGD := bVsA.GF - bVsA.GA
		if aGD != bGD {
			return aGD > bGD
		}
		if aVsB.GF != bVsA.GF {
			return aVsB.GF > bVsA.GF
		}
		return tieRng[a.TeamID] < tieRng[b.TeamID]
	})
}

func h2hPoints(r h2hRecord) int {
	switch {
	case r.GF > r.GA:
		return 3
	case r.GF == r.GA:
		return 1
	default:
		return 0
	}
}
