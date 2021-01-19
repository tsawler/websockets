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

	mux.Get("/ws", http.HandlerFunc(handlers.WsEndPoint))
	mux.Post("/ws/send", http.HandlerFunc(handlers.WsSend))

	return mux

}
