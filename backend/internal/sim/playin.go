package sim

import (
	"math/rand"

	"github.com/chengyang/sim1.0/backend/internal/config"
)

type PlayInInput struct {
	Seed7  string
	Seed8  string
	Seed9  string
	Seed10 string
}

type PlayInGameResult struct {
	Home      string
	Away      string
	HomeScore int
	AwayScore int
	Winner    string
	Loser     string
}

type ConferencePlayInResult struct {
	Game1    PlayInGameResult // 7 (home) vs 8 — winner = playoff 7 seed
	Game2    PlayInGameResult // 9 (home) vs 10 — loser eliminated
	Game3    PlayInGameResult // loser of G1 (home) vs winner of G2 — winner = playoff 8 seed
	Playoff7 string           // advances as 7th seed
	Playoff8 string           // advances as 8th seed
}

type PlayInResult struct {
	East ConferencePlayInResult
	West ConferencePlayInResult
}

func SimulatePlayIn(cfg *config.Bundle, east, west PlayInInput, seed int64) PlayInResult {
	rng := rand.New(rand.NewSource(seed))
	return PlayInResult{
		East: simulateConferencePlayIn(cfg, east, rng),
		West: simulateConferencePlayIn(cfg, west, rng),
	}
}

func simulateConferencePlayIn(cfg *config.Bundle, in PlayInInput, rng *rand.Rand) ConferencePlayInResult {
	g1 := runPlayInGame(cfg, in.Seed7, in.Seed8, rng)
	g2 := runPlayInGame(cfg, in.Seed9, in.Seed10, rng)
	g3 := runPlayInGame(cfg, g1.Loser, g2.Winner, rng)
	return ConferencePlayInResult{
		Game1:    g1,
		Game2:    g2,
		Game3:    g3,
		Playoff7: g1.Winner,
		Playoff8: g3.Winner,
	}
}

func runPlayInGame(cfg *config.Bundle, home, away string, rng *rand.Rand) PlayInGameResult {
	eng, err := NewEngine(cfg, home, away, rng.Int63())
	if err != nil {
		return PlayInGameResult{Home: home, Away: away, HomeScore: 100, AwayScore: 99, Winner: home, Loser: away}
	}
	res := eng.RunUntilFinal()
	winner, loser := home, away
	if res.State.AwayScore > res.State.HomeScore {
		winner, loser = away, home
	}
	return PlayInGameResult{
		Home:      home,
		Away:      away,
		HomeScore: res.State.HomeScore,
		AwayScore: res.State.AwayScore,
		Winner:    winner,
		Loser:     loser,
	}
}
