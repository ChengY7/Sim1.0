package main

import (
	"log"
	"net/http"
	"os"

	"github.com/chengyang/sim1.0/backend/internal/api"
	"github.com/chengyang/sim1.0/backend/internal/config"
)

func main() {
	cfg, err := config.Load(config.Dir())
	if err != nil {
		log.Fatal(err)
	}

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	h := api.NewHandlers(cfg)
	log.Printf("Sim1.0 API on %s", addr)
	log.Fatal(http.ListenAndServe(addr, api.NewRouter(h)))
}
