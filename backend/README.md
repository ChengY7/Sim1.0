# Backend (v0)

Possession-by-possession simulator using **seeded randomness** and **JSON config** — no external NBA API, no HTTP yet.

## Run

**CLI** (from `backend/`):

```bash
go run ./cmd/sim -home LAL -away BOS -seed 42
go run ./cmd/sim -home LAL -away BOS -seed 42 -quiet   # final line only
```

**HTTP API**:

```bash
go run ./cmd/server
# Swagger UI: http://localhost:8080/swagger/index.html
curl -s http://localhost:8080/teams
curl -s -X POST http://localhost:8080/simulate \
  -H 'Content-Type: application/json' \
  -d '{"home_team_id":"LAL","away_team_id":"BOS","seed":42}'
```

| Method | Path | Description |
|--------|------|-------------|
| GET | `/teams` | All team ids from `teams.json` |
| POST | `/simulate` | Run full game until clock ends; returns `state`, `events`, `seed` |
| GET | `/swagger/index.html` | Interactive API docs (Swagger UI) |

### Swagger setup

```bash
go install github.com/swaggo/swag/cmd/swag@latest
make swagger   # generates backend/docs/ from handler comments
```

## Config

| File | Purpose |
|------|---------|
| `config/teams.json` | All 30 teams + simple offense/defense multipliers |
| `config/outcomes.json` | Possession outcome types and base weights |
| `config/game.json` | Quarters, clock, **pace** (possessions per team per 48 min) |

Edit weights to tune how often you see 2PT, 3PT, FT, turnovers, etc.

## Team ratings

`offense` and `defense` in `teams.json` are multipliers derived from real NBA Offensive/Defensive Ratings (points per 100 possessions), scaled to amplify differences:

```
offense = 2 × (OffRtg / 115) − 1
defense = 2 × (115 / DefRtg) − 1
```

115 is the approximate league-average rating. The `2x − 1` transform doubles the spread around 1.0 so matchups feel meaningful. A higher `defense` value means better defense (it sits in the denominator: `mult = offense / defense`).

Per possession, `mult = offense_team.offense / defense_team.defense`. Outcomes tagged `offense_scale: "up"` (makes) are multiplied by `mult`; outcomes tagged `"down"` (misses, turnovers) are divided. The practical range is roughly 0.76 (weak offense vs elite defense) to 1.25 (elite offense vs weak defense).

## Possessions per game

NBA **pace** ≈ possessions per team per 48 minutes (typical ~98–102).

```
total possessions ≈ 2 × pace
seconds per possession ≈ (4 × 12 × 60) / (2 × pace) = 2880 / (2 × pace)
```

Example: `pace: 100` → ~**200** total possessions, ~**14.4s** per possession on the clock.

Actual count varies slightly due to `tick_jitter_sec` in `game.json`.

## Flags

- `-home` / `-away` — team ids from `teams.json` (e.g. `LAL`, `BOS`)
- `-seed` — same seed ⇒ same game (useful for debugging)
- `-quiet` — print only the final summary line
