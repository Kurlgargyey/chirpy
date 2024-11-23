package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Kurlgargyey/chirpy/internal/database"
	"github.com/google/uuid"
)

type chirpRequestBody struct {
	Body   string `json:"body"`
	UserID string `json:"user_id"`
}

type chirpResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) createChirpHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		defer r.Body.Close()
		var requestBody chirpRequestBody
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			writeError(w, fmt.Sprintf("error decoding json: %s", err), 400)
			return
		}
		chirpParams := database.CreateChirpParams{
			Body:   requestBody.Body,
			UserID: uuid.MustParse(requestBody.UserID),
		}
		chirp, err := cfg.db.CreateChirp(r.Context(), chirpParams)
		if err != nil {
			writeError(w, fmt.Sprintf("error creating chirp: %s", err), 400)
			return
		}
		response := chirpResponse{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
		dat, _ := json.Marshal(response)
		w.WriteHeader(201)
		w.Write(dat)
	})
}
