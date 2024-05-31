package controller

import (
	"golang-chat/pkg/maps"
	"golang-chat/pkg/model"
	"golang-chat/pkg/notify"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
)

type Controller struct {
	Model    *model.Model
	Notifier *notify.Notifier
}

func (c *Controller) AddNewConnection(connection string) (model.Client, error) {
	resp, err := c.addNewConnectionSlave(c.Model.Myself.ConnectionString, connection, false)
	if err == nil {
		return *resp, nil
	}
	return model.Client{}, err
}

func (c *Controller) Reconnect(connection string) (model.Client, error) {
	resp, err := c.addNewConnectionSlave(c.Model.Myself.ConnectionString, connection, true)
	if err == nil {
		return *resp, nil
	}
	return model.Client{}, err
}

func (c *Controller) syncReconnectedClient(client model.Client, reconnection bool) {
	stables := make(map[model.Group][]model.StableMessage)
	pending := make(map[model.Group][]model.PendingMessage)
	// for group, clients := range c.Model.Groups {
	allGroups, allClients := maps.KeysValues(&c.Model.Groups)
	for index, group := range allGroups {
		clients := allClients[index]
		if slices.Contains(clients, client) {
			c.Model.GroupsLocks[group].Lock()
			stables[group] = c.Model.StableMessages[group]
			pending[group] = maps.Load(&c.Model.PendingMessages, group)
		}
	}
	log.Trace("Sending connection restore message to ", client.Proc_id)

	groups := make([]model.Group, len(stables))
	serializedStables := make([][]model.StableMessage, len(stables))
	serializedPendings := make([][]model.PendingMessage, len(stables))
	serializedVectorClocks := make([]model.VectorClock, len(stables))
	serializedClientsInGroups := make([][]model.SerializedClient, len(stables))
	consistencyModels := make([]model.ConsistencyModel, len(stables))
	i := 0
	for group := range stables {
		groups[i] = group
		consistencyModels[i] = maps.Load(&c.Model.GroupsConsistency, group)
		serializedStables[i] = stables[group]
		serializedPendings[i] = pending[group]
		serializedVectorClocks[i] = maps.Load(&c.Model.GroupsVectorClocks, group)
		if !reconnection {
			// for _, client := range c.Model.Groups[group] {
			for _, client := range maps.Load(&c.Model.Groups, group) {
				serializedClientsInGroups[i] = append(serializedClientsInGroups[i], model.SerializedClient{Proc_id: client.Proc_id, HostName: client.ConnectionString})
				// TODO if connectionString for remote client == localhost -> check if remoteclientString is set and send that
			}
		}
		i++
	}

	c.SendMessage(model.ConnectionRestoreMessage{
		BaseMessage:               model.BaseMessage{MessageType: model.CONN_RESTORE},
		StableMessages:            serializedStables,
		PendingMessages:           serializedPendings,
		ConsistencyModel:          consistencyModels,
		Groups:                    groups,
		SerializedClientsInGroups: serializedClientsInGroups,
		GroupsVectorClocks:        serializedVectorClocks,
	}, client)

	//controller.Model.Clients[client] = true
	maps.Store(&c.Model.Clients, client, true)
	allGroups, allClients = maps.KeysValues(&c.Model.Groups)
	for index, group := range allGroups {
		clients := allClients[index]
		if slices.Contains(clients, client) {
			c.Model.GroupsLocks[group].Unlock()
		}
	}
	c.Notifier.NotifyView("Connection restored to " + client.ConnectionString)
}

func (c *Controller) AddNewConnections(connection []string) {
	for _, conn := range connection {
		c.addNewConnectionSlave(c.Model.Myself.ConnectionString, conn, false)
	}
}

func (c *Controller) DisconnectClient(disconnectedClient model.Client) {
	c.Notifier.NotifyView("Lost connection to client: " + disconnectedClient.ConnectionString)
	defer func() {
		if maps.Load(&c.Model.Clients, disconnectedClient) {
			log.Debug("Starting retry connections for ", disconnectedClient.ConnectionString)
		}
		go c.StartRetryConnections(disconnectedClient)
	}()
	// actions to take regardless of the consistency model
	allGroups, allClients := maps.KeysValues(&c.Model.Groups)
	for index, group := range allGroups {
		clients := allClients[index]
		for _, _client := range clients {
			if _client == disconnectedClient {
				// todo: think about group locks here
				switch maps.Load(&c.Model.GroupsConsistency, group) {

				case model.GLOBAL:
					// if client is already marked as disconnected, do nothing
					if !maps.Load(&c.Model.Clients, disconnectedClient) {
						break
					}

					c.Model.DisconnectionLocks[group].Lock()
					//c.Model.Clients[disconnectedClient] = false
					maps.Store(&c.Model.Clients, disconnectedClient, false)
					c.Model.DisconnectionLocks[group].Unlock()

					// todo: stop sending messages (locks?) and modifing group data
					c.Model.GroupsLocks[group].Lock()
					clientsToNotify := make([]model.Client, 0)
					for _, activeClient := range maps.Load(&c.Model.Groups, group) {
						if maps.Load(&c.Model.Clients, activeClient) {
							clientsToNotify = append(clientsToNotify, activeClient)
						}
					}

					// initiate the disconnection ack array
					acks := make(map[string]struct{})
					// c.Model.DisconnectionAcks[group] = acks
					maps.Store(&c.Model.DisconnectionAcks, group, acks)
					c.Model.DisconnectionLocks[group] = &sync.Mutex{}

					// get all pending messages for group sent by disconnected client
					disconnectedPendings := []model.PendingMessage{}
					for _, pendingMessage := range c.Model.PendingMessages[group] {
						if pendingMessage.Client == disconnectedClient {
							disconnectedPendings = append(disconnectedPendings, pendingMessage)
						}
					}

					// send a message CLIENT_DISCONNECTED to all active clients
					c.multicastMessage(
						model.ClientDisconnectMessage{
							BaseMessage:     model.BaseMessage{MessageType: model.CLIENT_DISC},
							Group:           group,
							Client:          model.SerializedClient{Proc_id: disconnectedClient.Proc_id, HostName: disconnectedClient.ConnectionString},
							PendingMessages: disconnectedPendings,
						}, clientsToNotify)

					// wait for acks from all the clients
					for len(acks) < len(clientsToNotify) {
						for _, activeClient := range clientsToNotify {
							acknowledged := false
							inActiveWindow := true
							for !acknowledged && inActiveWindow {
								c.Model.DisconnectionLocks[group].Lock()
								_, acknowledged = acks[activeClient.Proc_id]
								//_, inActiveWindow = c.Model.Clients[activeClient]
								maps.Load(&c.Model.Clients, activeClient)
								c.Model.DisconnectionLocks[group].Unlock()
								time.Sleep(100 * time.Millisecond)
							}
							log.Debugln("exit ack loop for ", activeClient.Proc_id, " acknowledged: ", acknowledged, " inActiveWindow: ", inActiveWindow)
						}
					}

					// check if majority partitioned
					// if len(acks)+1 > (len(c.Model.Groups[group]))/2 {
					if len(acks)+1 > (len(maps.Load(&c.Model.Groups, group)))/2 {
						log.Infoln("Group ", group.Name, " majority partitioned after client ", disconnectedClient.Proc_id, " disconnected")
						// try to accept the messages with the new active window
						c.tryAcceptTopGlobals(group)
					}
					// resume sending messages (locks?)
					c.Model.GroupsLocks[group].Unlock()
				case model.CAUSAL:
					//controller.Model.Clients[disconnectedClient] = false
					maps.Store(&c.Model.Clients, disconnectedClient, false)
				default:
					//controller.Model.Clients[disconnectedClient] = false
					maps.Store(&c.Model.Clients, disconnectedClient, false)
				}

				break
			}
		}
	}

}

func (c *Controller) StartServer(port string, extIp string) {
	c.Model.ServerPort = port
	c.Model.Myself.ConnectionString = extIp
	InitWebServer(port, c)
}

func (c *Controller) StartRetryConnections(client model.Client) {
	for {
		time.Sleep(3000 * time.Millisecond)

		//connected := c.Model.Clients[client]
		connected := maps.Load(&c.Model.Clients, client)
		if connected {
			log.Info("Client ", client.Proc_id, " is already connected - Stop retrying")
			return
		}
		// we retry only if the client is not connected and the client is lexicographically smaller than the current client to avoid cycles
		if strings.Compare(c.Model.Myself.Proc_id, client.Proc_id) > 0 {
			log.Trace("Retrying connection to ", client.ConnectionString)
			client, err := c.Reconnect(client.ConnectionString)
			if err != nil {
				log.Trace("Failed to connect to ", client.ConnectionString)
			} else {
				log.Trace("Successfully reconnected to ", client.ConnectionString)
				c.Notifier.NotifyView("Successfully reconnected to "+client.ConnectionString, color.BgGreen)
				return
			}
		}
	}
}

func (c *Controller) StartRetryMessages() {
	for {
		c.Model.MessageExitBufferLock.Lock()
		clientsTemp := maps.Clone(&c.Model.MessageExitBuffer)
		c.Model.MessageExitBufferLock.Unlock()
		for client := range clientsTemp {
			if !maps.Load(&c.Model.Clients, client) {
				continue
			}

			c.Model.MessageExitBufferLock.Lock()
			oldLen := len(maps.Load(&c.Model.MessageExitBuffer, client))
			oldMsg := []byte{}
			if oldLen != 0 {
				oldMsg = maps.Load(&c.Model.MessageExitBuffer, client)[0].Message
			}
			c.Model.MessageExitBufferLock.Unlock()

			if oldLen == 0 {
				continue
			}

			time.Sleep(100 * time.Millisecond)
			retryNeeded := false

			c.Model.MessageExitBufferLock.Lock()
			newLen := len(maps.Load(&c.Model.MessageExitBuffer, client))
			if newLen >= oldLen {
				if slices.Compare(oldMsg, maps.Load(&c.Model.MessageExitBuffer, client)[0].Message) == 0 {
					log.Trace("Stale messages found in MessageExitBuffer, retrying... ", client.ConnectionString)
					retryNeeded = true
				}
			}
			c.Model.MessageExitBufferLock.Unlock()
			if retryNeeded {
				if !maps.Load(&c.Model.Clients, client) {
					continue
				}
				sendMessageSlave(maps.Load(&c.Model.ClientWs, client.ConnectionString), client, true)
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// Tries to accept the received message. Returns true if the buffer is empty, false otherwise
// If accepted the message is moved from the PendingBuffer to the StableBuffer
func (c *Controller) tryAcceptMessage(message model.TextMessage, client model.Client) bool {

	c.Model.GroupsLocks[message.Group].Lock()

	// _logP, _ := json.Marshal(c.Model.PendingMessages[message.Group])
	// _logS, _ := json.Marshal(c.Model.StableMessages[message.Group])
	// log.Debugln("Buffer Pending: ", string(_logP))
	// log.Debugln("Buffer Stable: ", string(_logS))

	newMessage := true
	switch maps.Load(&c.Model.GroupsConsistency, message.Group) {
	case model.CAUSAL:
		pendingMessage := model.PendingMessage{Content: message.Content, Client: client, VectorClock: message.VectorClock}
		c.Model.PendingMessages[message.Group] = append(c.Model.PendingMessages[message.Group], pendingMessage)
		newMessage = c.tryAcceptCasualMessages(message.Group)
	case model.GLOBAL:
		newMessage = c.tryAcceptGlobalMessages(message, client)
	case model.FIFO:
		newMessage = c.tryAcceptFIFOMessages(message, client)
	default:
		log.Panic("Unknown consistency model")
	}

	if newMessage {
		c.Notifier.Notify(message.Group)
	}
	c.Model.GroupsLocks[message.Group].Unlock()

	// _logP, _ = json.Marshal(c.Model.PendingMessages[message.Group])
	// _logS, _ = json.Marshal(c.Model.StableMessages[message.Group])
	// log.Debugln("Buffer Pending: ", string(_logP))
	// log.Debugln("Buffer Stable: ", string(_logS))
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

	// c.Model.Groups[group] = clients
	maps.Store(&c.Model.Groups, group, clients)
	// c.Model.GroupsConsistency[group] = consistencyModel
	maps.Store(&c.Model.GroupsConsistency, group, consistencyModel)
	c.Model.GroupsLocks[group] = &sync.Mutex{}
	maps.Store(&c.Model.GroupsVectorClocks, group, model.VectorClock{Clock: map[string]int{}})
	switch consistencyModel {
	case model.CAUSAL:
		for _, client := range clients {
			clock := maps.Load(&c.Model.GroupsVectorClocks, group)
			clock.Clock[client.Proc_id] = 0
			maps.Store(&c.Model.GroupsVectorClocks, group, clock)
		}
	case model.GLOBAL:
		// In GLOBAL consistency model, the vector clock is used to keep track of the scalar clock of the group
		clock := maps.Load(&c.Model.GroupsVectorClocks, group)
		clock.Clock[c.Model.Myself.Proc_id] = 0
		maps.Store(&c.Model.GroupsVectorClocks, group, clock)
		// Intialize the map for message acks for the group
		//c.Model.MessageAcks[group] = make(map[model.ScalarClockToProcId]map[string]bool)
		maps.Store(&c.Model.MessageAcks, group, make(map[model.ScalarClockToProcId]map[string]bool))
	}
	return group
}

func (c *Controller) WaitForConnection(client model.Client) bool {
	for {
		// a client is no longer pending once it has been added to the clients list
		_, ok := maps.LoadAndCheck(&c.Model.PendingClients, client.ConnectionString)
		if !ok {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
}
