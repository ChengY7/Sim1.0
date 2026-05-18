# Backend v0 — design notes

## No external API (for now)

Stats come from JSON on disk, not `nba_api` or NBA.com. Randomness is **seeded** (`math/rand`) so runs are reproducible.

## JSON vs hardcoding

| Data | Store in JSON? | Why |
|------|----------------|-----|
| Team names/ids | Yes (`teams.json`) | Easy to list all 30; tweak ratings without recompiling |
| Outcome types (2pt, 3pt, TO, …) | Yes (`outcomes.json`) | Tune league-wide frequencies in one place |
| Pace / game length | Yes (`game.json`) | Drives how many possessions fit in 48 min |
| Game state / play-by-play | No (in memory) | Ephemeral per run until you add save/load |

Later you can replace `teams.json` with ingest from `nba_api` without changing the simulator interface.

## HTTP API — add later

For React, a small REST surface is enough:

| When | Endpoint | Purpose |
|------|----------|---------|
| Later | `POST /games` | `{ home_id, away_id, seed? }` → create game |
| Later | `POST /games/{id}/step` | Advance one possession |
| Later | `GET /games/{id}` | Score, clock, status |
| Later | `GET /games/{id}/events` | Play-by-play |

**v0 skips HTTP** and uses the CLI (`cmd/sim`) so you can focus on the possession model first.

## Possessions per game

Do **not** pick a possession count by hand. Derive it from **pace** and **game length**:

1. **Game seconds** = `quarters × quarter_seconds` (default 4 × 720 = 2880).
2. **Pace** (in `game.json`) = offensive possessions per team per 48 minutes (~100 in the NBA).
3. **Total possessions** (both teams) ≈ `2 × pace` (e.g. ~200 at pace 100).
4. **Seconds per possession** ≈ `game_seconds / (2 × pace)` (e.g. ~14.4s).

The sim runs until the clock hits zero in Q4; possession count emerges from pace + jitter.

## Other things to consider

1. **Seeded RNG** — always support `-seed` for debugging.
2. **Possession log** — slice of events; enables replay and tests.
3. **Keep sim pure** — `internal/sim` has no HTTP imports; wire API in `internal/api` later.
4. **Fouls / FTs** — v0 treats foul as a single event; later chain: foul → FT possessions.
5. **Offensive rebounds** — ignore in v0; add as optional outcome later.
6. **Tests** — golden-file test: fixed seed + expected score band or event sequence.
