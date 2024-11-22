package main

import (
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
			w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())))
		})
}

func (cfg *apiConfig) resetHandler() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			cfg.fileserverHits.Store(0)
		})
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
	srvMux.HandleFunc("GET /healthz",
		func(response http.ResponseWriter, req *http.Request) {
			response.Header().Add("Content-Type", "text/plain; charset=utf-8")
			response.Write([]byte("OK"))
		})
	srvMux.Handle("GET /metrics", apiCfg.metricsHandler())
	srvMux.Handle("/reset", apiCfg.resetHandler())

	//run server
	server := http.Server{
		Handler: srvMux,
		Addr:    ":8080",
	}
	server.ListenAndServe()
}
