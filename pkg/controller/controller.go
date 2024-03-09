package controller

import (
	"encoding/json"
	"golang-chat/pkg/model"
	"golang-chat/pkg/notify"
	"time"

	log "github.com/sirupsen/logrus"
)

type Controller struct {
	Model    *model.Model
	Notifier *notify.Notifier
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

// Tries to accept the received message. Returns true if the buffer is empty, false otherwise
// If accepted the message is moved from the PendingBuffer to the StableBuffer
func (c *Controller) tryAcceptMessage(message model.TextMessage, client model.Client) bool {
	_logP, _ := json.Marshal(c.Model.PendingMessages[message.Group])
	_logS, _ := json.Marshal(c.Model.StableMessages[message.Group])
	log.Debug("Buffer Pending: ", string(_logP))
	log.Debug("Buffer Stable: ", string(_logS))
	pendingMessages := c.Model.PendingMessages[message.Group]
	stableMessages := c.Model.StableMessages[message.Group]
	callView := false
	switch c.Model.GroupsConsistency[message.Group] {
	case model.CAUSAL:
		callView = c.tryAcceptCasualMessages(&pendingMessages, &stableMessages, message, client)
	case model.GLOBAL:
		callView = c.tryAcceptGlobalMessages(&pendingMessages, &stableMessages, message, client)
	case model.LINEARIZABLE:
		callView = c.tryAcceptLinearizableMessages(&pendingMessages, &stableMessages, message, client)
	case model.FIFO:
		callView = c.tryAcceptFIFOMessages(&pendingMessages, &stableMessages, message, client)
	default:
		log.Panic("Unknown consistency model")
	}
	if callView {
		c.Notifier.Notify(message.Group)
	}
	_logP, _ = json.Marshal(c.Model.PendingMessages[message.Group])
	_logS, _ = json.Marshal(c.Model.StableMessages[message.Group])
	log.Debug("Buffer Pending: ", string(_logP))
	log.Debug("Buffer Stable: ", string(_logS))
	return false
}

func (c *Controller) CreateGroup(groupName string, consistencyModel model.ConsistencyModel, clients []model.Client) model.Group {
	// Add the group to the model
	group := model.Group{Name: groupName, Madeby: c.Model.Proc_id}
	c.Model.Groups[group] = clients
	c.Model.GroupsConsistency[group] = consistencyModel

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
