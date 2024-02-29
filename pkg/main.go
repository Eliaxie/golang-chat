package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"golang-chat/pkg/model"
	"log"
	"net/http"

	//"os"
	"strings"

	"golang.org/x/net/websocket"
	"golang-chat/pkg/view"
)

// create a variable model of type Model
var globModel model.Model = model.Model{}

func messageHandler(ws *websocket.Conn) {
	globModel.Clients[model.NewClient(ws)] = true

	for {
		var data []byte
		err := websocket.Message.Receive(ws, &data)
		if err != nil {
			log.Println(err)
			delete(globModel.Clients, model.NewClient(ws))
			break
		}

		var msg model.Message
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Println("Error deserializing message:", err)
			continue
		}

		// Handle based on message type
		switch msg.GetMessageType() {
		case "TEXT":
			var textMsg model.TextMessage
			if err := json.Unmarshal(data, &textMsg); err != nil {
				log.Println("Error parsing TextMessage:", err)
			} else {
				fmt.Println("Received text message:", textMsg.Content)
			}

		case "CONN_INIT":
			var connInitMsg model.ConnectionInitMessage
			if err := json.Unmarshal(data, &connInitMsg); err != nil {
				log.Println("Error parsing ConnectionInitMessage:", err)
			} else {
				fmt.Println("Received connection init from:", connInitMsg.ClientID)
			}

		default:
			log.Println("Unknown message type:", msg.GetMessageType())
		}
	}
}

func broadcast(msg string) {
	for c := range globModel.Clients {
		err := websocket.Message.Send(c.Ws, msg)
		if err != nil {
			log.Println(err)
			delete(globModel.Clients, c)
		}
	}
}

func startServer(port string) {
	http.Handle("/ws", websocket.Handler(messageHandler))
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func readMessages(ws *websocket.Conn) {
	for {
		var msg string
		if err := websocket.Message.Receive(ws, &msg); err != nil {
			log.Println(err)
			break
		}
		fmt.Println("Received message:", msg)
	}
}

// Helper function to send a message over the websocket
func sendMessage(ws *websocket.Conn, msg model.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return websocket.Message.Send(ws, string(data))
}

func connectAndCommunicate(reader *bufio.Reader) {
	fmt.Print("Enter WebSocket server address to connect to (e.g., ws://localhost:8080/ws): ")
	serverAddress, _ := reader.ReadString('\n')
	serverAddress = strings.TrimSpace(serverAddress)

	ws, err := websocket.Dial(serverAddress, "", "http://localhost")
	if err != nil {
		log.Fatal(err)
	}
	globModel.Clients[model.NewClient(ws)] = true

	go readMessages(ws)

	for {
		fmt.Print("Enter message to send: ")
		txt, _ := reader.ReadString('\n')
		msg := model.TextMessage{MessageType: model.TEXT, Content: txt}

		sendMessage(ws, msg)
	}
}

func main() {

	// initialize model
	globModel = model.Model{
		Clients:            make(map[model.Client]bool),
		GroupsBuffers:      make(map[model.GroupName][]model.PendingMessage),
		Groups:             make(map[model.GroupName][]model.Client),
		GroupsVectorClocks: make(map[model.GroupName]model.VectorClock),
	}

	// reader := bufio.NewReader(os.Stdin)
	// fmt.Print("Enter port number to start the server on (e.g., 8080): ")
	// port, _ := reader.ReadString('\n')
	// port = strings.TrimSpace(port)

	// fmt.Printf("Starting WebSocket server on port %s\n", port)
	// go startServer(port)

	// connectAndCommunicate(reader)
	view.Start()

}
