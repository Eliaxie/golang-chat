package controller

import (
	"encoding/json"
	"golang-chat/pkg/model"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

var globModel *model.Model
var controller *Controller

func InitWebServer(port string, c *Controller) {
	globModel = c.Model
	controller = c
	go startServer(port)
}

func (c *Controller) multicastMessage(message model.Message, clients []model.Client) {
	data, _ := json.Marshal(message)
	for _, client := range clients {
		sendMessageSlave(client.Ws, data)
	}
}

func (c *Controller) addNewConnectionSlave(serverAddress string) model.Client {
	ws, err := websocket.Dial(serverAddress, "", "http://localhost")
	if err != nil {
		log.Fatal(err)
	}
	client := model.NewClient(ws)
	c.Model.PendingClients[client] = true
	initializeClient(&client)
	go receiveLoop(ws, client)
	return client
}

func sendMessageSlave(ws *websocket.Conn, msg []byte) error {
	log.Println("Slave: Sending message:", string(msg))
	return websocket.Message.Send(ws, msg)
}

func startServer(port string) {
	http.Handle("/ws", websocket.Handler(messageHandler))
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func messageHandler(ws *websocket.Conn) {
	client := model.NewClient(ws)
	globModel.PendingClients[client] = true

	receiveLoop(ws, client)
}

func receiveLoop(ws *websocket.Conn, client model.Client) {
	for {
		var data []byte
		err := websocket.Message.Receive(ws, &data)
		if err != nil {
			log.Println(err)
			delete(globModel.Clients, client)
			break
		}

		log.Println("Handled data: ", string(data))

		var msg model.BaseMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Println("Error deserializing message:", err)
			continue
		}

		// Handle based on message type
		switch msg.GetMessageType() {
		case model.TEXT:
			var textMsg model.TextMessage
			if err := json.Unmarshal(data, &textMsg); err != nil {
				log.Println("Error parsing TextMessage:", err)
			} else {
				controller.HandleTextMessage(textMsg, client)
			}

		case model.CONN_INIT:
			var connInitMsg model.ConnectionInitMessage
			if err := json.Unmarshal(data, &connInitMsg); err != nil {
				log.Println("Error parsing ConnectionInitMessage:", err)
			} else {
				controller.HandleConnectionInitMessage(connInitMsg, client)
			}

		case model.CONN_INIT_RESPONSE:
			var connInitRespMsg model.ConnectionInitResponseMessage
			if err := json.Unmarshal(data, &connInitRespMsg); err != nil {
				log.Println("Error parsing ConnectionInitResponseMessage:", err)
			} else {
				controller.HandleConnectionInitResponseMessage(connInitRespMsg, client)
			}

		case model.SYNC_PEERS:
			var syncPeersMsg model.SyncPeersMessage
			if err := json.Unmarshal(data, &syncPeersMsg); err != nil {
				log.Println("Error parsing SyncPeersMessage:", err)
			} else {
				controller.HandleSyncPeersMessage(syncPeersMsg, client)
			}

		case model.GROUP_CREATE:
			var groupCreateMsg model.GroupCreateMessage
			if err := json.Unmarshal(data, &groupCreateMsg); err != nil {
				log.Println("Error parsing GroupCreateMessage:", err)
			} else {
				controller.HandleGroupCreateMessage(groupCreateMsg, client)
			}

		case model.CONN_RESTORE:
			var connRestoreMsg model.ConnectionRestoreMessage
			if err := json.Unmarshal(data, &connRestoreMsg); err != nil {
				log.Println("Error parsing ConnectionRestoreMessage:", err)
			} else {
				controller.HandleConnectionRestoreMessage(connRestoreMsg, client)
			}

		case model.SYNC_PEERS_RESPONSE:
			var syncPeersRespMsg model.SyncPeersResponseMessage
			if err := json.Unmarshal(data, &syncPeersRespMsg); err != nil {
				log.Println("Error parsing SyncPeersResponseMessage:", err)
			} else {
				controller.HandleSyncPeersResponseMessage(syncPeersRespMsg, client)
			}

		case model.CONN_RESTORE_RESPONSE:
			var connRestoreRespMsg model.ConnectionRestoreResponseMessage
			if err := json.Unmarshal(data, &connRestoreRespMsg); err != nil {
				log.Println("Error parsing ConnectionRestoreResponseMessage:", err)
			} else {
				controller.HandleConnectionRestoreResponseMessage(connRestoreRespMsg, client)
			}

		default:
			log.Println("Unknown message type:", msg.GetMessageType())
		}
	}
}
