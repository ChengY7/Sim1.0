package sim_test

import (
	"testing"

	"github.com/chengyang/sim1.0/backend/internal/config"
	"github.com/chengyang/sim1.0/backend/internal/sim"
)

func loadBundle(t *testing.T) *config.Bundle {
	t.Helper()
	b, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	return b
}

func newEngine(t *testing.T, home, away string, seed int64) *sim.Engine {
	t.Helper()
	e, err := sim.NewEngine(loadBundle(t), home, away, seed)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	return e
}

// --- PeriodLabel ---

func TestPeriodLabel(t *testing.T) {
	cases := []struct {
		quarter, quarters int
		want              string
	}{
		{1, 4, "Q1"},
		{2, 4, "Q2"},
		{3, 4, "Q3"},
		{4, 4, "Q4"},
		{5, 4, "OT"},
		{6, 4, "2OT"},
		{7, 4, "3OT"},
		{8, 4, "4OT"},
		{9, 4, "5OT"},
		{14, 4, "10OT"},
	}
	for _, c := range cases {
		got := sim.PeriodLabel(c.quarter, c.quarters)
		if got != c.want {
			t.Errorf("PeriodLabel(%d, %d) = %q, want %q", c.quarter, c.quarters, got, c.want)
		}
	}
}

// --- NewEngine ---

func TestNewEngine_UnknownTeam(t *testing.T) {
	b := loadBundle(t)
	_, err := sim.NewEngine(b, "LAL", "FAKE", 42)
	if err == nil {
		t.Fatal("expected error for unknown away team")
	}
	_, err = sim.NewEngine(b, "FAKE", "BOS", 42)
	if err == nil {
		t.Fatal("expected error for unknown home team")
	}
}

// --- RunUntilFinal ---

func TestRunUntilFinal_Completes(t *testing.T) {
	result := newEngine(t, "LAL", "BOS", 42).RunUntilFinal()

	if result.State.Status != "final" {
		t.Errorf("status = %q, want final", result.State.Status)
	}
	if result.Truncated {
		t.Error("game should not be truncated with a normal seed")
	}
	if len(result.Events) == 0 {
		t.Error("expected at least one event")
	}
	if result.State.HomeScore < 0 || result.State.AwayScore < 0 {
		t.Error("scores must be non-negative")
	}
}

func TestRunUntilFinal_Deterministic(t *testing.T) {
	run := func() sim.Result { return newEngine(t, "LAL", "BOS", 42).RunUntilFinal() }
	r1, r2 := run(), run()
	if r1.State.HomeScore != r2.State.HomeScore || r1.State.AwayScore != r2.State.AwayScore {
		t.Errorf("same seed produced different scores: %d-%d vs %d-%d",
			r1.State.HomeScore, r1.State.AwayScore,
			r2.State.HomeScore, r2.State.AwayScore)
	}
	if len(r1.Events) != len(r2.Events) {
		t.Errorf("same seed produced different event counts: %d vs %d", len(r1.Events), len(r2.Events))
	}
}

func TestRunUntilFinal_ScoresMatchEvents(t *testing.T) {
	result := newEngine(t, "LAL", "BOS", 42).RunUntilFinal()

	var home, away int
	for _, ev := range result.Events {
		if ev.Team == sim.Home {
			home += ev.Points
		} else {
			away += ev.Points
		}
	}
	if home != result.State.HomeScore {
		t.Errorf("home score from events %d != state %d", home, result.State.HomeScore)
	}
	if away != result.State.AwayScore {
		t.Errorf("away score from events %d != state %d", away, result.State.AwayScore)
	}
}

func TestRunUntilFinal_OT(t *testing.T) {
	b := loadBundle(t)
	result, _ := sim.NewEngine(b, "LAL", "BOS", 16)
	r := result.RunUntilFinal()

	if r.State.Quarter != b.Game.Quarters+1 {
		t.Errorf("expected OT (quarter %d), got %d", b.Game.Quarters+1, r.State.Quarter)
	}
	if label := sim.PeriodLabel(r.State.Quarter, b.Game.Quarters); label != "OT" {
		t.Errorf("period label = %q, want OT", label)
	}
	if r.State.HomeScore == r.State.AwayScore {
		t.Error("OT game must not end tied")
	}
}

func TestRunUntilFinal_2OT(t *testing.T) {
	b := loadBundle(t)
	result, _ := sim.NewEngine(b, "LAL", "BOS", 679)
	r := result.RunUntilFinal()

	if r.State.Quarter != b.Game.Quarters+2 {
		t.Errorf("expected 2OT (quarter %d), got %d", b.Game.Quarters+2, r.State.Quarter)
	}
	if label := sim.PeriodLabel(r.State.Quarter, b.Game.Quarters); label != "2OT" {
		t.Errorf("period label = %q, want 2OT", label)
	}
	if r.State.HomeScore == r.State.AwayScore {
		t.Error("2OT game must not end tied")
	}
}
