package main

import (
	"go-auth/actions"
	"log"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/healthcheck", actions.Healthcheck)

	s := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Fatal(s.ListenAndServe())
}
