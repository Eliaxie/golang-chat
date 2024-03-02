package controller

import (
	"encoding/json"
	"fmt"
	"golang-chat/pkg/model"
	"log"
)

type Controller struct {
	Model *model.Model
}

func (c *Controller) HandleTextMessage(textMsg model.TextMessage, client model.Client) {
	fmt.Println("Received text message:", textMsg.Content)
	c.Model.GroupsBuffers[textMsg.Group] =
		append(c.Model.GroupsBuffers[textMsg.Group], model.PendingMessage{Content: textMsg.Content, Client: client, VectorClock: textMsg.VectorClock})

	c.tryAcceptMessage(textMsg, client)
}

func (c *Controller) SendMessage(message model.BaseMessage, client model.Client) {
	data, _ := json.Marshal(message)
	log.Print(data)
}

func (c *Controller) SendTextMessage(text string, client model.Client) {
		msg := model.TextMessage{Content: text, Group: model.Group{Name: "default", Madeby: "default"}, VectorClock: model.VectorClock{}}
		data, _ := json.Marshal(msg)
		println("Sending message:", string(data))
		sendMessageSlave(client.Ws, data)
}

func (c *Controller) BroadcastMessage(text string) {
	for client := range c.Model.Clients {
		c.SendTextMessage(text, *client)
	}
}

func (c *Controller) AddNewConnection(connection string) {
	c.addNewConnectionSlave(connection)
}

func (c *Controller) AddNewConnections(connection []string) {
	for _, conn := range connection {
		c.addNewConnectionSlave(conn)
	}
}

func (c *Controller) StartServer(port string) {
	InitWebServer(port, c.Model)
}

func (c *Controller) tryAcceptMessage(message model.TextMessage, client model.Client) {
	// logic to accept message using vector clocks
	// push to view

}
