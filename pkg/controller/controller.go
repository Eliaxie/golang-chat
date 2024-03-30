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
	log.Debug("Buffer Pending: ", string(_logP))
	log.Debug("Buffer Stable: ", string(_logS))

	newMessage := true
	switch c.Model.GroupsConsistency[message.Group] {
	case model.CAUSAL:
		pendingMessage := model.PendingMessage{Content: message.Content, Client: client, VectorClock: message.VectorClock}
		c.Model.PendingMessages[message.Group] = append(c.Model.PendingMessages[message.Group], pendingMessage)
		newMessage = c.tryAcceptCasualMessages(message.Group)
	case model.GLOBAL:
		pendingMessage := model.PendingMessage{Content: message.Content, Client: client,
			ScalarClock: model.ScalarClockToProcId{Clock: message.VectorClock.Clock[client.Proc_id], Proc_id: client.Proc_id}}
		c.Model.PendingMessages[message.Group] =
			append(c.Model.PendingMessages[message.Group], pendingMessage)
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
	log.Debug("Buffer Pending: ", string(_logP))
	log.Debug("Buffer Stable: ", string(_logS))
	return false
}

func (c *Controller) CreateGroup(groupName string, consistencyModel model.ConsistencyModel, clients []model.Client) model.Group {
	// Add the group to the model
	group := model.Group{Name: groupName, Madeby: c.Model.Myself.Proc_id}
	clients = append(clients, c.Model.Myself)
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
	}
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
