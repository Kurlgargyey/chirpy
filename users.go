package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Kurlgargyey/chirpy/internal/auth"
	"github.com/Kurlgargyey/chirpy/internal/database"
	"github.com/google/uuid"
)

type userRequestBody struct {
	Password  string `json:"password"`
	Email     string `json:"email"`
	ExpiresIn int    `json:"expires_in_seconds"`
}
type loginRequestBody struct {
	userRequestBody
}

type userResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}
type loginResponse struct {
	userResponse
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func (cfg *apiConfig) createUserHandler() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			defer r.Body.Close()
			var requestBody userRequestBody
			if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
				writeError(w, fmt.Sprintf("error decoding json: %s", err), 400)
				return
			}
			hashed_pwd, err := auth.HashPassword(requestBody.Password)
			if err != nil {
				writeError(w, fmt.Sprintf("error hashing password: %s", err), 400)
				return
			}
			userParams := database.CreateUserParams{
				Email:          requestBody.Email,
				HashedPassword: hashed_pwd,
			}
			user, err := cfg.db.CreateUser(r.Context(), userParams)
			if err != nil {
				writeError(w, fmt.Sprintf("error creating user: %s", err), 400)
				return
			}
			response := userResponse{
				ID:        user.ID,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
				Email:     user.Email,
			}
			dat, _ := json.Marshal(response)
			w.WriteHeader(201)
			w.Write(dat)
		})
}

func (cfg *apiConfig) loginHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var requestBody loginRequestBody
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			writeError(w, fmt.Sprintf("error decoding json: %s", err), 400)
			return
		}
		user, fetch_err := cfg.db.GetUser(r.Context(), requestBody.Email)
		hash_err := auth.CheckPasswordHash(requestBody.Password, user.HashedPassword)

		if fetch_err != nil || hash_err != nil {
			w.WriteHeader(401)
			w.Write([]byte("incorrect email or password"))
			return
		}

		expiresIn := time.Hour

		token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, expiresIn)
		if err != nil {
			writeError(w, fmt.Sprintf("error obtaining JWT: %s", err), 400)
			return
		}
		refresh_token, err := auth.MakeRefreshToken()
		if err != nil {
			writeError(w, fmt.Sprintf("error obtaining refresh token: %s", err), 400)
			return
		}
		refreshTokenParams := database.CreateRefreshTokenParams{
			Token:     refresh_token,
			UserID:    user.ID,
			ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
		}
		_, refresh_err := cfg.db.CreateRefreshToken(r.Context(), refreshTokenParams)
		if refresh_err != nil {
			writeError(w, fmt.Sprintf("error committing refresh token to database: %s", refresh_err), 400)
			return
		}

		response := loginResponse{
			userResponse: userResponse{ID: user.ID,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
				Email:     user.Email},
			Token:        token,
			RefreshToken: refresh_token,
		}
		dat, _ := json.Marshal(response)
		w.Write(dat)
	})
}
