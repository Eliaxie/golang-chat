package controller

import (
	"golang-chat/pkg/model"
	"log"

	"golang.org/x/net/websocket"
)

func initializeClient(ws *websocket.Conn, client *model.Client) {
	// Send connection init message
	connInitMsg := model.ConnectionInitMessage{
		MessageType: model.CONN_INIT,
		ClientID:    client.Proc_id,
	}
	if err := websocket.JSON.Send(ws, connInitMsg); err != nil {
		log.Println("Error sending connection init message:", err)
	}

}
