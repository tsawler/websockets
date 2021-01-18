package main

import (
	"log"
	"net/http"
)

const port = ":8080"

func main() {

	mux := routes()

	log.Println("Staring application on port", port)

	_ = http.ListenAndServe(port, mux)
}
