package sim_test

import (
	"testing"

	"github.com/chengyang/sim1.0/backend/internal/fifa/config"
	"github.com/chengyang/sim1.0/backend/internal/fifa/sim"
)

var (
	spain   = config.Team{ID: "ESP", Name: "Spain", Group: "H", Elo: 2165}
	france  = config.Team{ID: "FRA", Name: "France", Group: "I", Elo: 2081}
	usa     = config.Team{ID: "USA", Name: "USA", Group: "D", Elo: 1733, Host: true}
	newZeal = config.Team{ID: "NZL", Name: "New Zealand", Group: "G", Elo: 1585}
	equal1  = config.Team{ID: "T1", Name: "Team 1", Group: "A", Elo: 1800}
	equal2  = config.Team{ID: "T2", Name: "Team 2", Group: "A", Elo: 1800}
)

// ── Determinism ───────────────────────────────────────────────────────────────

func TestDeterminism(t *testing.T) {
	r1 := sim.SimulateMatch(spain, france, true, "", 42)
	r2 := sim.SimulateMatch(spain, france, true, "", 42)
	if r1.Team1Goals != r2.Team1Goals || r1.Team2Goals != r2.Team2Goals ||
		r1.Winner != r2.Winner || r1.Period != r2.Period {
		t.Errorf("same seed produced different results:\n  %+v\n  %+v", r1, r2)
	}
	// Compare AET goals by value, not pointer address.
	aet1Same := (r1.Team1GoalsAET == nil) == (r2.Team1GoalsAET == nil) &&
		(r1.Team1GoalsAET == nil || *r1.Team1GoalsAET == *r2.Team1GoalsAET)
	aet2Same := (r1.Team2GoalsAET == nil) == (r2.Team2GoalsAET == nil) &&
		(r1.Team2GoalsAET == nil || *r1.Team2GoalsAET == *r2.Team2GoalsAET)
	if !aet1Same || !aet2Same {
		t.Errorf("same seed produced different AET goals")
	}
	if r1.Penalties != nil && r2.Penalties != nil {
		if r1.Penalties.Team1Scored != r2.Penalties.Team1Scored ||
			r1.Penalties.Team2Scored != r2.Penalties.Team2Scored {
			t.Errorf("same seed produced different penalty scores")
		}
	}
}

func TestDifferentSeedsProduceDifferentResults(t *testing.T) {
	same := 0
	for seed := int64(0); seed < 20; seed++ {
		r1 := sim.SimulateMatch(spain, france, false, "", seed)
		r2 := sim.SimulateMatch(spain, france, false, "", seed+1000)
		if r1.Team1Goals == r2.Team1Goals && r1.Team2Goals == r2.Team2Goals {
			same++
		}
	}
	if same == 20 {
		t.Error("all 20 seed pairs produced identical scores — RNG is broken")
	}
}

// ── Period and winner correctness ─────────────────────────────────────────────

func TestPeriodAndWinnerConsistency(t *testing.T) {
	for seed := int64(0); seed < 500; seed++ {
		r := sim.SimulateMatch(spain, france, true, "", seed)

		// Period must be one of the valid values.
		if r.Period != "90" && r.Period != "aet" && r.Period != "pens" {
			t.Errorf("seed %d: invalid period %q", seed, r.Period)
		}

		// Winner must be one of the two teams or empty (group-stage draw).
		if r.Winner != "" && r.Winner != spain.ID && r.Winner != france.ID {
			t.Errorf("seed %d: winner %q is not a participant", seed, r.Winner)
		}

		// With extra_time=true a draw after 90+AET must go to penalties — no empty winner.
		if r.Winner == "" {
			t.Errorf("seed %d: extra_time=true produced no winner (period=%s)", seed, r.Period)
		}

		// Goals must be non-negative.
		if r.Team1Goals < 0 || r.Team2Goals < 0 {
			t.Errorf("seed %d: negative goals (%d, %d)", seed, r.Team1Goals, r.Team2Goals)
		}
	}
}

func TestGroupStageDrawAllowed(t *testing.T) {
	drawFound := false
	for seed := int64(0); seed < 500; seed++ {
		r := sim.SimulateMatch(equal1, equal2, false, "", seed)
		if r.Period != "90" {
			t.Errorf("seed %d: group stage produced period %q, want \"90\"", seed, r.Period)
		}
		if r.Team1Goals == r.Team2Goals {
			if r.Winner != "" {
				t.Errorf("seed %d: draw should have empty winner, got %q", seed, r.Winner)
			}
			drawFound = true
		}
	}
	if !drawFound {
		t.Error("no draws found in 500 group-stage matches between equal teams — suspicious")
	}
}

func TestGroupStageNeverHasAETOrPens(t *testing.T) {
	for seed := int64(0); seed < 500; seed++ {
		r := sim.SimulateMatch(spain, france, false, "", seed)
		if r.Period != "90" {
			t.Errorf("seed %d: group stage period = %q, want \"90\"", seed, r.Period)
		}
		if r.Team1GoalsAET != nil || r.Team2GoalsAET != nil {
			t.Errorf("seed %d: group stage should not have AET goals", seed)
		}
		if r.Penalties != nil {
			t.Errorf("seed %d: group stage should not have penalties", seed)
		}
	}
}

// ── AET fields ────────────────────────────────────────────────────────────────

func TestAETFieldsOnlySetWhenRelevant(t *testing.T) {
	for seed := int64(0); seed < 500; seed++ {
		r := sim.SimulateMatch(spain, france, true, "", seed)
		switch r.Period {
		case "90":
			if r.Team1GoalsAET != nil || r.Team2GoalsAET != nil {
				t.Errorf("seed %d: period=90 should have no AET goals", seed)
			}
			if r.Penalties != nil {
				t.Errorf("seed %d: period=90 should have no penalties", seed)
			}
		case "aet", "pens":
			if r.Team1GoalsAET == nil || r.Team2GoalsAET == nil {
				t.Errorf("seed %d: period=%s should have AET goals", seed, r.Period)
			}
		}
		if r.Period == "pens" && r.Penalties == nil {
			t.Errorf("seed %d: period=pens must include penalties", seed)
		}
		if r.Period != "pens" && r.Penalties != nil {
			t.Errorf("seed %d: period=%s should not include penalties", seed, r.Period)
		}
	}
}

// ── Penalties structure ───────────────────────────────────────────────────────

func TestPenaltiesAlwaysHaveAWinner(t *testing.T) {
	found := 0
	for seed := int64(0); seed < 2000 && found < 20; seed++ {
		r := sim.SimulateMatch(equal1, equal2, true, "", seed)
		if r.Period != "pens" {
			continue
		}
		found++
		p := r.Penalties
		if p.Team1Scored == p.Team2Scored {
			t.Errorf("seed %d: penalty shootout ended in a draw (%d-%d)", seed, p.Team1Scored, p.Team2Scored)
		}
		if r.Winner == "" {
			t.Errorf("seed %d: pens result has no winner", seed)
		}
		if p.Team1Scored > p.Team2Scored && r.Winner != equal1.ID {
			t.Errorf("seed %d: team1 scored more penalties but winner=%q", seed, r.Winner)
		}
		if p.Team2Scored > p.Team1Scored && r.Winner != equal2.ID {
			t.Errorf("seed %d: team2 scored more penalties but winner=%q", seed, r.Winner)
		}
	}
	if found == 0 {
		t.Error("no penalty shootouts found in 2000 knockout matches between equal teams")
	}
}

func TestPenaltyKicksAlternateTeams(t *testing.T) {
	found := false
	for seed := int64(0); seed < 2000; seed++ {
		r := sim.SimulateMatch(equal1, equal2, true, "", seed)
		if r.Period != "pens" {
			continue
		}
		found = true
		kicks := r.Penalties.Kicks
		if len(kicks)%2 != 0 {
			t.Errorf("seed %d: odd number of kicks (%d)", seed, len(kicks))
		}
		for i, k := range kicks {
			wantTeam := equal1.ID
			if i%2 == 1 {
				wantTeam = equal2.ID
			}
			if k.Team != wantTeam {
				t.Errorf("seed %d: kick %d: got team %q, want %q", seed, i, k.Team, wantTeam)
			}
		}
		break
	}
	if !found {
		t.Error("no penalty shootouts found to test kick ordering")
	}
}

func TestPenaltyKickNumbersAreSequential(t *testing.T) {
	found := false
	for seed := int64(0); seed < 2000; seed++ {
		r := sim.SimulateMatch(equal1, equal2, true, "", seed)
		if r.Period != "pens" {
			continue
		}
		found = true
		kicks := r.Penalties.Kicks
		// Kicks come in pairs; each pair shares the same kick_no.
		for i := 0; i < len(kicks)-1; i += 2 {
			wantNo := i/2 + 1
			if kicks[i].KickNo != wantNo || kicks[i+1].KickNo != wantNo {
				t.Errorf("seed %d: kick pair %d has kick_nos %d/%d, want %d",
					seed, i/2, kicks[i].KickNo, kicks[i+1].KickNo, wantNo)
			}
		}
		break
	}
	if !found {
		t.Error("no penalty shootouts found to test kick numbering")
	}
}

// ── Host advantage ────────────────────────────────────────────────────────────

func TestHostAdvantageIncreasesWinRate(t *testing.T) {
	const n = 2000
	winsWithout, winsWith := 0, 0

	for seed := int64(0); seed < n; seed++ {
		r := sim.SimulateMatch(usa, spain, false, "", seed)
		if r.Winner == usa.ID {
			winsWithout++
		}

		rH := sim.SimulateMatch(usa, spain, false, usa.ID, seed)
		if rH.Winner == usa.ID {
			winsWith++
		}
	}

	if winsWith <= winsWithout {
		t.Errorf("host advantage did not increase win rate: without=%d with=%d (n=%d)",
			winsWithout, winsWith, n)
	}
}

func TestHostAdvantageAppliesOnlyToHostTeam(t *testing.T) {
	// Passing a non-participant as host_team_id should have no effect.
	r1 := sim.SimulateMatch(spain, france, false, "", 42)
	r2 := sim.SimulateMatch(spain, france, false, "XYZ", 42)
	if r1 != r2 {
		t.Errorf("unknown host_team_id changed the result: %+v vs %+v", r1, r2)
	}
}

// ── Statistical sanity ────────────────────────────────────────────────────────

func TestStrongerTeamWinsMoreOften(t *testing.T) {
	const n = 2000
	spainWins := 0
	for seed := int64(0); seed < n; seed++ {
		r := sim.SimulateMatch(spain, newZeal, false, "", seed)
		if r.Winner == spain.ID {
			spainWins++
		}
	}
	winRate := float64(spainWins) / n
	// Spain (Elo 2165) vs New Zealand (1585): ~580 pt gap; expect ~80%+ win rate
	if winRate < 0.75 {
		t.Errorf("Spain win rate vs New Zealand = %.1f%%, expected > 75%%", winRate*100)
	}
}

func TestDrawRateIsReasonable(t *testing.T) {
	const n = 2000
	draws := 0
	for seed := int64(0); seed < n; seed++ {
		r := sim.SimulateMatch(equal1, equal2, false, "", seed)
		if r.Winner == "" {
			draws++
		}
	}
	drawRate := float64(draws) / n
	// WC draw rate is ~25%; allow a wide band for equal-strength teams
	if drawRate < 0.15 || drawRate > 0.45 {
		t.Errorf("draw rate between equal teams = %.1f%%, expected 15-45%%", drawRate*100)
	}
}

func TestScoresAreReasonable(t *testing.T) {
	const n = 2000
	totalGoals := 0
	for seed := int64(0); seed < n; seed++ {
		r := sim.SimulateMatch(spain, france, false, "", seed)
		totalGoals += r.Team1Goals + r.Team2Goals
	}
	avg := float64(totalGoals) / n
	// WC average is ~2.5 goals/game; allow 2.0–3.5 for a strong vs strong match
	if avg < 2.0 || avg > 3.5 {
		t.Errorf("average goals/game = %.2f, expected 2.0–3.5", avg)
	}
}
