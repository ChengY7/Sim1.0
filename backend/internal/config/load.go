package config

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"sort"
)

//go:embed data
var defaultFS embed.FS

type Team struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Offense float64 `json:"offense"`
	Defense float64 `json:"defense"`
}

// Outcome represents one possible possession result.
// OffenseScale controls how the outcome weight shifts with the offense/defense matchup:
//   "up"   — good offense increases this outcome (scoring plays)
//   "down" — good offense decreases this outcome (misses, turnovers)
//   ""     — neutral; weight is unchanged (fouls)
type Outcome struct {
	Type         string  `json:"type"`
	Points       int     `json:"points"`
	Weight       float64 `json:"weight"`
	OffenseScale string  `json:"offense_scale"`
}

type outcomesFile struct {
	Outcomes []Outcome `json:"outcomes"`
}

// Game holds clock and pace. Pace = possessions per 48 min per team (NBA-style).
type Game struct {
	Quarters          int     `json:"quarters"`
	QuarterSeconds    int     `json:"quarter_seconds"`
	OTSeconds         int     `json:"ot_seconds"`
	Pace              float64 `json:"pace"`
	TickJitterSec     int     `json:"tick_jitter_sec"`
	FreeThrowPct      float64 `json:"free_throw_pct"`
	FreeThrowsPerFoul int     `json:"free_throws_per_foul"`
}

func (g Game) gameSeconds() float64 {
	return float64(g.Quarters * g.QuarterSeconds)
}

// ExpectedTotalPossessions is both teams combined (~2 * pace).
func (g Game) ExpectedTotalPossessions() float64 {
	return 2 * g.Pace
}

// SecondsPerPossession spreads game clock across all offensive possessions.
func (g Game) SecondsPerPossession() float64 {
	return g.gameSeconds() / g.ExpectedTotalPossessions()
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
	var of outcomesFile
	if err := json.Unmarshal(outcomesData, &of); err != nil {
		return nil, fmt.Errorf("parse outcomes: %w", err)
	}
	if len(of.Outcomes) == 0 {
		return nil, fmt.Errorf("outcomes.json must define at least one outcome")
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
	if game.OTSeconds <= 0 {
		game.OTSeconds = 300
	}
	if game.FreeThrowPct <= 0 {
		game.FreeThrowPct = 0.75
	}
	if game.FreeThrowsPerFoul <= 0 {
		game.FreeThrowsPerFoul = 2
	}
	if game.TickJitterSec < 0 {
		game.TickJitterSec = 0
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

// SortedTeams returns all teams sorted by ID.
func (b *Bundle) SortedTeams() []Team {
	out := make([]Team, 0, len(b.Teams))
	for _, t := range b.Teams {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
