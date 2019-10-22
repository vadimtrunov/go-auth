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

	configPath := "cnf/server.cnf"
	server, err := configure.HTTPServer(configPath)
	if err != nil {
		log.Printf("Can't find CNF file. Use default configiration. To use your own configuration create file: '%s'/n", configPath)
		server = &http.Server{Addr: ":8080"}
	}

	log.Printf("Serve HTTP on %s", server.Addr)
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Server stopped")
}
