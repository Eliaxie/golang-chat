package controller

import (
	"encoding/json"
	"fmt"
	"golang-chat/pkg/model"
)

type Controller struct {
	Model *model.Model
}

func (c *Controller) HandleTextMessage(textMsg model.TextMessage, client model.Client) {
	fmt.Println("Received text message:", textMsg.Content)
	c.Model.GroupsBuffers[textMsg.Group] =
		append(c.Model.GroupsBuffers[textMsg.Group], model.PendingMessage{Content: textMsg.Content, Client: client, VectorClock: textMsg.VectorClock})
}

func (c *Controller) SendMessage(text string, client model.Client) {
	msg := model.TextMessage{Content: text, Group: model.GroupName{Name: "default", Madeby: "default"}, VectorClock: model.VectorClock{}}
	data, _ := json.Marshal(msg)
	println("Sending message:", string(data))
	sendMessageSlave(client.Ws, data)
}

func (c *Controller) BroadcastMessage(text string) {
	for client := range c.Model.Clients {
		c.SendMessage(text, *client)
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
