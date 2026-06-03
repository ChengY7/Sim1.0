# Season Rating Files

Each file `nba_YYYY-YY.json` holds per-team offensive and defensive multipliers for one NBA season.

## Format

```json
{
  "season": "2024-25",
  "teams": [
    { "id": "BOS", "offense": 1.051, "defense": 1.033 },
    ...
  ]
}
```

All 30 teams must be present. The file name must match the pattern `nba_YYYY-YY.json`; the loader sorts files descending and uses the latest as the default season.

## Calculating Ratings from NBA Stats

Source: NBA.com Advanced Team Stats — filter by season, stat type = "Advanced".
Required columns: **OffRtg**, **DefRtg** (points scored/allowed per 100 possessions).

### Formula

```
league_avg  = sum(OffRtg for all 30 teams) / 30

offense     = OffRtg  / league_avg
defense     = league_avg / DefRtg
```

Both multipliers are centered around 1.0 (league average = 1.0).

- `offense > 1.0` → above-average scoring offense
- `defense > 1.0` → above-average defense (harder to score against)
- `defense` inverts DefRtg because lower DefRtg = better defense in NBA stats,
  but the sim uses higher = stronger for both dimensions.

The practical output range is roughly **0.93–1.07** across a typical NBA season.

### How the multipliers are used in the sim

Per possession: `mult = offense_team.offense / defense_team.defense`

Outcomes tagged `offense_scale: "up"` (makes) are multiplied by `mult`;
outcomes tagged `"down"` (misses, turnovers) are divided by `mult`.

### Example (2024-25, league_avg = 113.68)

| Team | OffRtg | DefRtg | offense | defense |
|------|--------|--------|---------|---------|
| OKC  | 119.2  | 106.6  | 1.049   | 1.066   |
| CLE  | 121.0  | 111.8  | 1.064   | 1.017   |
| UTA  | 110.2  | 119.4  | 0.969   | 0.952   |
| WAS  | 105.8  | 118.0  | 0.931   | 0.963   |
