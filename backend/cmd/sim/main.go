package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/chengyang/sim1.0/backend/internal/config"
	"github.com/chengyang/sim1.0/backend/internal/sim"
)

func main() {
	home := flag.String("home", "LAL", "home team id (use -teams to list all)")
	away := flag.String("away", "BOS", "away team id")
	seed := flag.Int64("seed", 0, "RNG seed (0 = time-based)")
	quiet := flag.Bool("quiet", false, "only print final line")
	teams := flag.Bool("teams", false, "list all team ids and exit")
	flag.Parse()

	bundle, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	if *teams {
		for _, t := range sortedTeams(bundle) {
			fmt.Printf("%-6s %s\n", t.ID, t.Name)
		}
		return
	}

	s := *seed
	if s == 0 {
		s = time.Now().UnixNano()
	}

	engine, err := sim.NewEngine(bundle, *home, *away, s)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	g := bundle.Game
	expected := int(g.ExpectedTotalPossessions())

	if !*quiet {
		fmt.Printf("Sim1.0 — %s vs %s (seed %d)\n", bundle.Teams[*home].Name, bundle.Teams[*away].Name, s)
		fmt.Printf("Pace %g → ~%d total possessions, ~%.1fs per possession\n\n",
			g.Pace, expected, g.SecondsPerPossession())
	}

	result := engine.RunUntilFinal()
	state := result.State

	if !*quiet {
		for _, ev := range result.Events {
			fmt.Printf("P%03d  %s\n", ev.Possession, ev.Text)
		}
	}

	label := "Final"
	if state.Status != "final" {
		label = "Stopped early"
	}
	fmt.Printf("\n%s: %s %d — %d %s | %s %s | %d possessions (~%d expected) | %s\n",
		label,
		state.HomeName, state.HomeScore, state.AwayScore, state.AwayName,
		sim.PeriodLabel(state.Quarter, bundle.Game.Quarters), formatClock(state.ClockSec),
		state.Possession, expected, state.Status)
}

func sortedTeams(b *config.Bundle) []config.Team {
	out := make([]config.Team, 0, len(b.Teams))
	for _, t := range b.Teams {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func formatClock(sec int) string {
	if sec < 0 {
		sec = 0
	}
	return fmt.Sprintf("%d:%02d", sec/60, sec%60)
}
