package controller

import (
	"encoding/json"
	"golang-chat/pkg/model"
	"net/http"

	log "github.com/sirupsen/logrus"

	"golang.org/x/net/websocket"
)

var controller *Controller

func InitWebServer(port string, c *Controller) {
	controller = c
	go startServer(port)
}

func (c *Controller) multicastMessage(message model.Message, clients []model.Client) {
	data, _ := json.Marshal(message)
	for _, client := range clients {
		if client != c.Model.Myself {
			sendMessageSlave(c.Model.ClientWs[client.ConnectionString], data)
		}
	}
}

func (c *Controller) addNewConnectionSlave(origin string, serverAddress string) *model.Client {
	ws, err := websocket.Dial(serverAddress, "ws", origin)
	if err != nil {
		log.Fatal(err)
	}
	client := &model.Client{Proc_id: "", ConnectionString: serverAddress}
	c.Model.PendingClients[serverAddress] = client
	c.Model.ClientWs[serverAddress] = ws

	initializeClient(c.Model.Myself.Proc_id, client)
	go receiveLoop(ws, client)
	return client
}

func sendMessageSlave(ws *websocket.Conn, msg []byte) error {
	log.Debug("Slave: Sending message:", string(msg))
	return websocket.Message.Send(ws, msg)
}

func startServer(port string) {
	http.Handle("/ws", websocket.Handler(messageHandler))
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func messageHandler(ws *websocket.Conn) {
	client := &model.Client{Proc_id: "", ConnectionString: ws.Request().RemoteAddr}
	controller.Model.PendingClients[ws.Request().RemoteAddr] = client
	controller.Model.ClientWs[ws.Request().RemoteAddr] = ws

	receiveLoop(ws, client)
}

func receiveLoop(ws *websocket.Conn, client *model.Client) {
	for {
		var data []byte
		err := websocket.Message.Receive(ws, &data)
		if err != nil {
			log.Errorln(err)
			delete(controller.Model.Clients, *client)
			break
		}

		log.Infoln("Handled data: ", string(data))

		var msg model.BaseMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Error("Error deserializing message:", err)
			continue
		}

		// Handle based on message type
		switch msg.GetMessageType() {
		case model.TEXT:
			var textMsg model.TextMessage
			if err := json.Unmarshal(data, &textMsg); err != nil {
				log.Error("Error parsing TextMessage:", err)
			} else {
				controller.HandleTextMessage(textMsg, client)
			}

		case model.CONN_INIT:
			var connInitMsg model.ConnectionInitMessage
			if err := json.Unmarshal(data, &connInitMsg); err != nil {
				log.Error("Error parsing ConnectionInitMessage:", err)
			} else {
				controller.HandleConnectionInitMessage(connInitMsg, client)
			}

		case model.CONN_INIT_RESPONSE:
			var connInitRespMsg model.ConnectionInitResponseMessage
			if err := json.Unmarshal(data, &connInitRespMsg); err != nil {
				log.Error("Error parsing ConnectionInitResponseMessage:", err)
			} else {
				controller.HandleConnectionInitResponseMessage(connInitRespMsg, client)
			}

		case model.SYNC_PEERS:
			var syncPeersMsg model.SyncPeersMessage
			if err := json.Unmarshal(data, &syncPeersMsg); err != nil {
				log.Error("Error parsing SyncPeersMessage:", err)
			} else {
				controller.HandleSyncPeersMessage(syncPeersMsg, client)
			}

		case model.GROUP_CREATE:
			var groupCreateMsg model.GroupCreateMessage
			if err := json.Unmarshal(data, &groupCreateMsg); err != nil {
				log.Error("Error parsing GroupCreateMessage:", err)
			} else {
				controller.HandleGroupCreateMessage(groupCreateMsg, client)
			}

		case model.CONN_RESTORE:
			var connRestoreMsg model.ConnectionRestoreMessage
			if err := json.Unmarshal(data, &connRestoreMsg); err != nil {
				log.Error("Error parsing ConnectionRestoreMessage:", err)
			} else {
				controller.HandleConnectionRestoreMessage(connRestoreMsg, client)
			}

		case model.SYNC_PEERS_RESPONSE:
			var syncPeersRespMsg model.SyncPeersResponseMessage
			if err := json.Unmarshal(data, &syncPeersRespMsg); err != nil {
				log.Error("Error parsing SyncPeersResponseMessage:", err)
			} else {
				controller.HandleSyncPeersResponseMessage(syncPeersRespMsg, client)
			}

		case model.CONN_RESTORE_RESPONSE:
			var connRestoreRespMsg model.ConnectionRestoreResponseMessage
			if err := json.Unmarshal(data, &connRestoreRespMsg); err != nil {
				log.Error("Error parsing ConnectionRestoreResponseMessage:", err)
			} else {
				controller.HandleConnectionRestoreResponseMessage(connRestoreRespMsg, client)
			}
		case model.MESSAGE_ACK:
			var ackMsg model.MessageAck
			if err := json.Unmarshal(data, &ackMsg); err != nil {
				log.Error("Error parsing MessageAck:", err)
			} else {
				controller.HandleMessageAck(ackMsg, client)
			}
		default:
			log.Error("Unknown message type:", msg.GetMessageType())
		}
	}
}
