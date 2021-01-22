package handlers

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

// clients is a map of connected clients
var clients = make(map[*websocket.Conn]bool)

// upgradeConnection is the websocket upgrader
var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type WsPayload struct {
	Action   string `json:"action"`
	Message  string `json:"message"`
	UserName string `json:"username"`
}

type WsJsonResponse struct {
	Action      string `json:"action"`
	Message     string `json:"message"`
	MessageType string `json:"message_type"`
}

type WebSocketConnection struct {
	*websocket.Conn
}

// WsEndPoint handles websocket connections
func WsEndPoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Println(fmt.Sprintf("Client Connected from %s", r.RemoteAddr))
	var response WsJsonResponse
	response.Message = "<em><small>Connected to server ... </small></em>"

	err = ws.WriteJSON(response)
	if err != nil {
		log.Println(err)
	}

	conn := WebSocketConnection{Conn: ws}
	clients[ws] = true

	go HandleConnectionAction(&conn)
}

// HandleConnectionAction broadcasts messages to all connected clients
func HandleConnectionAction(conn *WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("ERROR", fmt.Sprintf("%v", r))
		}
	}()

	for {
		payload := WsPayload{}

		err := conn.ReadJSON(&payload)
		if err != nil {
			log.Println(err)
			continue
		}

		var response WsJsonResponse

		for client := range clients {
			if payload.Action == "broadcast" {
				response.Action = "broadcast"
				response.Message = fmt.Sprintf("<strong>%s:</strong> %s", payload.UserName, payload.Message)

				err := client.WriteJSON(response)
				if err != nil {
					log.Printf("Websocket error on broadcast: %s", err)
					_ = client.Close()
					delete(clients, client)
				}
			} else if payload.Action == "alert" {
				response.Action = "alert"
				response.Message = payload.Message
				response.MessageType = "success"
				err := client.WriteJSON(response)
				if err != nil {
					log.Printf("Websocket error on alert: %s", err)
					_ = client.Close()
					delete(clients, client)
				}
			}
		}
	}
}
