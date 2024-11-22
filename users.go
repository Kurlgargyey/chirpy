package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type userRequestBody struct {
	Email string `json:"email"`
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
			user, err := cfg.db.CreateUser(r.Context(), requestBody.Email)
			if err != nil {
				writeError(w, fmt.Sprintf("error creating user: %s", err), 400)
				return
			}
			dat, _ := json.Marshal(user)
			w.Header()
			w.Write(dat)
		})
}
