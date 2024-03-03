package controller

import (
	"encoding/json"
	"golang-chat/pkg/model"
	"log"
)

// Only send function not checking if client is in the model
func (c *Controller) SendInitMessage(message model.ConnectionInitMessage, client model.Client) {
	data, _ := json.Marshal(message)
	log.Print(string(data))
	sendMessageSlave(client.Ws, data)
}

func (c *Controller) SendMessage(message model.Message, client model.Client) {
	if !controller.Model.Clients[client] {
		return
	}
	data, _ := json.Marshal(message)
	log.Print(string(data))
	sendMessageSlave(client.Ws, data)
}

func (c *Controller) SendTextMessage(text string, client model.Client) {
	if !globModel.Clients[client] {
		return
	}
	msg := model.TextMessage{
		BaseMessage: model.BaseMessage{MessageType: model.TEXT},
		Content:     text, Group: model.Group{Name: "default", Madeby: "default"}, VectorClock: model.VectorClock{}}
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
	c.multicastMessage(model.TextMessage{
		BaseMessage: model.BaseMessage{MessageType: model.TEXT},
		Content:     text, Group: group, VectorClock: model.VectorClock{}}, c.Model.Groups[group])
}
