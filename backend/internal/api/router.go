package api

import "net/http"

func NewRouter(h *Handlers) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /teams", h.ListTeams)
	mux.HandleFunc("POST /simulate", h.Simulate)
	return cors(mux)
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
