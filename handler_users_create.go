package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)


func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
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