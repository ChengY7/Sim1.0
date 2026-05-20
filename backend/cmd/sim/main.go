package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/chengyang/sim1.0/backend/internal/config"
	"github.com/chengyang/sim1.0/backend/internal/sim"
)

func main() {
	home := flag.String("home", "LAL", "home team id (use -teams to list all)")
	away := flag.String("away", "BOS", "away team id (use -teams to list all)")
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
		for _, t := range bundle.SortedTeams() {
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

	result := engine.RunUntilFinal()
	state := result.State

	if !*quiet {
		fmt.Printf("Sim1.0 — %s vs %s (seed %d)\n", state.HomeName, state.AwayName, s)
		fmt.Printf("Pace %g → ~%d total possessions, ~%.1fs per possession\n\n",
			g.Pace, expected, g.SecondsPerPossession())
		for _, ev := range result.Events {
			clock := fmt.Sprintf("%s %s", ev.Period, formatClock(ev.ClockSec))
			fmt.Printf("P%03d  %-10s  %-40s %d-%d\n", ev.Possession, clock, ev.Text, ev.HomeScore, ev.AwayScore)
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

func formatClock(sec int) string {
	if sec < 0 {
		sec = 0
	}
	return fmt.Sprintf("%d:%02d", sec/60, sec%60)
}
