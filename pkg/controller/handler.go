package controller

import (
	"golang-chat/pkg/model"
	"log"
)

func (c *Controller) HandleConnectionInitMessage(connInitMsg model.ConnectionInitMessage, client model.Client) {
	client.Proc_id = connInitMsg.ClientID
	controller.Model.Clients[client] = true
	delete(controller.Model.PendingClients, client)

	// Send reply INIT Message with my clientID
	controller.SendMessage(model.ConnectionInitResponseMessage{
		BaseMessage: model.BaseMessage{MessageType: model.CONN_INIT_RESPONSE},
		ClientID:    globModel.Name,
	}, client)
}

func (c *Controller) HandleConnectionInitResponseMessage(connInitRespMsg model.ConnectionInitResponseMessage, client model.Client) {
	delete(controller.Model.PendingClients, client)
	client.Proc_id = connInitRespMsg.ClientID
	controller.Model.Clients[client] = true
}

func (c *Controller) HandleConnectionRestoreMessage(connRestoreMsg model.ConnectionRestoreMessage, client model.Client) {
	panic("unimplemented")
}

func (c *Controller) HandleConnectionRestoreResponseMessage(connRestoreRespMsg model.ConnectionRestoreResponseMessage, client model.Client) {
	panic("unimplemented")
}

func (c *Controller) HandleSyncPeersMessage(syncPeersMsg model.SyncPeersMessage, client model.Client) {
	serializedClients := []model.SerializedClient{}
	for client, active := range controller.Model.Clients {
		if active {
			serializedClients = append(serializedClients, model.SerializedClient{Proc_id: client.Proc_id, HostName: client.Ws.RemoteAddr().String()})
		}
	}
	c.SendMessage(model.SyncPeersResponseMessage{
		BaseMessage: model.BaseMessage{MessageType: model.SYNC_PEERS_RESPONSE},
		Peers:       serializedClients,
	}, client)
}

func (c *Controller) HandleSyncPeersResponseMessage(syncPeersRespMsg model.SyncPeersResponseMessage, client model.Client) {
	for _, peer := range syncPeersRespMsg.Peers {
		controller.AddNewConnection(peer.HostName)
	}
}

func (c *Controller) HandleGroupCreateMessage(groupCreateMsg model.GroupCreateMessage, client model.Client) {
	var _clients []model.Client
	for _, client := range groupCreateMsg.Clients {
		_clients = append(_clients, c.AddNewConnection(client.HostName))
	}

	c.Model.Groups[groupCreateMsg.Group] = _clients
	c.Model.GroupsConsistency[groupCreateMsg.Group] = groupCreateMsg.ConsistencyModel
	c.Model.GroupsBuffers[groupCreateMsg.Group] = []model.PendingMessage{}
	c.Model.GroupsVectorClocks[groupCreateMsg.Group] = model.VectorClock{Clock: map[string]int{}}
}

func (c *Controller) HandleTextMessage(textMsg model.TextMessage, client model.Client) {
	log.Println("Received text message:", textMsg.Content)
	c.Model.GroupsBuffers[textMsg.Group] =
		append(c.Model.GroupsBuffers[textMsg.Group], model.PendingMessage{Content: textMsg.Content, Client: client, VectorClock: textMsg.VectorClock})

	c.tryAcceptMessage(textMsg, client)
}