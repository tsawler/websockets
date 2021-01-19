package main

import (
	"log"
	"net/http"
	"websockets-course/internal/handlers"
)

const port = ":8080"

func main() {

	mux := routes()

	go handlers.BroadcastMessage()

	log.Println("Starting application on port", port)

	_ = http.ListenAndServe(port, mux)
}
