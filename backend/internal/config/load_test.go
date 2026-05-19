package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/chengyang/sim1.0/backend/internal/config"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

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

func TestLoad_InvalidDefense(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "teams.json"),
		`[{"id":"X","name":"X","offense":1.0,"defense":0}]`)
	writeFile(t, filepath.Join(dir, "outcomes.json"),
		`{"outcomes":[{"type":"make_2pt","points":2,"weight":1}]}`)
	writeFile(t, filepath.Join(dir, "game.json"), `{}`)

	_, err := config.LoadDir(dir)
	if err == nil {
		t.Fatal("expected error for zero defense")
	}
}

func TestLoad_InvalidOffense(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "teams.json"),
		`[{"id":"X","name":"X","offense":0,"defense":1.0}]`)
	writeFile(t, filepath.Join(dir, "outcomes.json"),
		`{"outcomes":[{"type":"make_2pt","points":2,"weight":1}]}`)
	writeFile(t, filepath.Join(dir, "game.json"), `{}`)

	_, err := config.LoadDir(dir)
	if err == nil {
		t.Fatal("expected error for zero offense")
	}
}

func TestLoad_Defaults(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "teams.json"),
		`[{"id":"X","name":"X","offense":1.0,"defense":1.0}]`)
	writeFile(t, filepath.Join(dir, "outcomes.json"),
		`{"outcomes":[{"type":"make_2pt","points":2,"weight":1}]}`)
	// empty game.json — all fields should fall back to defaults
	writeFile(t, filepath.Join(dir, "game.json"), `{}`)

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
}

func TestTeam_UnknownID(t *testing.T) {
	b, _ := config.Load()
	_, err := b.Team("FAKE")
	if err == nil {
		t.Fatal("expected error for unknown team id")
	}
}

func TestTeam_KnownID(t *testing.T) {
	b, _ := config.Load()
	team, err := b.Team("LAL")
	if err != nil {
		t.Fatalf("Team(LAL): %v", err)
	}
	if team.ID != "LAL" {
		t.Errorf("id = %q, want LAL", team.ID)
	}
}
