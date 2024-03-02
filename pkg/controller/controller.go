package controller

import (
	"encoding/json"
	"golang-chat/pkg/model"
	"log"
)

type Controller struct {
	Model *model.Model
}

func (c *Controller) AddNewConnection(connection string) model.Client {
	return c.addNewConnectionSlave(connection)
}

func (c *Controller) AddNewConnections(connection []string) {
	for _, conn := range connection {
		c.addNewConnectionSlave(conn)
	}
}

func (c *Controller) StartServer(port string) {
	InitWebServer(port, c)
}

func (c *Controller) tryAcceptMessage(message model.TextMessage, client model.Client) {
	_log, _ := json.Marshal(c.Model.GroupsBuffers[message.Group])
	log.Println("Buffer: ", string(_log))
}

func (c *Controller) CreateGroup(groupName string, clients []model.Client) {
	var serializedClients []model.SerializedClient
	for _, client := range clients {
		serializedClients = append(serializedClients, model.SerializedClient{Proc_id: client.Proc_id, HostName: client.Ws.RemoteAddr().String()})
	}

	c.multicastMessage(model.GroupCreateMessage{Group: model.Group{Name: groupName, Madeby: globModel.Name}, Clients: serializedClients}, clients)
}
