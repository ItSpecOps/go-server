package main

import (
	"time"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body     string    `json:"body"`
	UserID  uuid.UUID `json:"user_id"`
}

type createChirpParams struct {
	Body string `json:"body"`
	UserID string `json:"user_id"`
}
