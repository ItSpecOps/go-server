package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ItSpecOps/go-server/internal/database"
	"github.com/ItSpecOps/go-server/internal/auth"
	"github.com/google/uuid"
)


func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	var req createChirpParams

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "missing or malformed JWT", err)
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid or expired JWT", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}


	if userID == uuid.Nil {
		respondWithError(w, http.StatusBadRequest, "invalid user_id", nil)
		return
	}

	// Validate Chirp length
	if len(req.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "chirp body exceeds 140 characters", nil)
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
		Body:   cleaned,
		UserID: userID,
	})
	if err != nil {
		fmt.Printf("CreateUser error: %v\n", err) // Log the actual error
		http.Error(w, `{"error":"could not create chirp"}`, http.StatusInternalServerError)
		return
	}
	resp := Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}