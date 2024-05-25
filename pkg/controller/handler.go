package controller

import (
	"golang-chat/pkg/model"
	"slices"
	"sync"

	log "github.com/sirupsen/logrus"
)

var clientReconnectionSynchronizationLock sync.Mutex

// map to track clients that are already connecting
var handleConnectionLock sync.Mutex

func (c *Controller) HandleConnectionInitMessage(connInitMsg model.ConnectionInitMessage, client *model.Client) {
	handleConnectionLock.Lock()
	defer handleConnectionLock.Unlock()

	var connectionFlow = model.FirstConnection

	// check if a client is reconnecting
	reconnection := false
	oldConnectionString := client.ConnectionString
	for _client := range c.Model.Clients {
		if _client.Proc_id == connInitMsg.ClientID {
			*client = _client
			reconnection = true
			break
		}
	}

	if reconnection && !connInitMsg.Reconnection {
		connectionFlow = model.ReconnectionPeerCrashed
	} else if reconnection && connInitMsg.Reconnection {
		connectionFlow = model.ReconnectionNetwork
	} else if !reconnection && connInitMsg.Reconnection {
		connectionFlow = model.ReconnectionSelfCrashed
	} else if !reconnection && !connInitMsg.Reconnection {
		connectionFlow = model.FirstConnection
	}

	// check if someone is trying to reconnect to me but I don't know him. I want to be the one who reconnects.
	// Also checks if the client is already connected
	if connectionFlow == model.ReconnectionSelfCrashed || c.Model.Clients[*client] {
		controller.SendMessage(model.ConnectionInitResponseMessage{
			BaseMessage: model.BaseMessage{MessageType: model.CONN_INIT_RESPONSE},
			ClientID:    c.Model.Myself.Proc_id,
			Refused:     true,
		}, *client)
		return
	}

	log.Debug("Connecting client: ", connInitMsg.ClientID, " ", client.ConnectionString, " Reconnection: ", reconnection, " connectionInit.Reconnection: ", connInitMsg.Reconnection)
	if connectionFlow == model.FirstConnection {
		client.Proc_id = connInitMsg.ClientID
		client.ConnectionString = connInitMsg.ServerIp
		c.Model.MessageExitBuffer[*client] = make([][]byte, 0)
	}
	controller.Model.ClientWs[client.ConnectionString] = controller.Model.ClientWs[oldConnectionString]
	delete(controller.Model.PendingClients, oldConnectionString)
	if connectionFlow == model.FirstConnection {
		controller.Model.Clients[*client] = true
	}

	// Send reply INIT Message with my clientID
	controller.SendMessage(model.ConnectionInitResponseMessage{
		BaseMessage: model.BaseMessage{MessageType: model.CONN_INIT_RESPONSE},
		ClientID:    c.Model.Myself.Proc_id,
	}, *client)

	// If the client hasn't crashed I don't need to sync it, I just set it as active
	if connectionFlow == model.ReconnectionPeerCrashed {
		c.syncReconnectedClient(*client, connInitMsg.Reconnection)
	} else if connectionFlow == model.ReconnectionNetwork {
		c.Model.Clients[*client] = true
	}

}

func (c *Controller) HandleConnectionInitResponseMessage(connInitRespMsg model.ConnectionInitResponseMessage, client *model.Client) {
	if connInitRespMsg.Refused {
		log.Debug("Connection refused")
		delete(controller.Model.PendingClients, client.ConnectionString)
		c.Model.ClientWs[client.ConnectionString].Close()
		return
	}
	log.Debug("Connection accepted by ", client)
	delete(controller.Model.PendingClients, client.ConnectionString)
	client.Proc_id = connInitRespMsg.ClientID
	controller.Model.Clients[*client] = true
}

func (c *Controller) HandleConnectionRestoreMessage(connRestoreMsg model.ConnectionRestoreMessage, client *model.Client) {
	clientReconnectionSynchronizationLock.Lock()
	for i, group := range connRestoreMsg.Groups {
		switch connRestoreMsg.ConsistencyModel[i] {
		case model.CAUSAL:
			if c.Model.Groups[group] == nil {
				groupClients := make([]model.Client, 0)
				for _, _remoteClient := range connRestoreMsg.SerializedClientsInGroups[i] {
					clientToAdd := model.Client{Proc_id: _remoteClient.Proc_id, ConnectionString: _remoteClient.HostName}

					if _remoteClient.Proc_id == c.Model.Myself.Proc_id {
						clientToAdd = model.Client{Proc_id: _remoteClient.Proc_id, ConnectionString: c.Model.Myself.ConnectionString}
					} else {
						for localClient, active := range c.Model.Clients {
							if localClient.Proc_id == _remoteClient.Proc_id {
								if !active {
									c.AddNewConnection(_remoteClient.HostName)
								}
								clientToAdd = localClient
							}
						}
					}
					groupClients = append(groupClients, clientToAdd)
				}
				c.createGroup(group, model.CAUSAL, groupClients)
			}
			for _, message := range connRestoreMsg.StableMessages[i] {
				if !slices.Contains(c.Model.StableMessages[group], message) {
					c.Model.StableMessages[group] = append(c.Model.StableMessages[group], message)
				}
			}
			for proc := range c.Model.GroupsVectorClocks[group].Clock {
				if connRestoreMsg.GroupsVectorClocks[i].Clock[proc] > c.Model.GroupsVectorClocks[group].Clock[proc] {
					c.Model.GroupsVectorClocks[group].Clock[proc] = connRestoreMsg.GroupsVectorClocks[i].Clock[proc]
				}
			}
			for _, message := range connRestoreMsg.PendingMessages[i] {
				alreadyInPending := false
				for _, pending := range c.Model.PendingMessages[group] {
					if pending.Content.UUID == message.Content.UUID {
						alreadyInPending = true
						break
					}
				}
				if !alreadyInPending {
					c.appendSortedPending(message, group)
				}
			}
			c.tryAcceptCasualMessages(group)
		case model.GLOBAL:
			// not implemented
		default:
		}
	}

	clientReconnectionSynchronizationLock.Unlock()

	c.SendMessage(model.ConnectionRestoreResponseMessage{
		BaseMessage: model.BaseMessage{MessageType: model.CONN_RESTORE_RESPONSE},
	}, *client)

}

func (c *Controller) HandleConnectionRestoreResponseMessage(connRestoreRespMsg model.ConnectionRestoreResponseMessage, client *model.Client) {
	// empty for now
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
		if c.Model.Myself.Proc_id < serializedClient.Proc_id {
			_, err := c.AddNewConnection(serializedClient.HostName)
			if err != nil {
				log.Errorln("Error adding new connection", err)
				continue
			}
		}

		// new client with proc_id and hostName of groupClients
		_clients = append(_clients, model.Client{Proc_id: serializedClient.Proc_id, ConnectionString: serializedClient.HostName})

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

	disconnectedClient := model.Client{Proc_id: clientDisconnectMsg.Client.Proc_id, ConnectionString: clientDisconnectMsg.Client.HostName}

	// remove client from active window
	c.Model.DisconnectionLocks[clientDisconnectMsg.Group].Lock()
	c.Model.Clients[disconnectedClient] = false
	c.Model.DisconnectionLocks[clientDisconnectMsg.Group].Unlock()

	// add all new pending messages
	c.Model.GroupsLocks[clientDisconnectMsg.Group].Lock()
	var pendingsToKeep []model.PendingMessage
	intersection := make(map[string]struct{})
	for _, message := range c.Model.PendingMessages[clientDisconnectMsg.Group] {
		// if not the client that disconnected
		if message.Client.Proc_id != disconnectedClient.Proc_id {
			pendingsToKeep = append(pendingsToKeep, message)
			continue
		}
		// if the client that disconnected check if the message is in the newMessages
		for _, newPending := range clientDisconnectMsg.PendingMessages {
			if newPending.Content.UUID == message.Content.UUID {
				pendingsToKeep = append(pendingsToKeep, message)
				intersection[newPending.Content.UUID] = struct{}{}
				break
			}
		}
	}
	c.Model.PendingMessages[clientDisconnectMsg.Group] = pendingsToKeep

	for _, newPending := range clientDisconnectMsg.PendingMessages {
		// if pending already in intersection do nothing ( message was already received)
		if _, found := intersection[newPending.Content.UUID]; found {
			continue
		}
		c.appendSortedPending(newPending, clientDisconnectMsg.Group)

		// increment own clock
		c.Model.GroupsVectorClocks[clientDisconnectMsg.Group].Clock[c.Model.Myself.Proc_id]++

		// send acks if the message is new
		activeClients := make([]model.Client, 0)
		for _, groupMember := range c.Model.Groups[clientDisconnectMsg.Group] {
			if c.Model.Clients[groupMember] {
				activeClients = append(activeClients, groupMember)
			}
		}

		c.multicastMessage(model.MessageAck{
			BaseMessage: model.BaseMessage{MessageType: model.MESSAGE_ACK},
			Group:       clientDisconnectMsg.Group, Reference: newPending.ScalarClock}, activeClients)

		// ensure the message ack map is initialized
		if c.Model.MessageAcks[clientDisconnectMsg.Group][newPending.ScalarClock] == nil {
			c.Model.MessageAcks[clientDisconnectMsg.Group][newPending.ScalarClock] = map[string]bool{}
		}
		// mark the message sender as acked
		c.Model.MessageAcks[clientDisconnectMsg.Group][newPending.ScalarClock][client.Proc_id] = true
		c.tryAcceptTopGlobals(clientDisconnectMsg.Group)
	}
	c.Model.GroupsLocks[clientDisconnectMsg.Group].Unlock()
	// send back acknolwedgement
	c.SendMessage(model.DisconnectAckMessage{
		BaseMessage: model.BaseMessage{MessageType: model.DISC_ACK},
		Group:       clientDisconnectMsg.Group,
		ClientID:    disconnectedClient.Proc_id}, *client)
}

func (c *Controller) HandleDisconnectAckMessage(disconnectAckMsg model.DisconnectAckMessage, client *model.Client) {
	c.Model.DisconnectionLocks[disconnectAckMsg.Group].Lock()
	c.Model.DisconnectionAcks[disconnectAckMsg.Group][client.Proc_id] = struct{}{}
	c.Model.DisconnectionLocks[disconnectAckMsg.Group].Unlock()
}
