package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ItSpecOps/go-server/internal/database"
	"github.com/google/uuid"
)


func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
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
		Body:   cleaned,
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
		Body:      dbChirp.Body,
		UserID:    uuid.MustParse(dbChirp.UserID),
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}