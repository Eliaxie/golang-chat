package controller

import (
	"encoding/json"
	"golang-chat/pkg/model"

	log "github.com/sirupsen/logrus"

	"github.com/google/uuid"
)

// Only send function not checking if client is in the model
func (c *Controller) SendInitMessage(message model.ConnectionInitMessage, client model.Client) {
	data, _ := json.Marshal(message)
	log.Debug(string(data))
	sendMessageSlave(client.Ws, data)
}

func (c *Controller) SendMessage(message model.Message, client model.Client) {
	if !controller.Model.Clients[client] {
		return
	}
	data, _ := json.Marshal(message)
	log.Debug(string(data))
	sendMessageSlave(client.Ws, data)
}

func (c *Controller) SendTextMessage(text string, client model.Client) {
	if !globModel.Clients[client] {
		return
	}
	msg := model.TextMessage{
		BaseMessage: model.BaseMessage{MessageType: model.TEXT},
		Content:     model.UniqueMessage{Text: text, UUID: uuid.New().String()}, Group: model.Group{Name: "default", Madeby: "default"}, VectorClock: model.VectorClock{}}
	data, _ := json.Marshal(msg)
	println("Sending message:", string(data))
	sendMessageSlave(client.Ws, data)
}

func (c *Controller) BroadcastMessage(text string) {
	for client := range c.Model.Clients {
		c.SendTextMessage(text, client)
	}
}

func (c *Controller) SendGroupMessage(text string, group model.Group) {
	textMessage := model.TextMessage{
		BaseMessage: model.BaseMessage{MessageType: model.TEXT},
		Content:     model.UniqueMessage{Text: text, UUID: uuid.New().String()}, Group: group, VectorClock: model.VectorClock{}}

	c.multicastMessage(textMessage, c.Model.Groups[group])
	c.tryAcceptMessage(textMessage, model.Client{})
}
