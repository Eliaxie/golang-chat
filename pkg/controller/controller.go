package controller

import (
	"encoding/json"
	"golang-chat/pkg/model"
	"golang-chat/pkg/notify"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type Controller struct {
	Model    *model.Model
	Notifier *notify.Notifier
}

func (c *Controller) AddNewConnection(connection string) model.Client {
	return *c.addNewConnectionSlave("ws://localhost:"+c.Model.ServerPort+"/ws", connection)
}

func (c *Controller) AddNewConnections(connection []string) {
	for _, conn := range connection {
		c.addNewConnectionSlave("ws://localhost:"+c.Model.ServerPort+"/ws", conn)
	}
}

func (c *Controller) DisconnectClient(disconnectedClient model.Client) {
	log.Infoln("Lost connection to client: ", disconnectedClient.ConnectionString)

	// actions to take regardless of the consistency model

	for group, clients := range c.Model.Groups {
		for _, _client := range clients {
			if _client == disconnectedClient {
				// todo: think about group locks here
				switch c.Model.GroupsConsistency[group] {

				case model.GLOBAL:
					// if client is already marked as disconnected, do nothing
					if !c.Model.Clients[disconnectedClient] {
						break
					}

					c.Model.DisconnectionLocks[group].Lock()
					c.Model.Clients[disconnectedClient] = false
					c.Model.DisconnectionLocks[group].Unlock()

					// todo: stop sending messages (locks?) and modifing group data
					c.Model.GroupsLocks[group].Lock()
					clientsToNotify := make([]model.Client, 0)
					for _, activeClient := range c.Model.Groups[group] {
						if c.Model.Clients[activeClient] {
							clientsToNotify = append(clientsToNotify, activeClient)
						}
					}

					// initiate the disconnection ack array
					acks := make(map[string]struct{})
					c.Model.DisconnectionAcks[group] = acks
					c.Model.DisconnectionLocks[group] = &sync.Mutex{}

					// send a message CLIENT_DISCONNECTED to all active clients
					c.multicastMessage(model.ClientDisconnectMessage{BaseMessage: model.BaseMessage{MessageType: model.CLIENT_DISC},
						Group: group, ClientID: disconnectedClient.Proc_id}, clientsToNotify)

					// wait for acks from all the clients
					for len(acks) < len(clientsToNotify) {
						for _, activeClient := range clientsToNotify {
							acknowledged := false
							inActiveWindow := true
							for !acknowledged && inActiveWindow {
								c.Model.DisconnectionLocks[group].Lock()
								_, acknowledged = acks[activeClient.Proc_id]
								_, inActiveWindow = c.Model.Clients[activeClient]
								c.Model.DisconnectionLocks[group].Unlock()
								time.Sleep(100 * time.Millisecond)
							}
							log.Debugln("exit ack loop for ", activeClient.Proc_id, " acknowledged: ", acknowledged, " inActiveWindow: ", inActiveWindow)
						}
					}

					// check if majority partitioned
					if len(acks)+1 > (len(c.Model.Groups[group]))/2 {
						log.Infoln("Group ", group.Name, " majority partitioned after client ", disconnectedClient.Proc_id, " disconnected")
						// try to accept the messages with the new active window
						c.tryAcceptTopGlobals(group)
					}
					// resume sending messages (locks?)
					c.Model.GroupsLocks[group].Unlock()
				case model.CAUSAL:
					delete(controller.Model.Clients, disconnectedClient)
				default:
					delete(controller.Model.Clients, disconnectedClient)
				}

				break
			}
		}
	}

}

func (c *Controller) StartServer(port string) {
	c.Model.ServerPort = port
	InitWebServer(port, c)
}

// Tries to accept the received message. Returns true if the buffer is empty, false otherwise
// If accepted the message is moved from the PendingBuffer to the StableBuffer
func (c *Controller) tryAcceptMessage(message model.TextMessage, client model.Client) bool {

	c.Model.GroupsLocks[message.Group].Lock()

	_logP, _ := json.Marshal(c.Model.PendingMessages[message.Group])
	_logS, _ := json.Marshal(c.Model.StableMessages[message.Group])
	log.Debugln("Buffer Pending: ", string(_logP))
	log.Debugln("Buffer Stable: ", string(_logS))

	newMessage := true
	switch c.Model.GroupsConsistency[message.Group] {
	case model.CAUSAL:
		pendingMessage := model.PendingMessage{Content: message.Content, Client: client, VectorClock: message.VectorClock}
		c.Model.PendingMessages[message.Group] = append(c.Model.PendingMessages[message.Group], pendingMessage)
		newMessage = c.tryAcceptCasualMessages(message.Group)
	case model.GLOBAL:
		newMessage = c.tryAcceptGlobalMessages(message, client)
	case model.LINEARIZABLE:
		newMessage = c.tryAcceptLinearizableMessages(message, client)
	case model.FIFO:
		newMessage = c.tryAcceptFIFOMessages(message, client)
	default:
		log.Panic("Unknown consistency model")
	}

	if newMessage {
		c.Notifier.Notify(message.Group)
	}
	c.Model.GroupsLocks[message.Group].Unlock()

	_logP, _ = json.Marshal(c.Model.PendingMessages[message.Group])
	_logS, _ = json.Marshal(c.Model.StableMessages[message.Group])
	log.Debugln("Buffer Pending: ", string(_logP))
	log.Debugln("Buffer Stable: ", string(_logS))
	return false
}

// Creates a group and sends the group create message to all the involved clients
func (c *Controller) CreateGroup(groupName string, consistencyModel model.ConsistencyModel, clients []model.Client) model.Group {

	clients = append(clients, c.Model.Myself)
	group := c.createGroup(model.Group{Name: groupName, Madeby: c.Model.Myself.Proc_id}, consistencyModel, clients)

	// Send the group create message to all the clients
	var serializedClients []model.SerializedClient
	for _, client := range clients {
		serializedClients = append(serializedClients,
			model.SerializedClient{Proc_id: client.Proc_id,
				HostName: client.ConnectionString})
	}

	c.multicastMessage(
		model.GroupCreateMessage{
			BaseMessage:      model.BaseMessage{MessageType: model.GROUP_CREATE},
			ConsistencyModel: consistencyModel,
			Group:            model.Group{Name: groupName, Madeby: c.Model.Myself.Proc_id},
			Clients:          serializedClients}, clients)
	return group
}

func (c *Controller) createGroup(group model.Group, consistencyModel model.ConsistencyModel, clients []model.Client) model.Group {

	c.Model.Groups[group] = clients
	c.Model.GroupsConsistency[group] = consistencyModel
	c.Model.GroupsLocks[group] = &sync.Mutex{}
	c.Model.GroupsVectorClocks[group] = model.VectorClock{Clock: map[string]int{}}
	switch consistencyModel {
	case model.CAUSAL:
		for _, client := range clients {
			c.Model.GroupsVectorClocks[group].Clock[client.Proc_id] = 0
		}
	case model.GLOBAL:
		// In GLOBAL consistency model, the vector clock is used to keep track of the scalar clock of the group
		c.Model.GroupsVectorClocks[group].Clock[c.Model.Myself.Proc_id] = 0
		// Intialize the map for message acks for the group
		c.Model.MessageAcks[group] = make(map[model.ScalarClockToProcId]map[string]bool)
	}
	return group
}

func (c *Controller) WaitForConnection(client model.Client) bool {
	for {
		// a client is no longer pending once it has been added to the clients list
		_, ok := c.Model.PendingClients[client.ConnectionString]
		if !ok {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
}
