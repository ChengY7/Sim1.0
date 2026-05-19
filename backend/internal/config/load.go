package config

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"

	_ "embed"

	"embed"
)

//go:embed data
var defaultFS embed.FS

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
	Quarters          int     `json:"quarters"`
	QuarterSeconds    int     `json:"quarter_seconds"`
	Pace              float64 `json:"pace"`
	TickJitterSec     int     `json:"tick_jitter_sec"`
	FreeThrowPct      float64 `json:"free_throw_pct"`
	FreeThrowsPerFoul int     `json:"free_throws_per_foul"`
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

// Load returns a Bundle using the configs embedded at build time.
func Load() (*Bundle, error) {
	sub, err := fs.Sub(defaultFS, "data")
	if err != nil {
		return nil, err
	}
	return load(sub)
}

// LoadDir returns a Bundle from JSON files in dir, overriding the embedded defaults.
func LoadDir(dir string) (*Bundle, error) {
	return load(os.DirFS(dir))
}

func load(fsys fs.FS) (*Bundle, error) {
	teamsData, err := fs.ReadFile(fsys, "teams.json")
	if err != nil {
		return nil, fmt.Errorf("read teams: %w", err)
	}
	var teams []Team
	if err := json.Unmarshal(teamsData, &teams); err != nil {
		return nil, fmt.Errorf("parse teams: %w", err)
	}

	outcomesData, err := fs.ReadFile(fsys, "outcomes.json")
	if err != nil {
		return nil, fmt.Errorf("read outcomes: %w", err)
	}
	var of OutcomesFile
	if err := json.Unmarshal(outcomesData, &of); err != nil {
		return nil, fmt.Errorf("parse outcomes: %w", err)
	}

	gameData, err := fs.ReadFile(fsys, "game.json")
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
	if game.FreeThrowPct <= 0 {
		game.FreeThrowPct = 0.75
	}
	if game.FreeThrowsPerFoul <= 0 {
		game.FreeThrowsPerFoul = 2
	}

	byID := make(map[string]Team, len(teams))
	for _, t := range teams {
		if t.Offense <= 0 || t.Defense <= 0 {
			return nil, fmt.Errorf("team %q: offense and defense must be > 0", t.ID)
		}
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
