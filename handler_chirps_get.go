package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)


func (cfg *apiConfig) handlerChirpsGet(w http.ResponseWriter, r *http.Request) {
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
			Body:      dbChirp.Body,
			UserID:    uuid.MustParse(dbChirp.UserID),
		}
		resp = append(resp, chirp)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (cfg *apiConfig) handlerChirpGet(w http.ResponseWriter, r *http.Request) {
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
		Body:      dbChirp.Body,
		UserID:    uuid.MustParse(dbChirp.UserID),
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}