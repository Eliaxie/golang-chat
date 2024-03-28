package controller

import (
	"golang-chat/pkg/model"
	"strings"
	"sync"
)

func (c *Controller) HandleConnectionInitMessage(connInitMsg model.ConnectionInitMessage, client *model.Client) {

	oldConnectionString := client.ConnectionString
	client.Proc_id = connInitMsg.ClientID
	client.ConnectionString = "ws://" + strings.Split(controller.Model.ClientWs[client.ConnectionString].Request().Host, ":")[0] + ":" + connInitMsg.ServerPort + "/ws"
	controller.Model.ClientWs[client.ConnectionString] = controller.Model.ClientWs[oldConnectionString]
	delete(controller.Model.ClientWs, oldConnectionString)
	delete(controller.Model.PendingClients, oldConnectionString)

	controller.Model.Clients[*client] = true
	// Send reply INIT Message with my clientID
	controller.SendMessage(model.ConnectionInitResponseMessage{
		BaseMessage: model.BaseMessage{MessageType: model.CONN_INIT_RESPONSE},
		ClientID:    c.Model.Myself.Proc_id,
	}, *client)
}

func (c *Controller) HandleConnectionInitResponseMessage(connInitRespMsg model.ConnectionInitResponseMessage, client *model.Client) {
	delete(controller.Model.PendingClients, client.ConnectionString)
	client.Proc_id = connInitRespMsg.ClientID
	controller.Model.Clients[*client] = true
}

func (c *Controller) HandleConnectionRestoreMessage(connRestoreMsg model.ConnectionRestoreMessage, client *model.Client) {
	panic("unimplemented")
}

func (c *Controller) HandleConnectionRestoreResponseMessage(connRestoreRespMsg model.ConnectionRestoreResponseMessage, client *model.Client) {
	panic("unimplemented")
}

func (c *Controller) HandleSyncPeersMessage(syncPeersMsg model.SyncPeersMessage, client *model.Client) {
	serializedClients := []model.SerializedClient{}
	for client, active := range controller.Model.Clients {
		if active {
			serializedClients = append(serializedClients, model.SerializedClient{Proc_id: client.Proc_id, HostName: client.ConnectionString})
		}
	}
	c.SendMessage(model.SyncPeersResponseMessage{
		BaseMessage: model.BaseMessage{MessageType: model.SYNC_PEERS_RESPONSE},
		Peers:       serializedClients,
	}, *client)
}

func (c *Controller) HandleSyncPeersResponseMessage(syncPeersRespMsg model.SyncPeersResponseMessage, client *model.Client) {
	for _, peer := range syncPeersRespMsg.Peers {
		controller.AddNewConnection(peer.HostName)
	}
}

func (c *Controller) HandleGroupCreateMessage(groupCreateMsg model.GroupCreateMessage, client *model.Client) {
	var _clients []model.Client
	for _, serializedClient := range groupCreateMsg.Clients {
		// remove self from the list of clients
		if c.Model.Myself.Proc_id == serializedClient.Proc_id {
			continue
		}
		if serializedClient.Proc_id == client.Proc_id {
			continue
		}

		client := c.AddNewConnection(serializedClient.HostName)
		client.Proc_id = serializedClient.Proc_id
		// new client with proc_id and hostName of groupClients
		_clients = append(_clients, client)

	}
	// add original client to the list of clients
	_clients = append(_clients, *client)

	c.Model.Groups[groupCreateMsg.Group] = _clients
	c.Model.GroupsLocks[groupCreateMsg.Group] = &sync.Mutex{}
	c.Model.PendingMessages[groupCreateMsg.Group] = []model.PendingMessage{}
	c.Model.StableMessages[groupCreateMsg.Group] = []model.StableMessages{}
	
	c.Model.GroupsConsistency[groupCreateMsg.Group] = groupCreateMsg.ConsistencyModel
	c.Model.GroupsVectorClocks[groupCreateMsg.Group] = model.VectorClock{Clock: map[string]int{}}
	switch groupCreateMsg.ConsistencyModel {
	case model.CAUSAL:
		for _, client := range _clients {
			c.Model.GroupsVectorClocks[groupCreateMsg.Group].Clock[client.Proc_id] = 0
		}
	case model.GLOBAL:
		// In GLOBAL consistency model, the vector clock is used to keep track of the scalar clock of the group
		c.Model.GroupsVectorClocks[groupCreateMsg.Group].Clock[c.Model.Myself.Proc_id] = 0
	}
}

func (c *Controller) HandleTextMessage(textMsg model.TextMessage, client *model.Client) {
	c.tryAcceptMessage(textMsg, *client)
}
