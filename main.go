package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bbutkovic/deezer-remote/handlers"
	"github.com/bbutkovic/deezer-remote/hub"
	"github.com/gorilla/mux"
)

func main() {
	hub := hub.NewHub()
	go hub.Run()
	router := mux.NewRouter()

	//This is the base of our API
	api := router.PathPrefix("/api/").Subrouter()
	api.HandleFunc("/player/token", func(w http.ResponseWriter, r *http.Request) {
		handlers.NewTokenHandler(hub, w, r)
	}).Methods("GET")
	api.HandleFunc("/player/{token}/ws", func(w http.ResponseWriter, r *http.Request) {
		handlers.PlayerWSHandler(hub, w, r)
	})
	api.HandleFunc("/player/{token}", func(w http.ResponseWriter, r *http.Request) {
		handlers.SendPlayerCommand(hub, w, r)
	}).Methods("POST")

	fs := http.FileServer(http.Dir("dist/"))
	router.PathPrefix("/").Handler(staticHandler(fs))

	//Initialize the web server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Starting web server at port " + port)

	server := &http.Server{
		Handler:      router,
		Addr:         ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

//Used for Single Page Application frontend route handler
//In case the route seems to point to an endpoint instead of a file simply serve index.html instead of 404ing
func staticHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqPath := r.URL.Path

		if strings.Contains(reqPath, ".") || reqPath == "/" {
			next.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, "dist/index.html")
	})
}
