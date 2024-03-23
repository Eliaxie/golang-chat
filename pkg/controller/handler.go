package controller

import (
	"golang-chat/pkg/model"
)

func (c *Controller) HandleConnectionInitMessage(connInitMsg model.ConnectionInitMessage, client model.Client) {
	client.Proc_id = connInitMsg.ClientID
	controller.Model.Clients[client] = true
	delete(controller.Model.PendingClients, client)

	// Send reply INIT Message with my clientID
	controller.SendMessage(model.ConnectionInitResponseMessage{
		BaseMessage: model.BaseMessage{MessageType: model.CONN_INIT_RESPONSE},
		ClientID:    globModel.Proc_id,
	}, client)
}

func (c *Controller) HandleConnectionInitResponseMessage(connInitRespMsg model.ConnectionInitResponseMessage, client model.Client) {
	client.Proc_id = connInitRespMsg.ClientID
	controller.Model.Clients[client] = true
	delete(controller.Model.PendingClients, client)
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
		// remove self from the list of clients
		if c.Model.Proc_id == client.Proc_id {
			continue
		}
		_clients = append(_clients, c.AddNewConnection(client.HostName))
	}
	// add original client to the list of clients
	_clients = append(_clients, client)

	c.Model.Groups[groupCreateMsg.Group] = _clients
	c.Model.GroupsConsistency[groupCreateMsg.Group] = groupCreateMsg.ConsistencyModel
	c.Model.PendingMessages[groupCreateMsg.Group] = []model.PendingMessage{}
	c.Model.StableMessages[groupCreateMsg.Group] = []model.StableMessages{}
	c.Model.GroupsVectorClocks[groupCreateMsg.Group] = model.VectorClock{Clock: map[string]int{}}
}

func (c *Controller) HandleTextMessage(textMsg model.TextMessage, client model.Client) {
	c.Model.PendingMessages[textMsg.Group] =
		append(c.Model.PendingMessages[textMsg.Group], model.PendingMessage{Content: textMsg.Content, Client: client, VectorClock: textMsg.VectorClock})

	c.tryAcceptMessage(textMsg, client)
}
