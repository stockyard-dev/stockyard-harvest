package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/stockyard-dev/stockyard-harvest/internal/server"
	"github.com/stockyard-dev/stockyard-harvest/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9810"
	}
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./harvest-data"
	}

	db, err := store.Open(dataDir)
	if err != nil {
		log.Fatalf("harvest: %v", err)
	}
	defer db.Close()

	srv := server.New(db, server.DefaultLimits())

	fmt.Printf("\n  Harvest — Self-hosted harvest and crop tracking for farms\n  Dashboard:  http://localhost:%s/ui\n  API:        http://localhost:%s/api\n  Questions? hello@stockyard.dev — I read every message\n\n", port, port)
	log.Printf("harvest: listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, srv))
}
