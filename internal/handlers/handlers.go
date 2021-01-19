package handlers

import (
	"fmt"
	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var broadcast = make(chan string)

var clients = make(map[*websocket.Conn]bool)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var views = jet.NewSet(
	jet.NewOSFileSystemLoader("./html"),
	jet.InDevelopmentMode(),
)

// Home displays the home page with some sample data
func Home(w http.ResponseWriter, r *http.Request) {
	err := renderPage(w, "home.jet", nil)
	if err != nil {
		_, _ = fmt.Fprint(w, "Error executing template:", err)
	}
}

// SendData displays the page for sending data via websockets
func SendData(w http.ResponseWriter, r *http.Request) {
	err := renderPage(w, "send.jet", nil)
	if err != nil {
		_, _ = fmt.Fprint(w, "Error executing template:", err)
	}
}

// WsEndPoint handles websocket connections
func WsEndPoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Println(fmt.Sprintf("Client Connected from %s", r.RemoteAddr))

	err = ws.WriteMessage(1, []byte("<em><small>Connected to server ... </small></em>"))
	if err != nil {
		log.Println(err)
	}

	clients[ws] = true

	WsReader(ws)
}

// WsSend handles posting a message and broadcasting it
func WsSend(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		return
	}

	for n, v := range r.PostForm {
		log.Println("n", n, "v", v)
	}

	msg := r.Form.Get("payload")
	log.Println("pushed", msg, "to channel")
	broadcast <- msg

	http.Redirect(w, r, "/send", http.StatusSeeOther)
}

// WsReader broadcasts messages
func WsReader(conn *websocket.Conn) {
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

// renderPage renders the page using Jet templates
func renderPage(w http.ResponseWriter, tmpl string, data jet.VarMap) error {
	view, err := views.GetTemplate(tmpl)
	if err != nil {
		log.Println("Unexpected template err:", err.Error())
		return err
	}

	err = view.Execute(w, data, nil)
	if err != nil {
		log.Println("Error executing template:", err)
		return err
	}
	return nil
}
