package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	channelsMu sync.RWMutex
	channels   = make(map[string]map[*websocket.Conn]bool)
	broadcast  = make(chan map[string][]byte)

	wg sync.WaitGroup
)

func handleMessages() {
	defer wg.Done()

	for {
		select {
		case msg := <-broadcast:
			channelsMu.RLock()
			conns, ok := channels[string(msg["channel_id"])]
			channelsMu.RUnlock()
			if ok {
				for conn := range conns {
					err := conn.WriteMessage(websocket.TextMessage, msg["message"])
					if err != nil {
						log.Println(err)
						// Remove the connection if there's an error writing to it
						channelsMu.Lock()
						delete(channels[string(msg["channel_id"])], conn)
						channelsMu.Unlock()
						conn.Close()
					}
				}
			}
		}
	}
}

func reader(conn *websocket.Conn, channelID string) {
	defer wg.Done()

	channelsMu.Lock()
	if _, ok := channels[channelID]; !ok {
		channels[channelID] = make(map[*websocket.Conn]bool)
	}
	channels[channelID][conn] = true
	channelsMu.Unlock()

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			channelsMu.Lock()
			delete(channels[channelID], conn)
			channelsMu.Unlock()
			conn.Close()
			return
		}
		log.Println(string(p))

		broadcast <- map[string][]byte{"channel_id": []byte(channelID), "message": p}
	}
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	channelID := extractChannelID(r.URL.Path)
	log.Printf("Client Connected to Channel %s\n", channelID)

	err = ws.WriteMessage(1, []byte("Hi Client!"))
	if err != nil {
		log.Println(err)
	}

	wg.Add(1)
	go reader(ws, channelID)
}
func publishMessage(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Extract the channel ID from the URL path
	channelID := extractChannelID(r.URL.Path)

	// Check if the required fields are present
	if channelID == "" {
		http.Error(w, "Invalid request format: Channel ID is missing", http.StatusBadRequest)
		return
	}

	var messageBytes []byte

	// Try to parse the body as JSON
	var payload map[string]interface{}
	err = json.Unmarshal(body, &payload)
	if err == nil {
		// If successful, check for a "message" field in the JSON
		if msg, ok := payload["message"]; ok {
			messageBytes, err = json.Marshal(msg)
			if err != nil {
				http.Error(w, "Failed to marshal JSON message", http.StatusInternalServerError)
				return
			}
		}
	} else {
		// If parsing as JSON fails, assume it's a simple string
		messageBytes = body
	}

	// Broadcast the message to the specified channel
	broadcast <- map[string][]byte{"channel_id": []byte(channelID), "message": messageBytes}

	// Respond with a success message
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Message published successfully"))
}

func extractChannelID(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 2 {
		return parts[2]
	}
	return ""
}

func setupRoutes() {
	http.HandleFunc("/publish/", publishMessage)
	http.HandleFunc("/ws/", wsEndpoint)
}

func main() {
	fmt.Println("Hello World")
	setupRoutes()

	wg.Add(1)
	go handleMessages()

	log.Fatal(http.ListenAndServe(":3500", nil))
	wg.Wait()
}
