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
curl -s http://localhost:8080/teams
curl -s -X POST http://localhost:8080/simulate \
  -H 'Content-Type: application/json' \
  -d '{"home_team_id":"LAL","away_team_id":"BOS","seed":42}'
```

| Method | Path | Description |
|--------|------|-------------|
| GET | `/teams` | All team ids from `teams.json` |
| POST | `/simulate` | Run full game until clock ends; returns `state`, `events`, `seed` |

## Config

| File | Purpose |
|------|---------|
| `config/teams.json` | All 30 teams + simple offense/defense multipliers |
| `config/outcomes.json` | Possession outcome types and base weights |
| `config/game.json` | Quarters, clock, **pace** (possessions per team per 48 min) |

Edit weights to tune how often you see 2PT, 3PT, FT, turnovers, etc.

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
