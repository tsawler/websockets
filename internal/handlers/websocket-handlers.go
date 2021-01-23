package handlers

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

const writeWait = 2 * time.Second

// clients is a map of connected clients
var clients = make(map[*websocket.Conn]bool)

// one channel for each action (and we only have two right now)
var connectChan = make(chan WsPayload)
var broadcastChan = make(chan WsPayload)
var alertChan = make(chan WsPayload)
var whoIsThereChan = make(chan WsPayload)

// upgradeConnection is the websocket upgrader
var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// WsPayload defines the data we receive from the client
type WsPayload struct {
	Action      string `json:"action"`
	Message     string `json:"message"`
	UserName    string `json:"username"`
	MessageType string `json:"message_type"`
}

// WsJsonResponse defines the json we send back to client
type WsJsonResponse struct {
	Action      string `json:"action"`
	Message     string `json:"message"`
	MessageType string `json:"message_type"`
}

// WebSocketConnection holds the websocket connection
type WebSocketConnection struct {
	*websocket.Conn
}

// WsEndPoint handles websocket connections
func WsEndPoint(w http.ResponseWriter, r *http.Request) {
	log.Println("Hit the WsEndPoint")
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

	go ListenForWS(&conn)

}

// ListenForWS is the goroutine that listens for our channels
func ListenForWS(conn *WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("ERROR", fmt.Sprintf("%v", r))
		}
	}()

	var payload WsPayload

	for {
		err := conn.ReadJSON(&payload)
		if err != nil {
			// do nothing; it's a connection
		} else {
			// send payload to appropriate channel
			switch payload.Action {
			case "broadcast":
				broadcastChan <- payload
			case "alert":
				log.Println("Sending", payload.Message, "to alert chan")
				alertChan <- payload
			}
		}
	}

}

// ListenToChannels listens to all channels and pushes data to broadcast function
func ListenToChannels() {
	var response WsJsonResponse
	for {
		select {
		// message to send to everyone from a user
		case b := <-broadcastChan:
			response.Action = "broadcast"
			response.Message = fmt.Sprintf("<strong>%s:</strong> %s", b.UserName, b.Message)
			log.Println("Sending broadcast of", response.Message)
			broadcastToAll(response)
		// send an alert
		case a := <-alertChan:
			response.Action = "alert"
			response.Message = a.Message
			response.MessageType = a.MessageType
			broadcastToAll(response)
		// list users in chat
		case w := <-whoIsThereChan:
			response.Action = "list"
			response.Message = w.Message
			broadcastToAll(response)
		// someone connected
		case c := <-connectChan:
			response.Action = "connected"
			response.Message = c.Message
			broadcastToAll(response)

			// do one for enter and one for leave
		}

	}
}

// broadcastToAll sends a response to all connected clients, as JSON
// note that the JSON will show up as part of the WS default json,
// under "data"
func broadcastToAll(response WsJsonResponse) {
	for client := range clients {
		err := client.WriteJSON(response)
		if err != nil {
			log.Printf("Websocket error on %s: %s", response.Action, err)
			_ = client.Close()
			delete(clients, client)
		}
	}
}
