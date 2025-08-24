package main
import _ "github.com/lib/pq"
// internal imports
import (
	"github.com/ItSpecOps/go-server/internal/database"
	"github.com/ItSpecOps/go-server/internal/auth"
)
// external imports
import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"os"
	"github.com/joho/godotenv"
	"database/sql"
	"github.com/google/uuid"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db *database.Queries
	platform string
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

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserParams
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	dbUser, err := cfg.db.CreateUser(r.Context(), req.Email)
	if err != nil {
		fmt.Printf("CreateUser error: %v\n", err) // Log the actual error
		http.Error(w, `{"error":"could not create user"}`, http.StatusInternalServerError)
		return
	}
	resp := User{
		ID:        uuid.MustParse(dbUser.ID),
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		fmt.Printf("GetChirps error: %v\n", err) // Log the actual error
		http.Error(w, `{"error":"could not get chirps"}`, http.StatusInternalServerError)
		return
	}
	var resp []Chirp
	for _, dbChirp := range dbChirps {
		chirp := Chirp{
			ID:        uuid.MustParse(dbChirp.ID),
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:     dbChirp.Body,
			UserID:  uuid.MustParse(dbChirp.UserID),
		}
		resp = append(resp, chirp)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	chirpIDString := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDString)
	if err != nil {
		http.Error(w, `{"error":"invalid chirp id"}`, http.StatusBadRequest)
		return
	}

	dbChirp, err := cfg.db.GetChirp(r.Context(), chirpID.String())
	if err != nil {
		http.Error(w, `{"error":"chirp not found"}`, http.StatusNotFound)
		return
	}

	resp := Chirp{
		ID:        uuid.MustParse(dbChirp.ID),
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:     dbChirp.Body,
		UserID:  uuid.MustParse(dbChirp.UserID),
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	var req createChirpParams
	type errorResponse struct {
		Error string `json:"error"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserID == "" || req.Body == "" {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	// Validate Chirp length
	if len(req.Body) > 140 {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "Chirp is too long"})
		return
	}

	// Profanity filter
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(req.Body, " ")
	for i, word := range words {
		for _, profane := range profaneWords {
			if strings.EqualFold(word, profane) {
				words[i] = "****"
				break
			}
		}
	}

	cleaned := strings.Join(words, " ")

	dbChirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body: cleaned,
		UserID: req.UserID,
	})
	if err != nil {
		fmt.Printf("CreateUser error: %v\n", err) // Log the actual error
		http.Error(w, `{"error":"could not create chirp"}`, http.StatusInternalServerError)
		return
	}
	resp := Chirp{
		ID:        uuid.MustParse(dbChirp.ID),
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:     dbChirp.Body,
		UserID:  uuid.MustParse(dbChirp.UserID),
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
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
		db: queries,
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
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerCreateChirp)

	// Get chirps endpoint
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)

	// Get chirp endpoint
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetChirp)

	// Create user endpoint
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	

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