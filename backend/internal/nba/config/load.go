package config

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"
)

//go:embed data
var defaultFS embed.FS

type Team struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Conference string  `json:"conference"` // "east" or "west"
	Division   string  `json:"division"`   // "atlantic","central","southeast","northwest","pacific","southwest"
	Offense    float64 // set from season file, not from teams.json
	Defense    float64 // set from season file, not from teams.json
}

// TeamRatings holds the per-season offensive and defensive multipliers for one team.
type TeamRatings struct {
	ID      string  `json:"id"`
	Offense float64 `json:"offense"`
	Defense float64 `json:"defense"`
}

// TeamRecord holds the actual end-of-season W/L for one team.
type TeamRecord struct {
	ID string `json:"id"`
	W  int    `json:"w"`
	L  int    `json:"l"`
}

// seasonFile is the on-disk format for a season ratings file.
type seasonFile struct {
	Season    string        `json:"season"`
	Teams     []TeamRatings `json:"teams"`
	Standings []TeamRecord  `json:"standings,omitempty"`
}

// Outcome represents one possible possession result.
// OffenseScale controls how the outcome weight shifts with the offense/defense matchup:
//
//	"up"   — good offense increases this outcome (scoring plays)
//	"down" — good offense decreases this outcome (misses, turnovers)
//	""     — neutral; weight is unchanged (fouls)
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

// CupGroup defines one NBA Cup group-stage group.
type CupGroup struct {
	Name       string   `json:"name"`
	Conference string   `json:"conference"`
	Teams      []string `json:"teams"`
}

// ScheduleGame is one regular-season game from schedule.json.
type ScheduleGame struct {
	Date     string `json:"date"`
	TimeET   string `json:"time_et"`
	Away     string `json:"away"`
	Home     string `json:"home"`
	NbaCup   bool   `json:"nba_cup"`
	Excluded bool   `json:"excluded"`
}

// Schedule is the top-level structure of schedule.json.
type Schedule struct {
	Season string         `json:"season"`
	Games  []ScheduleGame `json:"games"`
}

// DraftLottery holds the ball-combination counts for the NBA draft lottery.
// Combinations[i] is the number of combinations assigned to seed i+1 (0-indexed).
type DraftLottery struct {
	Combinations [14]int `json:"combinations"`
}

type Bundle struct {
	Teams            map[string]Team
	Outcomes         []Outcome
	Game             Game
	Schedule         Schedule
	CupGroups        []CupGroup
	DraftLottery     DraftLottery
	Seasons          map[string]map[string]TeamRatings // season → teamID → ratings
	SeasonStandings  map[string][]TeamRecord           // season → actual end-of-season W/L
	AvailableSeasons []string                          // sorted descending (newest first)
	DefaultSeason    string
}

// Load returns a Bundle using the configs embedded at build time.
func Load() (*Bundle, error) {
	sub, err := fs.Sub(defaultFS, "data")
	if err != nil {
		return nil, err
	}
	return load(sub)
}

// LoadDir returns a Bundle from JSON files in dir (not the embedded defaults).
func LoadDir(dir string) (*Bundle, error) {
	return load(os.DirFS(dir))
}

func load(fsys fs.FS) (*Bundle, error) {
	teamsData, err := fs.ReadFile(fsys, "teams.json")
	if err != nil {
		return nil, fmt.Errorf("read teams: %w", err)
	}
	var rawTeams []struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Conference string `json:"conference"`
		Division   string `json:"division"`
	}
	if err := json.Unmarshal(teamsData, &rawTeams); err != nil {
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
	for _, o := range of.Outcomes {
		if o.Weight <= 0 {
			return nil, fmt.Errorf("outcome %q: weight must be > 0, got %g", o.Type, o.Weight)
		}
		if o.OffenseScale != "" && o.OffenseScale != "up" && o.OffenseScale != "down" {
			return nil, fmt.Errorf("outcome %q: offense_scale %q must be \"\", \"up\", or \"down\"", o.Type, o.OffenseScale)
		}
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
	if game.FreeThrowPct <= 0 || game.FreeThrowPct > 1 {
		game.FreeThrowPct = 0.75
	}
	if game.FreeThrowsPerFoul <= 0 {
		game.FreeThrowsPerFoul = 2
	}
	if game.TickJitterSec < 0 {
		game.TickJitterSec = 0
	}

	// Load season rating files from seasons/.
	seasons := map[string]map[string]TeamRatings{}
	seasonStandings := map[string][]TeamRecord{}
	var availableSeasons []string
	if entries, err := fs.ReadDir(fsys, "seasons"); err == nil {
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || !strings.HasPrefix(name, "nba_") || !strings.HasSuffix(name, ".json") {
				continue
			}
			data, err := fs.ReadFile(fsys, "seasons/"+name)
			if err != nil {
				return nil, fmt.Errorf("read season file %s: %w", name, err)
			}
			var sf seasonFile
			if err := json.Unmarshal(data, &sf); err != nil {
				return nil, fmt.Errorf("parse season file %s: %w", name, err)
			}
			if sf.Season == "" {
				return nil, fmt.Errorf("season file %s: missing season field", name)
			}
			byTeam := make(map[string]TeamRatings, len(sf.Teams))
			for _, r := range sf.Teams {
				if r.Offense <= 0 || r.Defense <= 0 {
					return nil, fmt.Errorf("season %s team %q: offense and defense must be > 0", sf.Season, r.ID)
				}
				byTeam[r.ID] = r
			}
			seasons[sf.Season] = byTeam
			if len(sf.Standings) > 0 {
				seasonStandings[sf.Season] = sf.Standings
			}
			availableSeasons = append(availableSeasons, sf.Season)
		}
	}
	// Sort descending so newest season is first.
	sort.Sort(sort.Reverse(sort.StringSlice(availableSeasons)))

	defaultSeason := ""
	if len(availableSeasons) > 0 {
		defaultSeason = availableSeasons[0]
	}

	// Validate teams and apply default season ratings.
	byID := make(map[string]Team, len(rawTeams))
	for _, t := range rawTeams {
		if t.ID == "" || t.Name == "" {
			return nil, fmt.Errorf("team has empty id or name")
		}
		if t.Conference != "east" && t.Conference != "west" {
			return nil, fmt.Errorf("team %q: conference must be \"east\" or \"west\"", t.ID)
		}
		if t.Division == "" {
			return nil, fmt.Errorf("team %q: division is required", t.ID)
		}
		team := Team{ID: t.ID, Name: t.Name, Conference: t.Conference, Division: t.Division}
		if defaultSeason != "" {
			r, ok := seasons[defaultSeason][t.ID]
			if !ok {
				return nil, fmt.Errorf("season %q: missing ratings for team %q", defaultSeason, t.ID)
			}
			team.Offense = r.Offense
			team.Defense = r.Defense
		}
		byID[t.ID] = team
	}

	var schedule Schedule
	if scheduleData, err := fs.ReadFile(fsys, "schedule.json"); err == nil {
		if err := json.Unmarshal(scheduleData, &schedule); err != nil {
			return nil, fmt.Errorf("parse schedule: %w", err)
		}
	}

	var cupGroups []CupGroup
	if cgData, err := fs.ReadFile(fsys, "cup_groups.json"); err == nil {
		if err := json.Unmarshal(cgData, &cupGroups); err != nil {
			return nil, fmt.Errorf("parse cup_groups: %w", err)
		}
	}

	var draftLottery DraftLottery
	if dlData, err := fs.ReadFile(fsys, "draft_lottery.json"); err == nil {
		if err := json.Unmarshal(dlData, &draftLottery); err != nil {
			return nil, fmt.Errorf("parse draft_lottery: %w", err)
		}
	}

	return &Bundle{
		Teams:            byID,
		Outcomes:         of.Outcomes,
		Game:             game,
		Schedule:         schedule,
		CupGroups:        cupGroups,
		DraftLottery:     draftLottery,
		Seasons:          seasons,
		SeasonStandings:  seasonStandings,
		AvailableSeasons: availableSeasons,
		DefaultSeason:    defaultSeason,
	}, nil
}

// WithSeason returns a shallow copy of the Bundle with team ratings swapped to
// the given season. Returns an error if the season is unknown.
func (b *Bundle) WithSeason(season string) (*Bundle, error) {
	ratings, ok := b.Seasons[season]
	if !ok {
		return nil, fmt.Errorf("unknown season %q", season)
	}
	teams := make(map[string]Team, len(b.Teams))
	for id, t := range b.Teams {
		r, ok := ratings[id]
		if !ok {
			return nil, fmt.Errorf("season %q: no ratings for team %q", season, id)
		}
		t.Offense = r.Offense
		t.Defense = r.Defense
		teams[id] = t
	}
	copy := *b
	copy.Teams = teams
	return &copy, nil
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
