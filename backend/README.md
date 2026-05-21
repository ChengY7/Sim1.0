# Backend

Possession-by-possession NBA simulator using seeded randomness and JSON config.

## Run

**HTTP API** (default `:8080`):
```bash
go run ./cmd/server
```

**CLI** (single game, prints play-by-play):
```bash
go run ./cmd/sim -home LAL -away BOS -seed 42
go run ./cmd/sim -home LAL -away BOS -seed 42 -quiet   # final line only
```

## API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/seasons` | Available season rating files and default season |
| GET | `/teams` | All 30 team IDs, names, conferences, and divisions |
| POST | `/simulate` | Simulate a single game; returns state, events, seed |
| POST | `/simulate/season` | Simulate the full regular season; returns standings + NBA Cup bracket |
| POST | `/simulate/playin` | Simulate the play-in tournament (6 games); returns bracket + seeds 7/8 |
| POST | `/simulate/draft-lottery` | Run the draft lottery for 14 teams; returns picks 1–14 |
| GET | `/swagger/index.html` | Interactive API docs |

All `POST /simulate*` endpoints accept an optional `"season"` field (e.g. `"2024-25"`). Omit it to use the default (latest) season.

```bash
# Example
curl -s http://localhost:8080/seasons
curl -s -X POST http://localhost:8080/simulate \
  -H 'Content-Type: application/json' \
  -d '{"home_team_id":"LAL","away_team_id":"BOS","seed":42,"season":"2024-25"}'
```

## Config files

| File | Purpose |
|------|---------|
| `internal/config/data/teams.json` | 30 teams — ID, name, conference, division |
| `internal/config/data/seasons/nba_YYYY-YY.json` | Per-team offense/defense ratings for each season |
| `internal/config/data/outcomes.json` | Possession outcome types and base weights |
| `internal/config/data/game.json` | Quarters, clock, pace (possessions per team per 48 min) |
| `internal/config/data/schedule.json` | Full regular-season schedule |
| `internal/config/data/cup_groups.json` | NBA Cup group assignments |
| `internal/config/data/draft_lottery.json` | Ball-combination counts for 14 lottery seeds |

### Adding a new season

Create `internal/config/data/seasons/nba_YYYY-YY.json` with offense/defense multipliers for all 30 teams. The loader picks up any file matching that pattern automatically and uses the lexicographically latest as the default. See [the seasons README](internal/config/data/seasons/README.md) for the rating formula.

## Pace and possessions

`pace` in `game.json` = offensive possessions per team per 48 minutes (NBA average ~98–102).

```
total possessions ≈ 2 × pace
seconds per possession ≈ (quarters × quarter_seconds) / (2 × pace)
```

Example: `pace: 100` → ~200 total possessions, ~14.4 s/possession. Actual count varies with `tick_jitter_sec`.

## CLI flags

| Flag | Description |
|------|-------------|
| `-home` | Home team ID (e.g. `LAL`) |
| `-away` | Away team ID (e.g. `BOS`) |
| `-seed` | Integer seed — same seed produces the same game |
| `-quiet` | Print only the final score line |

## Swagger

```bash
go install github.com/swaggo/swag/cmd/swag@latest
make swagger   # regenerates backend/docs/ from handler comments
```
