package sim

import (
	"math/rand"
	"sort"

	"github.com/chengyang/sim1.0/backend/internal/nba/config"
)

// teamConference builds a team-ID → conference map from cfg.CupGroups.
func teamConferenceMap(groups []config.CupGroup) map[string]string {
	m := make(map[string]string, 30)
	for _, g := range groups {
		for _, id := range g.Teams {
			m[id] = g.Conference
		}
	}
	return m
}

// ── Cup-specific stats (group stage only) ─────────────────────────────────────

type cupTeamStats struct {
	id        string
	groupName string
	w, l      int
	ptsDiff   int
	ptsFor    int
	rsWins    int // RS wins snapshotted after last group-stage game
	h2hW      map[string]int
	h2hL      map[string]int
	tieRng    float64 // pre-assigned random value for tiebreaker 6
}

func newCupStats(id, group string) *cupTeamStats {
	return &cupTeamStats{
		id: id, groupName: group,
		h2hW: make(map[string]int),
		h2hL: make(map[string]int),
	}
}

func (s *cupTeamStats) addResult(win bool, pts, oppPts int, opponent string) {
	if win {
		s.w++
		s.h2hW[opponent]++
	} else {
		s.l++
		s.h2hL[opponent]++
	}
	s.ptsDiff += pts - oppPts
	s.ptsFor += pts
}

// ── Public result types ───────────────────────────────────────────────────────

// CupGameResult is the outcome of one cup knockout game.
type CupGameResult struct {
	Home      string `json:"home"`
	Away      string `json:"away"`
	HomeScore int    `json:"home_score"`
	AwayScore int    `json:"away_score"`
	Winner    string `json:"winner"`
	Counted   bool   `json:"counted"` // true = counts as regular season
}

// ConferenceCupResult holds the bracket and results for one conference.
type ConferenceCupResult struct {
	Seeds [4]string        `json:"seeds"` // index 0 = seed 1, index 3 = wildcard
	QF    [2]CupGameResult `json:"qf"`    // QF[0]=1v4, QF[1]=2v3
	SF    CupGameResult    `json:"sf"`
}

// CupResult is the complete cup knockout result.
type CupResult struct {
	East  ConferenceCupResult `json:"east"`
	West  ConferenceCupResult `json:"west"`
	Final CupGameResult       `json:"final"`
}

// SeasonResult bundles standings and cup data returned by SimulateSeason.
type SeasonResult struct {
	East []TeamSeasonStat
	West []TeamSeasonStat
	Cup  CupResult
}

// ── Bracket determination ─────────────────────────────────────────────────────

// determineCupSeeds returns [seed1, seed2, seed3, wildcard] for one conference.
func determineCupSeeds(conf string, groups []config.CupGroup, stats map[string]*cupTeamStats, rng *rand.Rand) [4]string {
	// Pre-assign random tiebreaker values so sort is stable and reproducible.
	for _, s := range stats {
		s.tieRng = rng.Float64()
	}

	var winners []string
	var nonWinners []string

	for _, g := range groups {
		if g.Conference != conf {
			continue
		}
		ranked := rankCupTeams(g.Teams, stats)
		winners = append(winners, ranked[0])
		nonWinners = append(nonWinners, ranked[1:]...)
	}

	rankedWinners := rankCupTeams(winners, stats)
	rankedNonWinners := rankCupTeams(nonWinners, stats)

	return [4]string{
		rankedWinners[0],
		rankedWinners[1],
		rankedWinners[2],
		rankedNonWinners[0], // wild card is always seed 4
	}
}

// rankCupTeams sorts ids by the 6-step cup tiebreaker.
func rankCupTeams(ids []string, stats map[string]*cupTeamStats) []string {
	out := make([]string, len(ids))
	copy(out, ids)
	sort.SliceStable(out, func(i, j int) bool {
		a, b := stats[out[i]], stats[out[j]]
		// 1. Cup wins
		if a.w != b.w {
			return a.w > b.w
		}
		// 2. Head-to-head record in cup (net wins between these two)
		netA := a.h2hW[out[j]] - a.h2hL[out[j]]
		netB := b.h2hW[out[i]] - b.h2hL[out[i]]
		if netA != netB {
			return netA > netB
		}
		// 3. Point differential across all cup games
		if a.ptsDiff != b.ptsDiff {
			return a.ptsDiff > b.ptsDiff
		}
		// 4. Total points scored in cup games
		if a.ptsFor != b.ptsFor {
			return a.ptsFor > b.ptsFor
		}
		// 5. Regular-season wins at time of bracket
		if a.rsWins != b.rsWins {
			return a.rsWins > b.rsWins
		}
		// 6. Random drawing
		return a.tieRng < b.tieRng
	})
	return out
}

// ── Cup simulation ────────────────────────────────────────────────────────────

func runCup(
	cfg *config.Bundle,
	stats map[string]*cupTeamStats,
	records map[string]*teamRecord,
	rng *rand.Rand,
) CupResult {
	var res CupResult

	// Seed both conferences.
	res.East.Seeds = determineCupSeeds("east", cfg.CupGroups, stats, rng)
	res.West.Seeds = determineCupSeeds("west", cfg.CupGroups, stats, rng)

	// QF: seeds 1 and 2 are home. Counts as regular season.
	res.East.QF[0] = playCupGame(cfg, res.East.Seeds[0], res.East.Seeds[3], true, records, rng)
	res.East.QF[1] = playCupGame(cfg, res.East.Seeds[1], res.East.Seeds[2], true, records, rng)
	res.West.QF[0] = playCupGame(cfg, res.West.Seeds[0], res.West.Seeds[3], true, records, rng)
	res.West.QF[1] = playCupGame(cfg, res.West.Seeds[1], res.West.Seeds[2], true, records, rng)

	// SF: higher seed (lower index) is home. Counts as regular season.
	res.East.SF = playSFGame(cfg, res.East.Seeds[:], res.East.QF, records, rng)
	res.West.SF = playSFGame(cfg, res.West.Seeds[:], res.West.QF, records, rng)

	// Final: coin flip for home. Does NOT count as regular season.
	eastFin := res.East.SF.Winner
	westFin := res.West.SF.Winner
	finHome, finAway := eastFin, westFin
	if rng.Intn(2) == 1 {
		finHome, finAway = westFin, eastFin
	}
	res.Final = playCupGame(cfg, finHome, finAway, false, records, rng)

	return res
}

func playCupGame(
	cfg *config.Bundle,
	home, away string,
	counted bool,
	records map[string]*teamRecord,
	rng *rand.Rand,
) CupGameResult {
	eng, _ := NewEngine(cfg, home, away, rng.Int63())
	r := eng.RunUntilFinal()
	homeWon := r.State.HomeScore > r.State.AwayScore
	winner := away
	if homeWon {
		winner = home
	}
	if counted {
		records[home].record(homeWon, true, r.State.HomeScore, r.State.AwayScore, records[away])
		records[away].record(!homeWon, false, r.State.AwayScore, r.State.HomeScore, records[home])
	}
	return CupGameResult{
		Home: home, Away: away,
		HomeScore: r.State.HomeScore, AwayScore: r.State.AwayScore,
		Winner: winner, Counted: counted,
	}
}

func playSFGame(
	cfg *config.Bundle,
	seeds []string,
	qf [2]CupGameResult,
	records map[string]*teamRecord,
	rng *rand.Rand,
) CupGameResult {
	w0 := qf[0].Winner // winner of 1v4
	w1 := qf[1].Winner // winner of 2v3
	// Higher original seed (lower slice index) hosts.
	i0 := cupSeedIndex(seeds, w0)
	i1 := cupSeedIndex(seeds, w1)
	home, away := w0, w1
	if i1 < i0 {
		home, away = w1, w0
	}
	return playCupGame(cfg, home, away, true, records, rng)
}

func cupSeedIndex(seeds []string, team string) int {
	for i, s := range seeds {
		if s == team {
			return i
		}
	}
	return len(seeds)
}

// ── Flex game generation and simulation ───────────────────────────────────────

// generateAndSimFlexGames ensures every non-SF team finishes with exactly
// 82 games by generating flex matchups to fill each team's remaining home
// and away deficits (targets: 41 H + 41 A).
//
// SF+ teams already have 82 games (QF + SF both count); finalists are also
// at 82 because the Final does not count. Those teams get no flex games.
//
// When two QF home-seeds in the same conference both lose (both needing an
// away flex game with no matching home partner), the global needHome list
// may have fewer entries than needAway. In that case we balance by flipping
// excess away entries into home entries — those teams end up at 42 H / 40 A
// instead of 41/41, but still reach exactly 82 total games.
func generateAndSimFlexGames(
	cfg *config.Bundle,
	records map[string]*teamRecord,
	rng *rand.Rand,
) {
	var needHome []string
	var needAway []string

	for _, g := range cfg.CupGroups {
		for _, id := range g.Teams {
			r := records[id]
			h := r.homeW + r.homeL
			a := r.awayW + r.awayL
			// SF+ teams already have 82 games — skip entirely.
			if h+a >= 82 {
				continue
			}
			for range make([]struct{}, max0(41-h)) {
				needHome = append(needHome, id)
			}
			for range make([]struct{}, max0(41-a)) {
				needAway = append(needAway, id)
			}
		}
	}

	// Balance the two lists. This imbalance arises when QF home-seeds (1 or 2)
	// lose their game: they sit at 41H/40A and land in needAway, but there is
	// no matching extra home entry to pair them with. We fix by flipping one
	// of those QF home-losers into needHome instead — they end up at 42H/40A
	// (still 82 total). We always prefer to flip QF participants (41H+40A or
	// 40H+41A) so that non-cup teams are never affected.
	for len(needHome) < len(needAway) {
		// Prefer a QF home-loser sitting at 41H/40A.
		idx := -1
		for i, id := range needAway {
			r := records[id]
			if r.homeW+r.homeL == 41 && r.awayW+r.awayL == 40 {
				idx = i
				break
			}
		}
		if idx < 0 {
			idx = len(needAway) - 1 // fallback: any team
		}
		needHome = append(needHome, needAway[idx])
		needAway = append(needAway[:idx], needAway[idx+1:]...)
	}
	for len(needAway) < len(needHome) {
		// Prefer a QF away-loser sitting at 40H/41A.
		idx := -1
		for i, id := range needHome {
			r := records[id]
			if r.homeW+r.homeL == 40 && r.awayW+r.awayL == 41 {
				idx = i
				break
			}
		}
		if idx < 0 {
			idx = len(needHome) - 1
		}
		needAway = append(needAway, needHome[idx])
		needHome = append(needHome[:idx], needHome[idx+1:]...)
	}

	rng.Shuffle(len(needHome), func(i, j int) { needHome[i], needHome[j] = needHome[j], needHome[i] })
	rng.Shuffle(len(needAway), func(i, j int) { needAway[i], needAway[j] = needAway[j], needAway[i] })

	for len(needHome) > 0 && len(needAway) > 0 {
		h := needHome[0]
		needHome = needHome[1:]

		// Rotate needAway until the away team differs from the home team.
		for len(needAway) > 1 && needAway[0] == h {
			needAway = append(needAway[1:], needAway[0])
		}
		a := needAway[0]
		needAway = needAway[1:]

		if h == a {
			continue // degenerate edge case: skip self-game
		}

		eng, _ := NewEngine(cfg, h, a, rng.Int63())
		r := eng.RunUntilFinal()
		homeWon := r.State.HomeScore > r.State.AwayScore
		records[h].record(homeWon, true, r.State.HomeScore, r.State.AwayScore, records[a])
		records[a].record(!homeWon, false, r.State.AwayScore, r.State.HomeScore, records[h])
	}
}

func max0(n int) int {
	if n > 0 {
		return n
	}
	return 0
}
