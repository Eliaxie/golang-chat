package controller

import (
	"encoding/json"
	"golang-chat/pkg/model"
	"log"
	"time"
)

type Controller struct {
	Model *model.Model
}

func (c *Controller) AddNewConnection(connection string) model.Client {
	return c.addNewConnectionSlave("ws://localhost:"+c.Model.ServerPort+"/ws", connection)
}

func (c *Controller) AddNewConnections(connection []string) {
	for _, conn := range connection {
		c.addNewConnectionSlave("ws://localhost:"+c.Model.ServerPort+"/ws", conn)
	}
}

func (c *Controller) StartServer(port string) {
	c.Model.ServerPort = port
	InitWebServer(port, c)
}

func (c *Controller) tryAcceptMessage(message model.TextMessage, client model.Client) {
	_log, _ := json.Marshal(c.Model.GroupsBuffers[message.Group])
	log.Println("Buffer: ", string(_log))
}

func (c *Controller) CreateGroup(groupName string, clients []model.Client) model.Group {
	// Add the group to the model
	group := model.Group{Name: groupName, Madeby: c.Model.Proc_id}
	c.Model.Groups[group] = clients

	// Send the group create message to all the clients
	var serializedClients []model.SerializedClient
	for _, client := range clients {
		serializedClients = append(serializedClients,
			model.SerializedClient{Proc_id: client.Proc_id,
				HostName: client.Ws.RemoteAddr().String()})
	}
	c.multicastMessage(
		model.GroupCreateMessage{
			BaseMessage: model.BaseMessage{MessageType: model.GROUP_CREATE},
			Group:       model.Group{Name: groupName, Madeby: globModel.Proc_id},
			Clients:     serializedClients}, clients)
	return group
}

func (c *Controller) WaitForConnection(client model.Client) bool {
	for {
		// a client is no longer pending once it has been added to the clients list
		_, ok := c.Model.PendingClients[client]
		if !ok {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
}
