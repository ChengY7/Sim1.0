package config

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
)

//go:embed data
var defaultFS embed.FS

type Team struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Group string `json:"group"`
	Elo   int    `json:"elo"`
	Host  bool   `json:"host"`
}

// Match is one group-stage fixture.
type Match struct {
	Date     string `json:"date"`
	Matchday int    `json:"matchday"`
	Group    string `json:"group"`
	Team1    string `json:"team1"`
	Team2    string `json:"team2"`
}

type Bundle struct {
	Teams   map[string]Team
	Matches []Match
}

func Load() (*Bundle, error) {
	sub, err := fs.Sub(defaultFS, "data")
	if err != nil {
		return nil, err
	}
	return load(sub)
}

func load(fsys fs.FS) (*Bundle, error) {
	data, err := fs.ReadFile(fsys, "wc2026_teams.json")
	if err != nil {
		return nil, fmt.Errorf("read wc2026_teams: %w", err)
	}

	var raw struct {
		Teams []Team `json:"teams"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse wc2026_teams: %w", err)
	}

	byID := make(map[string]Team, len(raw.Teams))
	for _, t := range raw.Teams {
		if t.ID == "" || t.Name == "" {
			return nil, fmt.Errorf("team missing id or name")
		}
		byID[t.ID] = t
	}

	schedData, err := fs.ReadFile(fsys, "wc2026_schedule.json")
	if err != nil {
		return nil, fmt.Errorf("read wc2026_schedule: %w", err)
	}
	var rawSched struct {
		Matches []Match `json:"matches"`
	}
	if err := json.Unmarshal(schedData, &rawSched); err != nil {
		return nil, fmt.Errorf("parse wc2026_schedule: %w", err)
	}
	for _, m := range rawSched.Matches {
		if _, ok := byID[m.Team1]; !ok {
			return nil, fmt.Errorf("schedule: unknown team %q", m.Team1)
		}
		if _, ok := byID[m.Team2]; !ok {
			return nil, fmt.Errorf("schedule: unknown team %q", m.Team2)
		}
	}

	return &Bundle{Teams: byID, Matches: rawSched.Matches}, nil
}

func (b *Bundle) Team(id string) (Team, error) {
	t, ok := b.Teams[id]
	if !ok {
		return Team{}, fmt.Errorf("unknown FIFA team id %q", id)
	}
	return t, nil
}

func (b *Bundle) SortedTeams() []Team {
	out := make([]Team, 0, len(b.Teams))
	for _, t := range b.Teams {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
