package sim

import (
	"math/rand"

	"github.com/chengyang/sim1.0/backend/internal/nba/config"
)

// playoffHomeSchedule encodes home-court for games 1-7.
// true = series home team plays at home, false = plays away.
var playoffHomeSchedule = [7]bool{true, true, false, false, true, false, true}

type PlayoffTeam struct {
	TeamID  string
	Seed    int
	SeasonW int
	SeasonL int
}

type PlayoffSeriesState struct {
	HomeTeam string // series home team (hosts G1,G2,G5,G7); empty = derive from bracket
	AwayTeam string // empty = derive from bracket
	HomeWins int
	AwayWins int
}

type PlayoffBracketState struct {
	EastR1 [4]PlayoffSeriesState
	EastR2 [2]PlayoffSeriesState
	EastR3 PlayoffSeriesState
	WestR1 [4]PlayoffSeriesState
	WestR2 [2]PlayoffSeriesState
	WestR3 PlayoffSeriesState
	Finals PlayoffSeriesState
}

type PlayoffInput struct {
	East    [8]PlayoffTeam
	West    [8]PlayoffTeam
	Bracket PlayoffBracketState
}

type PlayoffGameResult struct {
	GameNum   int
	Home      string
	Away      string
	HomeScore int
	AwayScore int
	Winner    string
}

type PlayoffSeriesResult struct {
	HomeTeam string
	AwayTeam string
	HomeWins int
	AwayWins int
	Winner   string
	Games    []PlayoffGameResult
}

type ConferencePlayoffResult struct {
	R1       [4]PlayoffSeriesResult
	R2       [2]PlayoffSeriesResult
	R3       PlayoffSeriesResult
	Champion string
}

type PlayoffResult struct {
	East     ConferencePlayoffResult
	West     ConferencePlayoffResult
	Finals   PlayoffSeriesResult
	Champion string
}

func SimulatePlayoffs(cfg *config.Bundle, input PlayoffInput, seed int64) PlayoffResult {
	rng := rand.New(rand.NewSource(seed))

	seedMap := func(teams [8]PlayoffTeam) map[string]int {
		m := make(map[string]int, 8)
		for _, t := range teams {
			m[t.TeamID] = t.Seed
		}
		return m
	}

	eastResult := simulateConference(cfg, input.East, input.Bracket.EastR1, input.Bracket.EastR2, input.Bracket.EastR3, seedMap(input.East), rng)
	westResult := simulateConference(cfg, input.West, input.Bracket.WestR1, input.Bracket.WestR2, input.Bracket.WestR3, seedMap(input.West), rng)

	finalsState := input.Bracket.Finals
	alreadyDecided := finalsState.HomeWins == 4 || finalsState.AwayWins == 4
	if finalsState.HomeTeam == "" && !alreadyDecided {
		finalsState.HomeTeam, finalsState.AwayTeam = finalsHome(
			eastResult.Champion, westResult.Champion,
			input.East[:], input.West[:],
			rng,
		)
	}

	finalsResult := simulateSeries(cfg, finalsState, rng)

	return PlayoffResult{
		East:     eastResult,
		West:     westResult,
		Finals:   finalsResult,
		Champion: finalsResult.Winner,
	}
}

func simulateConference(
	cfg *config.Bundle,
	teams [8]PlayoffTeam,
	r1States [4]PlayoffSeriesState,
	r2States [2]PlayoffSeriesState,
	r3State PlayoffSeriesState,
	seeds map[string]int,
	rng *rand.Rand,
) ConferencePlayoffResult {
	bySeed := make(map[int]string, 8)
	for _, t := range teams {
		bySeed[t.Seed] = t.TeamID
	}

	// R1: 1v8, 4v5, 3v6, 2v7 — lower seed is always home
	r1Pairs := [4][2]int{{1, 8}, {4, 5}, {3, 6}, {2, 7}}
	var r1 [4]PlayoffSeriesResult
	for i, pair := range r1Pairs {
		s := r1States[i]
		if s.HomeTeam == "" {
			s.HomeTeam = bySeed[pair[0]]
			s.AwayTeam = bySeed[pair[1]]
		}
		r1[i] = simulateSeries(cfg, s, rng)
	}

	// R2: winner(1v8) vs winner(4v5), winner(3v6) vs winner(2v7)
	var r2 [2]PlayoffSeriesResult
	for i, pair := range [2][2]int{{0, 1}, {2, 3}} {
		s := r2States[i]
		if s.HomeTeam == "" {
			s.HomeTeam, s.AwayTeam = lowerSeedFirst(r1[pair[0]].Winner, r1[pair[1]].Winner, seeds)
		}
		r2[i] = simulateSeries(cfg, s, rng)
	}

	// R3: Conference Finals
	if r3State.HomeTeam == "" {
		r3State.HomeTeam, r3State.AwayTeam = lowerSeedFirst(r2[0].Winner, r2[1].Winner, seeds)
	}
	r3 := simulateSeries(cfg, r3State, rng)

	return ConferencePlayoffResult{
		R1:       r1,
		R2:       r2,
		R3:       r3,
		Champion: r3.Winner,
	}
}

func simulateSeries(cfg *config.Bundle, state PlayoffSeriesState, rng *rand.Rand) PlayoffSeriesResult {
	home := state.HomeTeam
	away := state.AwayTeam
	homeWins := state.HomeWins
	awayWins := state.AwayWins

	var games []PlayoffGameResult

	for homeWins < 4 && awayWins < 4 {
		gameNum := homeWins + awayWins + 1
		homeIsHome := playoffHomeSchedule[gameNum-1]

		var gameHome, gameAway string
		if homeIsHome {
			gameHome, gameAway = home, away
		} else {
			gameHome, gameAway = away, home
		}

		gr := runPlayoffGame(cfg, gameHome, gameAway, gameNum, rng)
		games = append(games, gr)

		if gr.Winner == home {
			homeWins++
		} else {
			awayWins++
		}
	}

	winner := away
	if homeWins == 4 {
		winner = home
	}

	return PlayoffSeriesResult{
		HomeTeam: home,
		AwayTeam: away,
		HomeWins: homeWins,
		AwayWins: awayWins,
		Winner:   winner,
		Games:    games,
	}
}

func runPlayoffGame(cfg *config.Bundle, home, away string, gameNum int, rng *rand.Rand) PlayoffGameResult {
	eng, err := NewEngine(cfg, home, away, rng.Int63())
	if err != nil {
		return PlayoffGameResult{GameNum: gameNum, Home: home, Away: away, HomeScore: 100, AwayScore: 99, Winner: home}
	}
	res := eng.RunUntilFinal()
	winner := home
	if res.State.AwayScore > res.State.HomeScore {
		winner = away
	}
	return PlayoffGameResult{
		GameNum:   gameNum,
		Home:      home,
		Away:      away,
		HomeScore: res.State.HomeScore,
		AwayScore: res.State.AwayScore,
		Winner:    winner,
	}
}

// lowerSeedFirst returns (home, away) where home has the lower (better) seed.
func lowerSeedFirst(t1, t2 string, seeds map[string]int) (string, string) {
	if seeds[t1] < seeds[t2] {
		return t1, t2
	}
	return t2, t1
}

// finalsHome picks which conference champion hosts the NBA Finals.
// The team with more regular season wins hosts; coin flip on a tie.
func finalsHome(eastChamp, westChamp string, east, west []PlayoffTeam, rng *rand.Rand) (home, away string) {
	var eastW, westW int
	for _, t := range east {
		if t.TeamID == eastChamp {
			eastW = t.SeasonW
		}
	}
	for _, t := range west {
		if t.TeamID == westChamp {
			westW = t.SeasonW
		}
	}
	if eastW > westW || (eastW == westW && rng.Intn(2) == 0) {
		return eastChamp, westChamp
	}
	return westChamp, eastChamp
}
