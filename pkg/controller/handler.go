package controller

import (
	"golang-chat/pkg/model"
	"strings"
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
	// add original client and self to the list of clients
	_clients = append(_clients, *client)
	_clients = append(_clients, c.Model.Myself)
	// add connections for all clients other than self and sender
	for _, serializedClient := range groupCreateMsg.Clients {
		// remove self from the list of clients
		if c.Model.Myself.Proc_id == serializedClient.Proc_id {
			continue
		}
		if serializedClient.Proc_id == client.Proc_id {
			continue
		}

		clientConnection := c.AddNewConnection(serializedClient.HostName)
		clientConnection.Proc_id = serializedClient.Proc_id
		// new client with proc_id and hostName of groupClients
		_clients = append(_clients, clientConnection)

	}

	c.createGroup(groupCreateMsg.Group, groupCreateMsg.ConsistencyModel, _clients)
}

func (c *Controller) HandleTextMessage(textMsg model.TextMessage, client *model.Client) {
	c.tryAcceptMessage(textMsg, *client)
}

func (c *Controller) HandleMessageAck(messageAck model.MessageAck, client *model.Client) {
	c.Model.GroupsLocks[messageAck.Group].Lock()
	if c.Model.MessageAcks[messageAck.Group][messageAck.Reference] == nil {
		c.Model.MessageAcks[messageAck.Group][messageAck.Reference] = map[string]bool{}
	}
	c.Model.MessageAcks[messageAck.Group][messageAck.Reference][client.Proc_id] = true

	newMessage := false
	newMessage = c.tryAcceptTopGlobals(messageAck.Group)
	if newMessage {
		c.Notifier.Notify(messageAck.Group)
	}
	c.Model.GroupsLocks[messageAck.Group].Unlock()
}

func (c *Controller) HandleClientDisconnectMessage(clientDisconnectMsg model.ClientDisconnectMessage, client *model.Client) {
	// send back acknolwedgement
	c.SendMessage(model.DisconnectAckMessage{
		BaseMessage: model.BaseMessage{MessageType: model.DISC_ACK},
		Group:       clientDisconnectMsg.Group,
		ClientID:    clientDisconnectMsg.ClientID}, *client)
}

func (c *Controller) HandleDisconnectAckMessage(disconnectAckMsg model.DisconnectAckMessage, client *model.Client) {
	c.Model.DisconnectionLocks[disconnectAckMsg.Group].Lock()
	c.Model.DisconnectionAcks[disconnectAckMsg.Group][client.Proc_id] = struct{}{}
	c.Model.DisconnectionLocks[disconnectAckMsg.Group].Unlock()
}
