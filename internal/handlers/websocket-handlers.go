package handlers

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var broadcast = make(chan string)

var clients = make(map[*websocket.Conn]bool)

var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// WsEndPoint handles websocket connections
func WsEndPoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Println(fmt.Sprintf("Client Connected from %s", r.RemoteAddr))

	err = ws.WriteMessage(1, []byte("<em><small>Connected to server ... </small></em>"))
	if err != nil {
		log.Println(err)
	}

	clients[ws] = true
}

// WsSend handles posting a message and broadcasting it
func WsSend(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		return
	}

	msg := r.Form.Get("payload")

	broadcast <- msg

	http.Redirect(w, r, "/send", http.StatusSeeOther)
}

// BroadcastMessage broadcasts messages to all connected clients
func BroadcastMessage() {
	for {
		val := <-broadcast
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(val))
			if err != nil {
				log.Printf("Websocket error: %s", err)
				_ = client.Close()
				delete(clients, client)
			}
		}
	}
}
