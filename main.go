package main

import (
	"net/http"
)

func main() {
	srvMux := http.NewServeMux()
	server := http.Server{
		Handler: srvMux,
		Addr: ":8080",
	}
	server.ListenAndServe()
}