package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var connections = make(map[*websocket.Conn]bool)
var broadcast = make(chan []byte)

func handleMessages() {
	for {
		// grab any message from the broadcast channel
		msg := <-broadcast

		// send it out to every client that is currently connected
		for conn := range connections {
			err := conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func reader(conn *websocket.Conn) {
	// add the new connection to the pool
	connections[conn] = true

	for {
		// read in a message
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			delete(connections, conn)
			conn.Close()
			return
		}
		// print out that message for clarity
		log.Println(string(p))

		// broadcast the received message to all connected clients
		broadcast <- p
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home Page")
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	// upgrade this connection to a WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Client Connected")
	err = ws.WriteMessage(1, []byte("Hi Client!")) // Fix the syntax here
	if err != nil {
		log.Println(err)
	}

	reader(ws)
}

func setupRoutes() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/ws", wsEndpoint)
}

func main() {
	fmt.Println("Hello World")
	setupRoutes()

	// Start a goroutine to handle incoming messages and broadcast them to all clients
	go handleMessages()

	log.Fatal(http.ListenAndServe(":3500", nil))
}
