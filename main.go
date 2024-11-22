package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			cfg.fileserverHits.Add(1)
			next.ServeHTTP(w, req)
		})
}

func (cfg *apiConfig) metricsHandler() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "text/html")
			w.Write([]byte(fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())))
		})
}

func (cfg *apiConfig) resetHandler() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			cfg.fileserverHits.Store(0)
		})
}

type Chirp struct {
	Body string
}

type ValidationError struct {
	Error string `json:"error"`
}
type ValidResponse struct {
	Valid bool `json:"valid"`
}

func validateChirpHandler() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			body := r.Body
			var chirp Chirp
			if err := json.NewDecoder(body).Decode(&chirp); err != nil {
				writeError(w, fmt.Sprintf("%s", err))
				return
			}
			if len(chirp.Body) > 140 {
				writeError(w, "overlong chirp")
				return
			}
			if len(chirp.Body) == 0 {
				writeError(w, "empty chirp")
				return
			}
			dat, _ := json.Marshal(ValidResponse{Valid: true})
			w.WriteHeader(200)
			w.Write(dat)
		})
}

func writeError(w http.ResponseWriter, err string) {
	response := ValidationError{Error: err}
	dat, _ := json.Marshal(response)
	w.WriteHeader(400)
	w.Write(dat)
}

func main() {
	//define objects
	apiCfg := apiConfig{}
	srvMux := http.NewServeMux()
	fileServer := http.StripPrefix("/app",
		http.FileServer(http.Dir(".")))

	//register routes!
	srvMux.Handle("/app/",
		apiCfg.middlewareMetricsInc(fileServer))
	srvMux.HandleFunc("GET /api/healthz",
		func(response http.ResponseWriter, req *http.Request) {
			response.Header().Add("Content-Type", "text/plain; charset=utf-8")
			response.Write([]byte("OK"))
		})
	srvMux.Handle("GET /admin/metrics", apiCfg.metricsHandler())
	srvMux.Handle("POST /admin/reset", apiCfg.resetHandler())
	srvMux.Handle("POST /api/validate_chirp", validateChirpHandler())

	//run server
	server := http.Server{
		Handler: srvMux,
		Addr:    ":8080",
	}
	server.ListenAndServe()
}
