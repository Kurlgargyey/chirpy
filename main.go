package main

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"strings"
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
	Body *string `json:"body" required:"true"`
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
			dat, _ := json.Marshal(ValidResponse{Valid: true})
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
