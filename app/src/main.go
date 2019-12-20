package main

import (
	"go-auth/actions"
	"go-auth/configure"
	"go-auth/store"
	"log"
	"net/http"
)

func main() {
	log.Print("Starting service. Reading configs...")
	configPath := "cnf/server.cnf"
	server, err := configure.HTTPServer(configPath)
	if err != nil {
		log.Printf("Can't find CNF file. Use default configiration. To use your own configuration create file: '%s'/n", configPath)
		server = &http.Server{Addr: ":8080"}
	}

	log.Print("Oppening persistent DB connection...")
	if err := store.OpenDatabase("data/store.db"); err != nil {
		log.Fatal(err)
	}

	log.Printf("Serve HTTP on %s", server.Addr)
	routes()

	if err = server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
	store.CloseDatabase()
	log.Println("Server stopped")
}

func routes() {
	http.HandleFunc("/healthcheck", actions.Run(actions.Healthcheck, http.MethodGet))
	http.HandleFunc("/registration", actions.Run(actions.Registration, http.MethodPost))
	http.HandleFunc("/login", actions.Run(actions.Login, http.MethodPost))
}
