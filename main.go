package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/websocket"
)

// Client structure
type client struct {
	ws *websocket.Conn // WebSocket connection
}

var clients = make(map[client]bool)

func broadcast(msg string) {
	for c := range clients {
		err := websocket.Message.Send(c.ws, msg)
		if err != nil {
			log.Println(err)
			// Handle client disconnection on error
			delete(clients, c)
		}
	}
}

func messageHandler(ws *websocket.Conn) {
	clients[client{ws}] = true

	for {
		var msg string
		err := websocket.Message.Receive(ws, &msg)
		if err != nil {
			log.Println(err)
			// Handle client disconnection on error
			delete(clients, client{ws})
			break
		}
		broadcast(msg)
	}
}

func main() {
	http.Handle("/ws", websocket.Handler(messageHandler))

	// Read user input for the port number
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter port number to start the server on (e.g., 8080): ")
	port, _ := reader.ReadString('\n')
	port = strings.TrimSpace(port)

	fmt.Printf("Starting WebSocket server on port %s\n", port)

	go func() {
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatal("ListenAndServe:", err)
		}
	}()

	// Wait for the server to start
	time.Sleep(2 * time.Second)

	// Create server's own WebSocket connection
	serverWs, err := websocket.Dial("ws://localhost:"+port+"/ws", "", "http://localhost")
	if err != nil {
		log.Fatal(err)
	}
	clients[client{serverWs}] = true

	go func() {
		// Read user input for the WebSocket server address
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter WebSocket server address to connect to (e.g., ws://localhost:8080/ws): ")

		var ws *websocket.Conn
		for {
			serverAddress, _ := reader.ReadString('\n')
			serverAddress = strings.TrimSpace(serverAddress)

			// Connect to the WebSocket server
			ws, err := websocket.Dial(serverAddress, "", "http://localhost")
			if err == nil {
				clients[client{ws}] = true
				break
			}
			log.Println(err)
			log.Println("Please try again:")
		}

		// Goroutine to read messages from the WebSocket connection
		go func() {
			for {
				var msg string
				if err := websocket.Message.Receive(ws, &msg); err != nil {
					log.Println(err)
					break
				}
				fmt.Println("Received message:", msg)
			}
		}()

		// Read user input and send it over the WebSocket connection
		for {
			fmt.Print("Enter message to send: ")
			msg, _ := reader.ReadString('\n')
			msg = strings.TrimSpace(msg)

			if err := websocket.Message.Send(ws, msg); err != nil {
				log.Println(err)
				break
			}
		}
	}()

	// Block the main goroutine
	select {}
}
