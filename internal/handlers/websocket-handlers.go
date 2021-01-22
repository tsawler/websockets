package handlers

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

// clients is a map of connected clients
var clients = make(map[*websocket.Conn]bool)

// one channel for each action (and we only have two right now)
var broadcastChan = make(chan WsPayload)
var alertChan = make(chan WsPayload)

// upgradeConnection is the websocket upgrader
var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// WsPayload defines the data we receive from teh client
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

	var payload WsPayload

	err = conn.ReadJSON(&payload)
	if err != nil {
		log.Println(err)
		return
	}

	// send payload to appropriate channel
	switch payload.Action {
	case "broadcast":
		broadcastChan <- payload
	case "alert":
		alertChan <- payload
	default:
		// do nothing
	}
}

// ListenForWS is the goroutine that listens for our channels
func ListenForWS() {
	// when this function closes, for whatever reason, recover it
	// and write an error message
	defer func() {
		if r := recover(); r != nil {
			log.Println("ERROR", fmt.Sprintf("%v", r))
		}
	}()

	// this will hold the response we send to connected clients
	var response WsJsonResponse

	// run forever
	for {
		// when we get an entry into the channel, send back the appropriate
		// response
		select {
		case b := <-broadcastChan:
			response.Action = "broadcast"
			response.Message = fmt.Sprintf("<strong>%s:</strong> %s", b.UserName, b.Message)
			broadcastToAll(response)

		case a := <-alertChan:
			response.Action = "alert"
			response.Message = a.Message
			response.MessageType = a.MessageType
			broadcastToAll(response)

		default:
			// do nothing
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
			log.Printf("Websocket error on alert: %s", err)
			_ = client.Close()
			delete(clients, client)
		}
	}
}
