package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/Kurlgargyey/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	jwtSecret      string
}

func main() {
	//define objects
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	jwtSecret := os.Getenv("SECRET")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println("error connecting to database: %w", err)
		return
	}
	dbQueries := database.New(db)
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
		jwtSecret:      jwtSecret,
	}
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
	srvMux.Handle("POST /api/users", apiCfg.createUserHandler())
	srvMux.Handle("POST /api/login", apiCfg.loginHandler())
	srvMux.Handle("POST /api/chirps", apiCfg.createChirpHandler())
	srvMux.Handle("GET /api/chirps", apiCfg.getChirpsHandler())
	srvMux.Handle("GET /api/chirps/{chirpID}", apiCfg.getChirpHandler())
	srvMux.Handle("POST /api/refresh", apiCfg.refreshTokenHandler())

	//run server
	server := http.Server{
		Handler: srvMux,
		Addr:    ":8080",
	}
	server.ListenAndServe()
}
