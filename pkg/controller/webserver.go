package controller

import (
	"encoding/json"
	"fmt"
	"golang-chat/pkg/model"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

var globModel *model.Model = &model.Model{}

func InitWebServer(port string, model *model.Model) {
	globModel = model
	go startServer(port)
}

func (c *Controller) addNewConnectionSlave(serverAddress string) {
	ws, err := websocket.Dial(serverAddress, "", "http://localhost")
	if err != nil {
		log.Fatal(err)
	}
	client := model.NewClient(ws)
	c.Model.Clients[&client] = true
	initializeClient(ws, &client)
}

func sendMessageSlave(ws *websocket.Conn, msg model.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return websocket.Message.Send(ws, string(data))
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

func broadcast(msg string, globModel *model.Model) {
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

func messageHandler(ws *websocket.Conn) {
	client := model.NewClient(ws)
	globModel.Clients[&client] = true

	for {
		var data []byte
		err := websocket.Message.Receive(ws, &data)
		if err != nil {
			log.Println(err)
			delete(globModel.Clients, &client)
			break
		}

		log.Println(string(data))

		var msg model.Message
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Println("1Error deserializing message:", err)
			continue
		}

		// Handle based on message type
		switch msg.MessageType {
		case model.TEXT:
			var textMsg model.TextMessage
			if err := json.Unmarshal(data, &textMsg); err != nil {
				log.Println("Error parsing TextMessage:", err)
			} else {
				fmt.Println("Received text message:", textMsg.Content)
			}

		case model.CONN_INIT:
			var connInitMsg model.ConnectionInitMessage
			if err := json.Unmarshal(data, &connInitMsg); err != nil {
				log.Println("Error parsing ConnectionInitMessage:", err)
			} else {
				fmt.Println("Received connection init from:", connInitMsg.ClientID)
			}

		default:
			log.Println("Unknown message type:", msg.MessageType)
		}
	}
}
