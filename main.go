package main

import (
	"net/http"
)

func main() {
	srvMux := http.NewServeMux()
	srvMux.Handle("/app/",
		http.StripPrefix("/app",
			http.FileServer(http.Dir("."))))
	srvMux.HandleFunc("/healthz",
		func(response http.ResponseWriter, req *http.Request) {
			response.Header().Add("Content-Type", "text/plain; charset=utf-8")
			response.Write([]byte("OK"))
		})
	server := http.Server{
		Handler: srvMux,
		Addr:    ":8080",
	}
	server.ListenAndServe()
}
