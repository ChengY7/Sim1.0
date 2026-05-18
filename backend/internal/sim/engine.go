package sim

import (
	"fmt"
	"math/rand"

	"github.com/chengyang/sim1.0/backend/internal/config"
)

type Side string

const (
	Home Side = "home"
	Away Side = "away"
)

type Event struct {
	Possession int    `json:"possession"`
	Team       Side   `json:"team"`
	Type       string `json:"type"`
	Points     int    `json:"points"`
	Text       string `json:"text"`
}

type State struct {
	HomeID     string `json:"home_id"`
	AwayID     string `json:"away_id"`
	HomeName   string `json:"home_name"`
	AwayName   string `json:"away_name"`
	HomeScore  int    `json:"home_score"`
	AwayScore  int    `json:"away_score"`
	Offense    Side   `json:"offense"`
	Possession int    `json:"possession"`
	Quarter    int    `json:"quarter"`
	ClockSec   int    `json:"clock_sec"`
	Status     string `json:"status"` // in_progress | final
}

type Engine struct {
	cfg    *config.Bundle
	home   config.Team
	away   config.Team
	rng    *rand.Rand
}

func NewEngine(cfg *config.Bundle, homeID, awayID string, seed int64) (*Engine, error) {
	home, err := cfg.Team(homeID)
	if err != nil {
		return nil, err
	}
	away, err := cfg.Team(awayID)
	if err != nil {
		return nil, err
	}
	return &Engine{
		cfg:  cfg,
		home: home,
		away: away,
		rng:  rand.New(rand.NewSource(seed)),
	}, nil
}

func (e *Engine) NewGame() *State {
	return &State{
		HomeID:   e.home.ID,
		AwayID:   e.away.ID,
		HomeName: e.home.Name,
		AwayName: e.away.Name,
		Offense:  Away,
		Quarter:  1,
		ClockSec: e.cfg.Game.QuarterSeconds,
		Status:   "in_progress",
	}
}

func (e *Engine) Step(s *State) Event {
	if s.Status == "final" {
		return Event{Type: "game_over", Text: "Game over"}
	}

	s.Possession++
	offTeam, defTeam := e.teamsFor(s.Offense)
	mult := offTeam.Offense / defTeam.Defense

	outcome := e.pickOutcome(mult)
	ev := Event{
		Possession: s.Possession,
		Team:       s.Offense,
		Type:       outcome.Type,
		Points:     outcome.Points,
	}

	switch outcome.Type {
	case "make_2pt", "make_3pt", "make_ft":
		e.addScore(s, outcome.Points)
		ev.Text = fmt.Sprintf("%s %s (%d pts)", teamLabel(s, s.Offense), outcome.Type, outcome.Points)
	case "miss_2pt", "miss_3pt", "miss_ft":
		ev.Text = fmt.Sprintf("%s %s", teamLabel(s, s.Offense), outcome.Type)
	case "turnover":
		ev.Text = fmt.Sprintf("%s turnover", teamLabel(s, s.Offense))
	case "foul":
		ev.Text = fmt.Sprintf("%s foul", teamLabel(s, s.Offense))
	default:
		ev.Text = fmt.Sprintf("%s %s", teamLabel(s, s.Offense), outcome.Type)
	}

	e.tickClock(s)
	e.flipPossession(s)
	return ev
}

func (e *Engine) pickOutcome(offenseMult float64) config.Outcome {
	weights := make([]float64, len(e.cfg.Outcomes))
	var total float64
	for i, o := range e.cfg.Outcomes {
		w := o.Weight
		if o.Points > 0 {
			w *= offenseMult
		} else if o.Type == "turnover" {
			w /= offenseMult
		}
		weights[i] = w
		total += w
	}

	r := e.rng.Float64() * total
	var cum float64
	for i, o := range e.cfg.Outcomes {
		cum += weights[i]
		if r <= cum {
			return o
		}
	}
	return e.cfg.Outcomes[len(e.cfg.Outcomes)-1]
}

func (e *Engine) teamsFor(offense Side) (off, def config.Team) {
	if offense == Home {
		return e.home, e.away
	}
	return e.away, e.home
}

func (e *Engine) addScore(s *State, pts int) {
	if s.Offense == Home {
		s.HomeScore += pts
	} else {
		s.AwayScore += pts
	}
}

func (e *Engine) flipPossession(s *State) {
	if s.Offense == Home {
		s.Offense = Away
	} else {
		s.Offense = Home
	}
}

func (e *Engine) tickClock(s *State) {
	g := e.cfg.Game
	base := int(g.SecondsPerPossession())
	j := g.TickJitterSec
	elapsed := base + e.rng.Intn(2*j+1) - j
	if elapsed < 1 {
		elapsed = 1
	}

	s.ClockSec -= elapsed
	for s.ClockSec <= 0 && s.Status != "final" {
		s.Quarter++
		if s.Quarter > g.Quarters {
			s.Quarter = g.Quarters
			s.ClockSec = 0
			s.Status = "final"
			return
		}
		s.ClockSec += g.QuarterSeconds
	}
}

func teamLabel(s *State, side Side) string {
	if side == Home {
		return s.HomeName
	}
	return s.AwayName
}
