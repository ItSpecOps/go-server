package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/ItSpecOps/go-server/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	platform := cfg.platform
	if platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(`{"error":"forbidden"}`))
		return
	}
	// Delete all users
	err := cfg.db.DeleteAllUsers(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(`{"error":"could not reset users"}`))
		return
	}
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("OK"))
}

func main() {
	// setup DB connection
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	sqlDB, err := sql.Open("postgres", dbURL)
	if err != nil {
		panic(err)
	}
	queries := database.New(sqlDB)
	platform := os.Getenv("PLATFORM")

	mux := http.NewServeMux()
	apiCfg := &apiConfig{
		db:       queries,
		platform: platform,
	}

	// Healthz endpoint
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Metrics endpoint
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)

	// Reset endpoint
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	// Create chirp endpoint
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerChirpsCreate)

	// Get chirps endpoint
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerChirpsGet)

	// Get chirp endpoint
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerChirpGet)

	// Create user endpoint
	mux.HandleFunc("POST /api/users", apiCfg.handlerUsersCreate)

	// Fileserver at /app/ with metrics middleware
	fileserverHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileserverHandler))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	fmt.Println("Server starting on http://localhost:8080")
	server.ListenAndServe()
}