package main

import (
	"go-auth/actions"
	"go-auth/configure"
	"log"
	"net/http"
)

func main() {
	log.Println("Starting server. Reading configs...")
	http.HandleFunc("/healthcheck", actions.Healthcheck)
	server, err := configure.HTTPServer("cnf/server.cnf")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Serve HTTP on %s", server.Addr)
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Server stopped")
}
