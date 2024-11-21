package main

import (
	"net/http"
)

func main() {
	srvMux := http.NewServeMux()
	srvMux.Handle("/", http.FileServer(http.Dir(".")))
	server := http.Server{
		Handler: srvMux,
		Addr: ":8080",
	}
	server.ListenAndServe()
}