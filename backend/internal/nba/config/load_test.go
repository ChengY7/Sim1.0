package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/chengyang/sim1.0/backend/internal/nba/config"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// minTeam is the minimal valid team entry for test fixtures.
const minTeam = `[{"id":"X","name":"X","conference":"east","division":"atlantic"}]`

// minOutcomes is a minimal valid outcomes.json.
const minOutcomes = `{"outcomes":[{"type":"make_2pt","points":2,"weight":1}]}`

// minSeason is a valid season file for a single team "X".
const minSeason = `{"season":"2099-00","teams":[{"id":"X","offense":1.0,"defense":1.0}]}`

func TestLoad_Embedded(t *testing.T) {
	b, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(b.Teams) == 0 {
		t.Error("expected teams")
	}
	if len(b.Outcomes) == 0 {
		t.Error("expected outcomes")
	}
	if b.Game.Quarters != 4 {
		t.Errorf("quarters = %d, want 4", b.Game.Quarters)
	}
	if b.Game.Pace <= 0 {
		t.Errorf("pace = %g, want > 0", b.Game.Pace)
	}
}

func TestLoad_NegativeOutcomeWeight(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "teams.json"), minTeam)
	writeFile(t, filepath.Join(dir, "outcomes.json"),
		`{"outcomes":[{"type":"make_2pt","points":2,"weight":-1}]}`)
	writeFile(t, filepath.Join(dir, "game.json"), `{}`)

	_, err := config.LoadDir(dir)
	if err == nil {
		t.Fatal("expected error for negative outcome weight")
	}
}

func TestLoad_InvalidFreeThrowPct(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "teams.json"), minTeam)
	writeFile(t, filepath.Join(dir, "outcomes.json"), minOutcomes)
	writeFile(t, filepath.Join(dir, "game.json"), `{"free_throw_pct":1.5}`)
	writeFile(t, filepath.Join(dir, "seasons", "nba_2099-00.json"), minSeason)

	b, err := config.LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}
	if b.Game.FreeThrowPct != 0.75 {
		t.Errorf("free_throw_pct = %g, want default 0.75", b.Game.FreeThrowPct)
	}
}

func TestLoad_InvalidSeasonDefense(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "teams.json"), minTeam)
	writeFile(t, filepath.Join(dir, "outcomes.json"), minOutcomes)
	writeFile(t, filepath.Join(dir, "game.json"), `{}`)
	writeFile(t, filepath.Join(dir, "seasons", "nba_2099-00.json"),
		`{"season":"2099-00","teams":[{"id":"X","offense":1.0,"defense":0}]}`)

	_, err := config.LoadDir(dir)
	if err == nil {
		t.Fatal("expected error for zero defense in season file")
	}
}

func TestLoad_InvalidSeasonOffense(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "teams.json"), minTeam)
	writeFile(t, filepath.Join(dir, "outcomes.json"), minOutcomes)
	writeFile(t, filepath.Join(dir, "game.json"), `{}`)
	writeFile(t, filepath.Join(dir, "seasons", "nba_2099-00.json"),
		`{"season":"2099-00","teams":[{"id":"X","offense":0,"defense":1.0}]}`)

	_, err := config.LoadDir(dir)
	if err == nil {
		t.Fatal("expected error for zero offense in season file")
	}
}

func TestLoad_Defaults(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "teams.json"), minTeam)
	writeFile(t, filepath.Join(dir, "outcomes.json"), minOutcomes)
	writeFile(t, filepath.Join(dir, "game.json"), `{}`)
	writeFile(t, filepath.Join(dir, "seasons", "nba_2099-00.json"), minSeason)

	b, err := config.LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}
	if b.Game.Quarters != 4 {
		t.Errorf("quarters default = %d, want 4", b.Game.Quarters)
	}
	if b.Game.Pace != 100 {
		t.Errorf("pace default = %g, want 100", b.Game.Pace)
	}
	if b.Game.OTSeconds != 300 {
		t.Errorf("ot_seconds default = %d, want 300", b.Game.OTSeconds)
	}
	if b.Game.FreeThrowPct != 0.75 {
		t.Errorf("free_throw_pct default = %g, want 0.75", b.Game.FreeThrowPct)
	}
	if b.Game.FreeThrowsPerFoul != 2 {
		t.Errorf("free_throws_per_foul default = %d, want 2", b.Game.FreeThrowsPerFoul)
	}
	if b.Game.QuarterSeconds != 720 {
		t.Errorf("quarter_seconds default = %d, want 720", b.Game.QuarterSeconds)
	}
	if b.Game.TickJitterSec != 0 {
		t.Errorf("tick_jitter_sec default = %d, want 0", b.Game.TickJitterSec)
	}
}

func TestLoad_InvalidOffenseScale(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "teams.json"), minTeam)
	writeFile(t, filepath.Join(dir, "outcomes.json"),
		`{"outcomes":[{"type":"make_2pt","points":2,"weight":1,"offense_scale":"diagonal"}]}`)
	writeFile(t, filepath.Join(dir, "game.json"), `{}`)

	_, err := config.LoadDir(dir)
	if err == nil {
		t.Fatal("expected error for invalid offense_scale")
	}
}

func TestLoad_EmptyTeamID(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "teams.json"),
		`[{"id":"","name":"X","conference":"east","division":"atlantic"}]`)
	writeFile(t, filepath.Join(dir, "outcomes.json"), minOutcomes)
	writeFile(t, filepath.Join(dir, "game.json"), `{}`)

	_, err := config.LoadDir(dir)
	if err == nil {
		t.Fatal("expected error for empty team id")
	}
}

func TestLoad_EmptyOutcomes(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "teams.json"), minTeam)
	writeFile(t, filepath.Join(dir, "outcomes.json"), `{"outcomes":[]}`)
	writeFile(t, filepath.Join(dir, "game.json"), `{}`)

	_, err := config.LoadDir(dir)
	if err == nil {
		t.Fatal("expected error for empty outcomes")
	}
}

func TestLoad_AvailableSeasons(t *testing.T) {
	b, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(b.AvailableSeasons) == 0 {
		t.Error("expected at least one season")
	}
	if b.DefaultSeason == "" {
		t.Error("expected a default season")
	}
	// Newest season should be first.
	if len(b.AvailableSeasons) > 1 {
		if b.AvailableSeasons[0] < b.AvailableSeasons[1] {
			t.Errorf("seasons not sorted descending: %v", b.AvailableSeasons)
		}
	}
}

func TestLoad_WithSeason(t *testing.T) {
	b, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	for _, s := range b.AvailableSeasons {
		nb, err := b.WithSeason(s)
		if err != nil {
			t.Errorf("WithSeason(%q): %v", s, err)
			continue
		}
		for _, team := range nb.Teams {
			if team.Offense <= 0 {
				t.Errorf("season %q team %q: offense = %g, want > 0", s, team.ID, team.Offense)
			}
			if team.Defense <= 0 {
				t.Errorf("season %q team %q: defense = %g, want > 0", s, team.ID, team.Defense)
			}
		}
	}
}

func TestSortedTeams(t *testing.T) {
	b, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	teams := b.SortedTeams()
	if len(teams) != len(b.Teams) {
		t.Errorf("SortedTeams len = %d, want %d", len(teams), len(b.Teams))
	}
	for i := 1; i < len(teams); i++ {
		if teams[i].ID < teams[i-1].ID {
			t.Errorf("not sorted: %q before %q", teams[i-1].ID, teams[i].ID)
		}
	}
}

func TestTeam_UnknownID(t *testing.T) {
	b, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	_, err = b.Team("FAKE")
	if err == nil {
		t.Fatal("expected error for unknown team id")
	}
}

func TestTeam_KnownID(t *testing.T) {
	b, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	team, err := b.Team("LAL")
	if err != nil {
		t.Fatalf("Team(LAL): %v", err)
	}
	if team.ID != "LAL" {
		t.Errorf("id = %q, want LAL", team.ID)
	}
}
