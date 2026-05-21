package api

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(h *Handlers, corsOrigin string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /teams", h.ListTeams)
	mux.HandleFunc("POST /simulate", h.Simulate)
	mux.HandleFunc("POST /simulate/season", h.SimulateSeason)
	mux.HandleFunc("POST /simulate/draft-lottery", h.SimulateDraftLottery)
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
