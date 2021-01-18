package main

import (
	"github.com/bmizerany/pat"
	"net/http"
	"websockets-course/internal/handlers"
)

// routes handles application routes
func routes() http.Handler {
	mux := pat.New()

	mux.Get("/", http.HandlerFunc(handlers.Home))
	mux.Get("/send", http.HandlerFunc(handlers.SendData))

	return mux
}
