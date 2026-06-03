package api

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(h *Handlers, corsOrigin string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /nba/seasons", h.ListSeasons)
	mux.HandleFunc("GET /nba/teams", h.ListTeams)
	mux.HandleFunc("POST /nba/simulate", h.Simulate)
	mux.HandleFunc("POST /nba/simulate/season", h.SimulateSeason)
	mux.HandleFunc("POST /nba/simulate/playin", h.SimulatePlayIn)
	mux.HandleFunc("POST /nba/simulate/draft-lottery", h.SimulateDraftLottery)
	mux.HandleFunc("POST /nba/simulate/playoffs", h.SimulatePlayoffs)
	mux.HandleFunc("GET /fifa/teams", h.FIFAListTeams)
	mux.HandleFunc("POST /fifa/simulate", h.FIFASimulate)
	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)
	return cors(mux, corsOrigin)
}

func cors(next http.Handler, origin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
