package main

import (
	"log"
	"net/http"
	"websockets-course/internal/handlers"
)

// port is the port to listen on
const port = ":8080"

// main is the main function
func main() {
	// get our routes
	mux := routes()

	// start the goroutine to listen for pushes
	// to our channels
	log.Println("Starting websocket goroutine")
	go handlers.ListenForWS()

	log.Println("Starting application on port", port)

	// start the web server
	_ = http.ListenAndServe(port, mux)
}
