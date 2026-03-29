package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/youruser/geo-api/internal/handler"
	"github.com/youruser/geo-api/internal/repository"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default environment variables")
	}

	dataPath := env("GEO_DATA_PATH", "data/countries+states+cities.json")
	metadataPath := env("METADATA_PATH", "data/cities_metadata.json")
	port := env("PORT", "8082")

	log.Printf("Loading geo data from %s …", dataPath)
	start := time.Now()

	repo, err := repository.New(dataPath, metadataPath)
	if err != nil {
		log.Fatalf("Failed to load geo data: %v\n\n"+
			"Download the dataset from:\n"+
			"  https://github.com/dr5hn/countries-states-cities-database\n"+
			"and place countries+states+cities.json inside the data/ folder.", err)
	}

	stats := repo.GetStats()
	log.Printf("Loaded %d countries, %d states, %d cities in %s",
		stats.Countries, stats.States, stats.Cities, time.Since(start))

	mux := http.NewServeMux()
	h := handler.New(repo)
	h.RegisterRoutes(mux)

	// Admin routes and Static Web Dashboard
	h.RegisterAdminRoutes(mux)

	// Handle static dashboard at root
	fs := http.FileServer(http.Dir("./public"))
	mux.Handle("/", fs)

	// CORS + logging middleware
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      middleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Server listening on http://localhost:%s", port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// CORS headers (adjust origins for production)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
