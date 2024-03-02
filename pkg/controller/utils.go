package controller

import (
	"golang-chat/pkg/model"
)

func initializeClient(client *model.Client) {
	// Send connection init message
	connInitMsg := model.ConnectionInitMessage{
		BaseMessage: model.BaseMessage{MessageType: model.CONN_INIT},
		ClientID:    client.Proc_id,
	}
	controller.SendInitMessage(connInitMsg, *client)

}
