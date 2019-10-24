package main

import (
	"context"
	"go-auth/actions"
	"go-auth/configure"
	"go-auth/store"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
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
	if err := store.OpenDatabase(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Serve HTTP on %s", server.Addr)
	routes()

	go func() {
		if err = server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)

	store.CloseDatabase()
	log.Println("Server stopped")
}

func routes() {
	http.HandleFunc("/healthcheck", actions.Run(actions.Healthcheck, http.MethodGet))
	http.HandleFunc("/registration", actions.Run(actions.Registration, http.MethodPost))
}
