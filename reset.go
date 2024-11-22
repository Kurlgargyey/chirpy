package main

import (
	"encoding/json"
	"net/http"
)

func (cfg *apiConfig) resetHandler() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if cfg.platform != "dev" {
				writeError(w, "Access denied", 403)
				return
			}
			cfg.fileserverHits.Store(0)
			result, err := cfg.db.WipeUsers(r.Context())
			if err != nil {
				writeError(w, "error wiping users", 400)
				return
			}
			dat, _ := json.Marshal(result)
			w.Write(dat)
		})
}
