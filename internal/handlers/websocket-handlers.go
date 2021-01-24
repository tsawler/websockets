package handlers

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sort"
)

// clients is a map of connected clients
var clients = make(map[WebSocketConnection]string)

// one channel for each action
var connectChan = make(chan WsPayload)
var broadcastChan = make(chan WsPayload)
var alertChan = make(chan WsPayload)
var whoIsThereChan = make(chan WsPayload)
var enterChan = make(chan WsPayload)
var leaveChan = make(chan WsPayload)
var userNameChan = make(chan WsPayload)

var wsChan = make(chan WsPayload)

// upgradeConnection is the upgraded connection
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
	Action         string              `json:"action"`
	Message        string              `json:"message"`
	MessageType    string              `json:"message_type"`
	SkipSender     bool                `json:"-"`
	CurrentConn    WebSocketConnection `json:"-"`
	ConnectedUsers []string            `json:"connected_users"`
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
	clients[conn] = ""

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
			// do nothing
		} else {
			payload.Conn = *conn
			wsChan <- payload
		}
	}
}

// ListenToWsChannel listens to all channels and pushes data to broadcast function
func ListenToWsChannel() {
	var response WsJsonResponse
	for {
		e := <-wsChan

		switch e.Action {
		case "broadcast":
			response.Action = "broadcast"
			response.SkipSender = false
			response.Message = fmt.Sprintf("<strong>%s:</strong> %s", e.UserName, e.Message)
			broadcastToAll(response)

		case "alert":
			response.Action = "alert"
			response.SkipSender = false
			response.Message = e.Message
			response.MessageType = e.MessageType
			broadcastToAll(response)

		case "list_users":
			response.Action = "list"
			response.SkipSender = false
			response.Message = e.Message
			broadcastToAll(response)

		case "connect":
			response.Action = "connected"
			response.SkipSender = false
			response.Message = e.Message
			broadcastToAll(response)

		case "entered":
			response.SkipSender = true
			response.CurrentConn = e.Conn
			response.Action = "entered"
			response.Message = `<small class="text-muted"><em>New user in room</em></small>`
			broadcastToAll(response)

		case "left":
			response.SkipSender = false
			response.CurrentConn = e.Conn
			response.Action = "left"
			response.Message = fmt.Sprintf(`<small class="text-muted"><em>%s left</em></small>`, e.UserName)
			broadcastToAll(response)

			delete(clients, e.Conn)
			userList := getUserNameList()
			response.Action = "list_users"
			response.ConnectedUsers = userList
			response.SkipSender = false
			broadcastToAll(response)

		case "username":
			userList := addToUserList(e.Conn, e.UserName)
			response.Action = "list_users"
			response.ConnectedUsers = userList
			response.SkipSender = false
			broadcastToAll(response)
		}
	}
}

func getUserNameList() []string {
	var userNames []string
	for _, value := range clients {
		if value != "" {
			userNames = append(userNames, value)
		}
	}
	sort.Strings(userNames)

	return userNames
}

func addToUserList(conn WebSocketConnection, u string) []string {
	var userNames []string
	clients[conn] = u
	for _, value := range clients {
		if value != "" {
			userNames = append(userNames, value)
		}
	}
	sort.Strings(userNames)

	return userNames
}

// broadcastToAll sends a response to all connected clients, as JSON
// note that the JSON will show up as part of the WS default json,
// under "data"
func broadcastToAll(response WsJsonResponse) {
	for client := range clients {
		// skip sender, if appropriate
		if response.SkipSender && response.CurrentConn == client {
			continue
		}

		// broadcast to every connected client
		err := client.WriteJSON(response)
		if err != nil {
			log.Printf("Websocket error on %s: %s", response.Action, err)
			_ = client.Close()
			delete(clients, client)
		}
	}
}
