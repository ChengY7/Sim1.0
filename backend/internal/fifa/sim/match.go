package sim

import (
	"math"
	"math/rand"

	"github.com/chengyang/sim1.0/backend/internal/fifa/config"
)

const (
	baseGoals90       = 1.25  // WC average goals per team per 90 min
	baseGoalsAET      = 0.30  // per team in extra time (30 min, fatigue-adjusted)
	eloDivisor        = 600.0 // controls how steeply Elo gap affects expected goals
	dixonColesRho     = -0.13 // Dixon-Coles low-score correction (fitted to international data)
	penConversionBase = 0.75  // baseline international penalty conversion rate
	penEloScale       = 0.03  // max ±3% conversion adjustment per 1000 Elo difference
	maxPenRounds      = 25    // safety cap on sudden death
	maxGoals          = 10    // upper bound for probability table
	HostAdvantageElo  = 100   // Elo points added to a host team's effective rating (eloratings.net standard)
)

// PenaltyKick is a single penalty attempt.
type PenaltyKick struct {
	Team   string `json:"team"`
	KickNo int    `json:"kick_no"`
	Scored bool   `json:"scored"`
}

// PenaltyResult holds the full penalty shootout.
type PenaltyResult struct {
	Team1Scored int           `json:"team1_scored"`
	Team2Scored int           `json:"team2_scored"`
	Kicks       []PenaltyKick `json:"kicks"`
}

// MatchResult is the outcome of one simulated match.
// All World Cup matches are at neutral venues — team1/team2 are positional only.
type MatchResult struct {
	Team1ID       string         `json:"team1_id"`
	Team2ID       string         `json:"team2_id"`
	Team1Goals    int            `json:"team1_goals"`
	Team2Goals    int            `json:"team2_goals"`
	Team1GoalsAET *int           `json:"team1_goals_aet,omitempty"` // additional goals in ET
	Team2GoalsAET *int           `json:"team2_goals_aet,omitempty"`
	Penalties     *PenaltyResult `json:"penalties,omitempty"`
	Winner        string         `json:"winner,omitempty"` // empty = draw (group stage)
	Period        string         `json:"period"`           // "90" | "aet" | "pens"
}

// SimulateMatch runs one match between team1 and team2.
// hostTeamID optionally identifies a host nation playing on home soil — they receive
// a +100 Elo boost (the standard eloratings.net crowd-advantage adjustment).
// If extraTime is false, draws stand (group stage). If true, draws go to AET then penalties.
func SimulateMatch(team1, team2 config.Team, extraTime bool, hostTeamID string, seed int64) MatchResult {
	rng := rand.New(rand.NewSource(seed))

	elo1, elo2 := float64(team1.Elo), float64(team2.Elo)
	switch hostTeamID {
	case team1.ID:
		elo1 += HostAdvantageElo
	case team2.ID:
		elo2 += HostAdvantageElo
	}

	diff := elo1 - elo2
	lambda1 := baseGoals90 * math.Exp(diff/eloDivisor)
	lambda2 := baseGoals90 * math.Exp(-diff/eloDivisor)

	g1, g2 := sampleScoreDixonColes(lambda1, lambda2, rng)

	res := MatchResult{
		Team1ID:    team1.ID,
		Team2ID:    team2.ID,
		Team1Goals: g1,
		Team2Goals: g2,
	}

	switch {
	case g1 > g2:
		res.Period = "90"
		res.Winner = team1.ID
	case g2 > g1:
		res.Period = "90"
		res.Winner = team2.ID
	case !extraTime:
		res.Period = "90" // draw stands (group stage)
	default:
		// Extra time — reduced goal rate due to fatigue.
		aet1 := poissonSample(baseGoalsAET*math.Exp(diff/eloDivisor), rng)
		aet2 := poissonSample(baseGoalsAET*math.Exp(-diff/eloDivisor), rng)
		res.Team1GoalsAET = &aet1
		res.Team2GoalsAET = &aet2

		total1 := g1 + aet1
		total2 := g2 + aet2

		if total1 != total2 {
			res.Period = "aet"
			if total1 > total2 {
				res.Winner = team1.ID
			} else {
				res.Winner = team2.ID
			}
		} else {
			pens := simulatePenalties(team1, team2, rng)
			res.Period = "pens"
			res.Penalties = &pens
			if pens.Team1Scored > pens.Team2Scored {
				res.Winner = team1.ID
			} else {
				res.Winner = team2.ID
			}
		}
	}

	return res
}

// sampleScoreDixonColes draws a (home, away) scoreline using the Dixon-Coles
// correction, which inflates 0-0 and 1-1 draws and deflates 1-0 / 0-1 outcomes
// to better match observed international match distributions.
func sampleScoreDixonColes(lambdaH, lambdaA float64, rng *rand.Rand) (int, int) {
	total := 0.0
	probs := make([]float64, (maxGoals+1)*(maxGoals+1))
	for h := 0; h <= maxGoals; h++ {
		for a := 0; a <= maxGoals; a++ {
			p := poissonPMF(h, lambdaH) * poissonPMF(a, lambdaA) * dixonColesTau(h, a, lambdaH, lambdaA)
			probs[h*(maxGoals+1)+a] = p
			total += p
		}
	}

	r := rng.Float64() * total
	cum := 0.0
	for h := 0; h <= maxGoals; h++ {
		for a := 0; a <= maxGoals; a++ {
			cum += probs[h*(maxGoals+1)+a]
			if r <= cum {
				return h, a
			}
		}
	}
	return maxGoals, maxGoals
}

// dixonColesTau is the correction factor τ(i,j) from Dixon & Coles (1997).
func dixonColesTau(h, a int, lambdaH, lambdaA float64) float64 {
	rho := dixonColesRho
	switch {
	case h == 0 && a == 0:
		return 1 - lambdaH*lambdaA*rho
	case h == 1 && a == 0:
		return 1 + lambdaA*rho
	case h == 0 && a == 1:
		return 1 + lambdaH*rho
	case h == 1 && a == 1:
		return 1 - rho
	default:
		return 1
	}
}

// poissonPMF returns P(X = k) for X ~ Poisson(lambda).
func poissonPMF(k int, lambda float64) float64 {
	if lambda <= 0 {
		if k == 0 {
			return 1
		}
		return 0
	}
	return math.Exp(-lambda) * math.Pow(lambda, float64(k)) / factorial(k)
}

// poissonSample draws a single sample from Poisson(lambda) via Knuth's algorithm.
func poissonSample(lambda float64, rng *rand.Rand) int {
	L := math.Exp(-lambda)
	p, k := 1.0, 0
	for {
		p *= rng.Float64()
		if p <= L {
			return k
		}
		k++
	}
}

func factorial(n int) float64 {
	f := 1.0
	for i := 2; i <= n; i++ {
		f *= float64(i)
	}
	return f
}

// simulatePenalties runs a full penalty shootout: 5 rounds then sudden death.
// Both teams kick in each round of sudden death; the round ends when they diverge.
func simulatePenalties(team1, team2 config.Team, rng *rand.Rand) PenaltyResult {
	// Slightly adjust conversion rates based on Elo gap (max ±3%).
	eloDelta := float64(team1.Elo-team2.Elo) / 1000.0 * penEloScale
	conv1 := math.Max(0.60, math.Min(0.92, penConversionBase+eloDelta))
	conv2 := math.Max(0.60, math.Min(0.92, penConversionBase-eloDelta))

	var kicks []PenaltyKick
	scored1, scored2 := 0, 0

	// Rounds 1–5
	for i := 1; i <= 5; i++ {
		k1 := rng.Float64() < conv1
		k2 := rng.Float64() < conv2
		if k1 {
			scored1++
		}
		if k2 {
			scored2++
		}
		kicks = append(kicks,
			PenaltyKick{Team: team1.ID, KickNo: i, Scored: k1},
			PenaltyKick{Team: team2.ID, KickNo: i, Scored: k2},
		)
	}

	// Sudden death
	for kickNo := 6; scored1 == scored2 && kickNo <= maxPenRounds+5; kickNo++ {
		k1 := rng.Float64() < conv1
		k2 := rng.Float64() < conv2
		if k1 {
			scored1++
		}
		if k2 {
			scored2++
		}
		kicks = append(kicks,
			PenaltyKick{Team: team1.ID, KickNo: kickNo, Scored: k1},
			PenaltyKick{Team: team2.ID, KickNo: kickNo, Scored: k2},
		)
	}

	return PenaltyResult{Team1Scored: scored1, Team2Scored: scored2, Kicks: kicks}
}
