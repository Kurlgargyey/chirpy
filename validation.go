package main

import (
	"net/http"
	"encoding/json"
	"mime"
	"strings"
	"fmt"
)

type Chirp struct {
	Body *string `json:"body" required:"true"`
}

type ValidationError struct {
	Error string `json:"error"`
}
type CleanedChirp struct {
	CleanedBody string `json:"cleaned_body"`
}

func validateChirpHandler() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			maxChirpLength := 140
			w.Header().Add("Content-Type", "application/json")
			contentType := r.Header.Get("Content-Type")
			mediaType, _, err := mime.ParseMediaType(contentType)
			if err != nil || mediaType != "application/json" || contentType == "" {
				writeError(w, "Content-Type must be application/json", 415)
				return
			}
			body := r.Body
			var chirp Chirp
			decoder := json.NewDecoder(body)
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&chirp); err != nil {
				writeError(w, fmt.Sprintf("%s", err), 400)
				return
			}
			if chirp.Body == nil {
				writeError(w, "missing required fields: body", 400)
				return
			}
			*chirp.Body = strings.TrimSpace(*chirp.Body)
			if len(*chirp.Body) > maxChirpLength {
				writeError(w, "overlong chirp", 422)
				return
			}
			if len(*chirp.Body) == 0 {
				writeError(w, "empty chirp", 422)
				return
			}
			*chirp.Body = strings.ReplaceAll(*chirp.Body, "kerfluffle", "****")
			*chirp.Body = strings.ReplaceAll(*chirp.Body, "sharbert", "****")
			*chirp.Body = strings.ReplaceAll(*chirp.Body, "fornax", "****")
			dat, _ := json.Marshal(CleanedChirp{CleanedBody: *chirp.Body})
			w.WriteHeader(200)
			w.Write(dat)
		})
}

func writeError(w http.ResponseWriter, err string, code int) {
	response := ValidationError{Error: err}
	dat, _ := json.Marshal(response)
	w.WriteHeader(code)
	w.Write(dat)
}