package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// Upgrade HTTP connections to WebSocket protocol
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Track connected WebSocket clients
var clients = make(map[*websocket.Conn]bool)

// Broadcast channel for messages
var broadcast = make(chan []byte)

// WebSocket handler function
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	// Register new client
	clients[conn] = true
	log.Println("New WebSocket connection established")
	go sendNetworkInfo()

	// Ensure client is removed from the map when the function exits
	defer func() {
		delete(clients, conn)
		go sendNetworkInfo()
		log.Println("WebSocket connection closed")
	}()

	// Listen for messages from this client
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}
		log.Printf("Received message: %s\n", message)

		// Send message to broadcast channel
		broadcast <- message
	}
}

// Broadcast messages to all connected WebSocket clients
func handleMessages() {
	for {
		// Wait for a message on the broadcast channel
		message := <-broadcast

		// Send the message to every connected client
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Println("Write error:", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func sendNetworkInfo() {
	var connections = len(clients)

	var networkStatus string
	resp, err := http.Get("http://www.example.com")
	if err != nil {
		networkStatus = "disconnected"
	} else {
		networkStatus = "connected"
		resp.Body.Close()
	}

	message := fmt.Sprintf(`{
		"type": "network_info",
		"data": {
			"networkStatus": "%s",
			"connections": %d
		}
	}`, networkStatus, connections)

	broadcast <- []byte(message)
}

func main() {
	// Set up WebSocket endpoint
	http.HandleFunc("/ws", handleWebSocket)

	// Start handling messages in a separate goroutine
	go handleMessages()

	fmt.Println("WebSocket server started on :6672")
	log.Fatal(http.ListenAndServe(":6672", nil))
}
