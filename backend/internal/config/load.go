package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Team struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Offense float64 `json:"offense"`
	Defense float64 `json:"defense"`
}

type Outcome struct {
	Type   string  `json:"type"`
	Points int     `json:"points"`
	Weight float64 `json:"weight"`
}

type OutcomesFile struct {
	Outcomes []Outcome `json:"outcomes"`
}

// Game holds clock and pace. Pace = possessions per 48 min per team (NBA-style).
type Game struct {
	Quarters        int     `json:"quarters"`
	QuarterSeconds  int     `json:"quarter_seconds"`
	Pace            float64 `json:"pace"`
	TickJitterSec   int     `json:"tick_jitter_sec"`
}

func (g Game) GameSeconds() float64 {
	return float64(g.Quarters * g.QuarterSeconds)
}

// ExpectedTotalPossessions is both teams combined (~2 * pace).
func (g Game) ExpectedTotalPossessions() float64 {
	return 2 * g.Pace
}

// SecondsPerPossession spreads game clock across all offensive possessions.
func (g Game) SecondsPerPossession() float64 {
	return g.GameSeconds() / g.ExpectedTotalPossessions()
}

type Bundle struct {
	Teams    map[string]Team
	Outcomes []Outcome
	Game     Game
}

func Load(dir string) (*Bundle, error) {
	teamsPath := filepath.Join(dir, "teams.json")
	outcomesPath := filepath.Join(dir, "outcomes.json")
	gamePath := filepath.Join(dir, "game.json")

	teamsData, err := os.ReadFile(teamsPath)
	if err != nil {
		return nil, fmt.Errorf("read teams: %w", err)
	}
	var teams []Team
	if err := json.Unmarshal(teamsData, &teams); err != nil {
		return nil, fmt.Errorf("parse teams: %w", err)
	}

	outcomesData, err := os.ReadFile(outcomesPath)
	if err != nil {
		return nil, fmt.Errorf("read outcomes: %w", err)
	}
	var of OutcomesFile
	if err := json.Unmarshal(outcomesData, &of); err != nil {
		return nil, fmt.Errorf("parse outcomes: %w", err)
	}

	gameData, err := os.ReadFile(gamePath)
	if err != nil {
		return nil, fmt.Errorf("read game: %w", err)
	}
	var game Game
	if err := json.Unmarshal(gameData, &game); err != nil {
		return nil, fmt.Errorf("parse game: %w", err)
	}
	if game.Pace <= 0 {
		game.Pace = 100
	}
	if game.Quarters <= 0 {
		game.Quarters = 4
	}
	if game.QuarterSeconds <= 0 {
		game.QuarterSeconds = 720
	}

	byID := make(map[string]Team, len(teams))
	for _, t := range teams {
		byID[t.ID] = t
	}

	return &Bundle{
		Teams:    byID,
		Outcomes: of.Outcomes,
		Game:     game,
	}, nil
}

func (b *Bundle) Team(id string) (Team, error) {
	t, ok := b.Teams[id]
	if !ok {
		return Team{}, fmt.Errorf("unknown team id %q", id)
	}
	return t, nil
}
