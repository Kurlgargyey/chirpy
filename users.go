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
	Password string `json:"password"`
	Email    string `json:"email"`
}
type loginRequestBody struct {
	userRequestBody
}

type webhookRequestBody struct {
	Event string     `json:"event"`
	Data  userIdJSON `json:"data"`
}

type userIdJSON struct {
	UserID uuid.UUID `json:"user_id"`
}

type userResponse struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}
type loginResponse struct {
	userResponse
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type accessTokenResponse struct {
	Token string `json:"token"`
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
				ID:          user.ID,
				CreatedAt:   user.CreatedAt,
				UpdatedAt:   user.UpdatedAt,
				Email:       user.Email,
				IsChirpyRed: user.IsChirpyRed,
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
		user, fetchErr := cfg.db.GetUser(r.Context(), requestBody.Email)
		hashErr := auth.CheckPasswordHash(requestBody.Password, user.HashedPassword)

		if fetchErr != nil || hashErr != nil {
			w.WriteHeader(401)
			w.Write([]byte("incorrect email or password"))
			return
		}

		token, err := auth.MakeJWT(user.ID, cfg.jwtSecret)
		if err != nil {
			writeError(w, fmt.Sprintf("error obtaining JWT: %s", err), 400)
			return
		}
		refreshToken, err := auth.MakeRefreshToken()
		if err != nil {
			writeError(w, fmt.Sprintf("error obtaining refresh token: %s", err), 400)
			return
		}
		refreshTokenParams := database.CreateRefreshTokenParams{
			Token:     refreshToken,
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
				CreatedAt:   user.CreatedAt,
				UpdatedAt:   user.UpdatedAt,
				Email:       user.Email,
				IsChirpyRed: user.IsChirpyRed},
			Token:        token,
			RefreshToken: refreshToken,
		}
		dat, _ := json.Marshal(response)
		w.Write(dat)
	})
}

func (cfg *apiConfig) refreshTokenHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		bearerToken, bearerErr := auth.GetBearerToken(r.Header)
		if bearerErr != nil {
			writeError(w, "could not obtain a bearer token", 400)
			return
		}
		token, err := cfg.db.GetRefreshToken(r.Context(), bearerToken)
		if err != nil || token.RevokedAt.Valid {
			writeError(w, "could not obtain a valid refresh token", 401)
			return
		}
		accessToken, err := auth.MakeJWT(token.UserID, cfg.jwtSecret)
		if err != nil {
			writeError(w, "could not obtain a new access token", 401)
			return
		}
		dat, _ := json.Marshal(accessTokenResponse{Token: accessToken})
		w.Write(dat)
	})
}

func (cfg *apiConfig) revokeTokenHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		bearerToken, bearerErr := auth.GetBearerToken(r.Header)
		if bearerErr != nil {
			writeError(w, "could not obtain a bearer token", 400)
			return
		}
		err := cfg.db.RevokeToken(r.Context(), bearerToken)
		if err != nil {
			writeError(w, "could not revoke the refresh token", 400)
			return
		}
		w.WriteHeader(204)
	})
}

func (cfg *apiConfig) updateUserHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		bearerToken, bearerErr := auth.GetBearerToken(r.Header)
		if bearerErr != nil {
			writeError(w, "could not obtain a bearer token", 401)
			return
		}
		userID, validationErr := auth.ValidateJWT(bearerToken, cfg.jwtSecret)
		if validationErr != nil {
			w.WriteHeader(401)
			return
		}
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
		updateParams := database.UpdateUserParams{
			ID:             userID,
			Email:          requestBody.Email,
			HashedPassword: hashed_pwd,
		}
		newUser, err := cfg.db.UpdateUser(r.Context(), updateParams)
		if err != nil {
			writeError(w, "error updating user", 400)
			return
		}
		userResponse := userResponse{
			ID:          newUser.ID,
			CreatedAt:   newUser.CreatedAt,
			UpdatedAt:   newUser.UpdatedAt,
			Email:       newUser.Email,
			IsChirpyRed: newUser.IsChirpyRed,
		}
		dat, _ := json.Marshal(userResponse)
		w.Write(dat)
	})
}

func (cfg *apiConfig) upgradeUserHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		apiKey, keyErr := auth.GetAPIKey(r.Header)
		if keyErr != nil {
			writeError(w, "could not obtain an API key", 401)
			return
		}
		if apiKey != cfg.polkaKey {
			w.WriteHeader(401)
			return
		}
		var requestBody webhookRequestBody
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			writeError(w, fmt.Sprintf("error decoding json: %s", err), 400)
			return
		}
		if requestBody.Event != "user.upgraded" {
			w.WriteHeader(204)
			return
		}
		if err := cfg.db.UnlockRed(r.Context(), requestBody.Data.UserID); err != nil {
			w.WriteHeader(404)
			return
		}
		w.WriteHeader(204)
	})
}
