package controller

import (
	"crypto/sha1"
	"encoding/hex"
	"golang-chat/pkg/model"

	"github.com/denisbrodbeck/machineid"
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
	b, err := machineid.ID()
	if err != nil {
		b = "default"
	}
	hasher := sha1.New()
	hasher.Write([]byte(b))
	// hash b
	return hex.EncodeToString(hasher.Sum(nil))
}
