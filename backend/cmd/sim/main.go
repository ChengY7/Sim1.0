package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chengyang/sim1.0/backend/internal/config"
	"github.com/chengyang/sim1.0/backend/internal/sim"
)

const maxPossSafety = 300

func main() {
	home := flag.String("home", "LAL", "home team id (see config/teams.json)")
	away := flag.String("away", "BOS", "away team id")
	seed := flag.Int64("seed", 0, "RNG seed (0 = time-based)")
	quiet := flag.Bool("quiet", false, "only print final line")
	flag.Parse()

	cfgDir := "config"
	if _, err := os.Stat(cfgDir); os.IsNotExist(err) {
		cfgDir = filepath.Join("backend", "config")
	}

	bundle, err := config.Load(cfgDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
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

	state := engine.NewGame()
	if !*quiet {
		fmt.Printf("Sim1.0 — %s vs %s (seed %d)\n", state.HomeName, state.AwayName, s)
		fmt.Printf("Pace %g → ~%d total possessions, ~%.1fs per possession\n\n",
			g.Pace, expected, g.SecondsPerPossession())
	}

	for state.Status != "final" {
		if state.Possession >= maxPossSafety {
			fmt.Fprintf(os.Stderr, "safety stop at %d possessions\n", maxPossSafety)
			break
		}
		ev := engine.Step(state)
		if !*quiet {
			fmt.Printf("P%03d  %s\n", ev.Possession, ev.Text)
			if ev.Points > 0 {
				fmt.Printf("       Score: %s %d — %d %s (Q%d %s)\n",
					state.HomeName, state.HomeScore, state.AwayScore, state.AwayName,
					state.Quarter, formatClock(state.ClockSec))
			}
		}
	}

	label := "Final"
	if state.Status != "final" {
		label = "Stopped early"
	}
	fmt.Printf("\n%s: %s %d — %d %s | Q%d %s | %d possessions (~%d expected) | %s\n",
		label,
		state.HomeName, state.HomeScore, state.AwayScore, state.AwayName,
		state.Quarter, formatClock(state.ClockSec),
		state.Possession, expected, state.Status)
}

func formatClock(sec int) string {
	if sec < 0 {
		sec = 0
	}
	return fmt.Sprintf("%d:%02d", sec/60, sec%60)
}
