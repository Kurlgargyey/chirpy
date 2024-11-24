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

func (cfg *apiConfig) getChirpsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		defer r.Body.Close()
		chirps, err := cfg.db.GetAllChirps(r.Context())
		if err != nil {
			writeError(w, fmt.Sprintf("error retrieving chirps: %s", err), 400)
		}
		var responseArray []chirpResponse
		for _, chirp := range chirps {
			responseArray = append(responseArray, chirpResponse{
				ID:        chirp.ID,
				CreatedAt: chirp.CreatedAt,
				UpdatedAt: chirp.UpdatedAt,
				Body:      chirp.Body,
				UserID:    chirp.UserID,
			})
		}
		dat, _ := json.Marshal(responseArray)
		w.Write(dat)
	})
}

func (cfg *apiConfig) getChirpHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		defer r.Body.Close()

		chirp, err := cfg.db.GetChirp(r.Context(), uuid.MustParse(r.PathValue("chirpID")))
		if err != nil {
			writeError(w, fmt.Sprintf("error retrieving chirp: %s", err), 404)
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
		w.WriteHeader(200)
		w.Write(dat)
	})
}
