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
var connectChan = make(chan WsPayload)
var broadcastChan = make(chan WsPayload)
var alertChan = make(chan WsPayload)
var whoIsThereChan = make(chan WsPayload)
var enterChan = make(chan WsPayload)
var leaveChan = make(chan WsPayload)

// upgradeConnection is the websocket upgrader
var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// WsPayload defines the data we receive from the client
type WsPayload struct {
	Action      string              `json:"action"`
	Message     string              `json:"message"`
	UserName    string              `json:"username"`
	MessageType string              `json:"message_type"`
	Conn        WebSocketConnection `json:"-"`
}

// WsJsonResponse defines the json we send back to client
type WsJsonResponse struct {
	Action      string              `json:"action"`
	Message     string              `json:"message"`
	MessageType string              `json:"message_type"`
	SkipSender  bool                `json:"-"`
	CurrentConn WebSocketConnection `json:"-"`
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
			payload.Conn = *conn

			// send payload to appropriate channel
			switch payload.Action {
			case "broadcast":
				broadcastChan <- payload
			case "alert":
				alertChan <- payload
			case "entered":
				enterChan <- payload
			case "left":
				leaveChan <- payload
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
		case e := <-broadcastChan:
			response.Action = "broadcast"
			response.Message = fmt.Sprintf("<strong>%s:</strong> %s", e.UserName, e.Message)
			log.Println("Sending broadcast of", response.Message)
			broadcastToAll(response)
		// send an alert
		case e := <-alertChan:
			response.Action = "alert"
			response.Message = e.Message
			response.MessageType = e.MessageType
			broadcastToAll(response)
		// list users in chat
		case e := <-whoIsThereChan:
			response.Action = "list"
			response.Message = e.Message
			broadcastToAll(response)
		// someone connected
		case e := <-connectChan:
			response.Action = "connected"
			response.Message = e.Message
			broadcastToAll(response)
		// someone entered
		case e := <-enterChan:
			response.SkipSender = true
			response.CurrentConn = e.Conn
			response.Action = "entered"
			response.Message = `<small class="text-muted"><em>New user in room</em></small>`
			broadcastToAll(response)
		// someone left
		case e := <-leaveChan:
			response.SkipSender = true
			response.CurrentConn = e.Conn
			response.Action = "left"
			response.Message = fmt.Sprintf(`<small class="text-muted"><em>%s entered</em></small>`, e.UserName)
			broadcastToAll(response)
		}

	}
}

// broadcastToAll sends a response to all connected clients, as JSON
// note that the JSON will show up as part of the WS default json,
// under "data"
func broadcastToAll(response WsJsonResponse) {
	for client := range clients {

		if response.SkipSender && response.CurrentConn.Conn == client {
			continue
		}

		err := client.WriteJSON(response)
		if err != nil {
			log.Printf("Websocket error on %s: %s", response.Action, err)
			_ = client.Close()
			delete(clients, client)
		}
	}
}
