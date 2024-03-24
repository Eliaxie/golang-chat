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
	sendMessageSlave(controller.Model.ClientWs[client.ConnectionString], data)
}

func (c *Controller) SendMessage(message model.Message, client model.Client) {
	if !controller.Model.Clients[client] {
		return
	}
	data, _ := json.Marshal(message)
	log.Debug(string(data))
	sendMessageSlave(c.Model.ClientWs[client.ConnectionString], data)
}

func (c *Controller) SendTextMessage(text string, client model.Client) {
	if !c.Model.Clients[client] {
		return
	}
	msg := model.TextMessage{
		BaseMessage: model.BaseMessage{MessageType: model.TEXT},
		Content:     model.UniqueMessage{Text: text, UUID: uuid.New().String()}, Group: model.Group{Name: "default", Madeby: "default"}, VectorClock: model.VectorClock{}}
	data, _ := json.Marshal(msg)
	println("Sending message:", string(data))
	sendMessageSlave(c.Model.ClientWs[client.ConnectionString], data)
}

func (c *Controller) BroadcastMessage(text string) {
	for client := range c.Model.Clients {
		c.SendTextMessage(text, client)
	}
}

func (c *Controller) SendGroupMessage(text string, group model.Group) {
	vectorClock := c.Model.GroupsVectorClocks[group]
	vectorClock.Clock[c.Model.Myself.Proc_id]++
	textMessage := model.TextMessage{
		BaseMessage: model.BaseMessage{MessageType: model.TEXT},
		Content:     model.UniqueMessage{Text: text, UUID: uuid.New().String()}, Group: group, VectorClock: vectorClock}

	c.multicastMessage(textMessage, c.Model.Groups[group])
	c.Model.StableMessages[group] = append(c.Model.StableMessages[group], model.StableMessages{Content: textMessage.Content})
}
