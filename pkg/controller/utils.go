package controller

import (
	"crypto/rand"
	"encoding/hex"
	"golang-chat/pkg/model"
)

func initializeClient(proc_id string, client *model.Client) {
	// Send connection init message
	connInitMsg := model.ConnectionInitMessage{
		BaseMessage: model.BaseMessage{MessageType: model.CONN_INIT},
		ClientID:    proc_id,
	}
	controller.SendInitMessage(connInitMsg, *client)

}

func (c *Controller) GenerateUniqueID() string {
	b := make([]byte, 16) // generate 128-bit (16-byte) ID
	rand.Read(b)
	return hex.EncodeToString(b)
}
