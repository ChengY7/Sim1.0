# Sim1.0

Possession-by-possession NBA game simulator (Go backend + React frontend planned).

## Backend v0 (current)

CLI simulator with JSON config and seeded randomness — no external API yet.

```bash
cd backend
go run ./cmd/sim -home LAL -away BOS -seed 42          # CLI
go run ./cmd/server                                      # API on :8080
```

Runs until the game clock ends (~200 possessions at default pace). See [backend/README.md](backend/README.md) for API details.

See [backend/README.md](backend/README.md) and [docs/BACKEND_V0.md](docs/BACKEND_V0.md).
