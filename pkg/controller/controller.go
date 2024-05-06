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

func (c *Controller) DisconnectClient(client model.Client) {

	for group, clients := range c.Model.Groups {
		for index, _client := range clients {
			if _client == client {
				// todo: think about group locks here
				switch c.Model.GroupsConsistency[group] {

				case model.GLOBAL:
					// remove the client from the active window
					delete(c.Model.ActiveWindows[group], client.Proc_id) //todo: check if this is correct before lock
					// todo: stop sending messages (locks?)
					clientsToNotify := make([]model.Client, len(clients))
					copy(clientsToNotify, clients)
					// remove all non active clients from the list of clients to notify
					for i, client := range clientsToNotify {
						if _, ok := c.Model.ActiveWindows[group][client.Proc_id]; !ok {
							clientsToNotify = append(clientsToNotify[:i], clientsToNotify[i+1:]...)
						}
					}
					//TODO: CHECK IF ACTIVE WINDOWS MAY BE A MAP TO CLIENTS INSTEAD OF STRINGS

					// send a message CLIENT_DISCONNECTED to all the clients in the group that the client has disconnected
					c.multicastMessage(model.ClientDisconnectMessage{BaseMessage: model.BaseMessage{MessageType: model.CLIENT_DISC}, ClientID: client.Proc_id}, clientsToNotify)

					// initiate the disconnection ack array
					acks := make(map[string]struct{})
					c.Model.DisconnectionAcks[group] = acks
					// wait for acks from all the clients
					for client, _ := range c.Model.ActiveWindows[group] {
						_, ok := acks[client]
						if !ok {
							// wait for the ack
							time.Sleep(100 * time.Millisecond)
							//todo: timeout?
						}
					}

					// if in majority partition, remove the client from ACTIVE_WINDOW
					// ...
					// resume sending messages (locks?)
				case model.CAUSAL:
					delete(controller.Model.Clients, client)
				default:
					delete(controller.Model.Clients, client)
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
		// put a copy of clients in the active window
		c.Model.ActiveWindows[group] = make(map[string]struct{})
		for _, client := range clients {
			c.Model.ActiveWindows[group][client.Proc_id] = struct{}{}
		}
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
