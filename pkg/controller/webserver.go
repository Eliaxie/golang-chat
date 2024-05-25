package controller

import (
	"encoding/json"
	"golang-chat/pkg/model"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
)

var controller *Controller

func InitWebServer(port string, c *Controller) {
	controller = c
	go startServer(port)
}

func (c *Controller) multicastMessage(message model.Message, clients []model.Client) {
	// print list of clients
	log.Debugln("Multicasting message to clients: ", clients)
	for _, client := range clients {
		if client != c.Model.Myself {
			c.SendMessage(message, client)
		}
	}
}

func (c *Controller) addNewConnectionSlave(origin string, serverAddress string, reconnection bool) (*model.Client, error) {
	header := http.Header{}
	header.Set("origin", origin)
	ws, _, err := websocket.DefaultDialer.Dial(serverAddress, header)
	if err != nil {
		log.Trace("Error dialing:", err)
		return nil, err
	}

	client := &model.Client{Proc_id: "", ConnectionString: serverAddress}
	c.Model.PendingClients[serverAddress] = struct{}{}
	c.Model.ClientWs[serverAddress] = ws
	c.Model.MessageExitBuffer[*client] = make([][]byte, 0)

	initializeClient(c.Model.Myself.Proc_id, client, reconnection)
	go receiveLoop(ws, client)
	go ping(ws)
	return client, nil
}

func ping(ws *websocket.Conn) {
	for {
		time.Sleep(200 * time.Millisecond)
		if err := ws.WriteMessage(2, []byte("ping")); err != nil {
			log.Errorln(err)
			return
		}
	}
}

func sendMessageSlave(ws *websocket.Conn, client model.Client) error {
	controller.Model.MessageExitBufferLock.Lock()
	defer controller.Model.MessageExitBufferLock.Unlock()
	for _, msg := range controller.Model.MessageExitBuffer[client] {
		if err := ws.WriteMessage(1, msg); err != nil {
			log.Errorln(err)
			return err
		}
		controller.Model.MessageExitBuffer[client] = controller.Model.MessageExitBuffer[client][1:]
		log.Debugln("Slave: Sent message:", string(msg))
	}
	return nil
}

func startServer(port string) {
	http.HandleFunc("/ws", messageHandler)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func messageHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	connString := r.Header.Get("origin")
	client := &model.Client{Proc_id: "", ConnectionString: connString}
	controller.Model.PendingClients[connString] = struct{}{}
	controller.Model.ClientWs[connString] = conn
	go ping(conn)
	receiveLoop(conn, client)
}

func receiveLoop(ws *websocket.Conn, client *model.Client) {
	for {
		var data []byte
		ws.SetReadDeadline(time.Now().Add(1 * time.Second))
		messageType, data, err := ws.ReadMessage()
		if err != nil {
			log.Errorln(err)
			//todo: here we handle disconnections is every error a disconnection?
			controller.DisconnectClient(*client)
			break
		}
		if messageType == 2 {
			//ping
			continue
		}

		log.Debug("Handled data: ", string(data))
		var msg model.BaseMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Error("Error deserializing message:", err)
			continue
		}

		log.Infoln("Received message ", msg.GetMessageType().String(), " from ", client.Proc_id)
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
		case model.CLIENT_DISC:
			var discMsg model.ClientDisconnectMessage
			if err := json.Unmarshal(data, &discMsg); err != nil {
				log.Error("Error parsing ClientDisconnectMessage:", err)
			} else {
				controller.HandleClientDisconnectMessage(discMsg, client)
			}
		case model.DISC_ACK:
			var discAckMsg model.DisconnectAckMessage
			if err := json.Unmarshal(data, &discAckMsg); err != nil {
				log.Error("Error parsing ClientDisconnectAck:", err)
			} else {
				controller.HandleDisconnectAckMessage(discAckMsg, client)
			}
		default:
			log.Error("Unknown message type:", msg.GetMessageType())
		}
	}
}
