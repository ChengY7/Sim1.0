# Frontend

React + Vite UI for the NBA simulator. Proxies all API calls to the Go backend on `:8080`.

## Run

```bash
npm install
npm run dev    # dev server on :5173
npm run build  # production build → dist/
```

The backend must be running on `:8080` for any simulation to work.

## Features

| Tab | Description |
|-----|-------------|
| Simulate Game | Pick two teams, run a possession-by-possession game, view play-by-play |
| Season | Simulate the full regular season; view East/West standings, NBA Cup bracket |
| Play-In | Simulate the play-in tournament once the season is done |
| Draft Lottery | View lottery odds table; simulate the draft once play-in is done |

The season selector (coming soon) switches all simulations between available seasons (e.g. `2024-25`, `2025-26`).
