package controller

import (
	"encoding/json"
	"golang-chat/pkg/maps"
	"golang-chat/pkg/model"

	log "github.com/sirupsen/logrus"

	"github.com/google/uuid"
)

func (c *Controller) SendMessage(message model.Message, client model.Client) {
	data, error := json.Marshal(message)
	if error != nil {
		log.Fatal("Error marshalling message: ", error)
	}
	c.Model.MessageExitBufferLock.Lock()
	c.Model.MessageExitBuffer[client] = append(c.Model.MessageExitBuffer[client], model.MessageWithType{MessageType: message.GetMessageType(), Message: data})
	c.Model.MessageExitBufferLock.Unlock()
	log.Infoln("Sending message " + message.GetMessageType().String() + " to " + client.ConnectionString)
	if maps.Load(&controller.Model.Clients, client) || message.GetMessageType() == model.CONN_RESTORE || message.GetMessageType() == model.CONN_INIT || message.GetMessageType() == model.CONN_INIT_RESPONSE {
		sendMessageSlave(maps.Load(&c.Model.ClientWs, client.ConnectionString), client, maps.Load(&c.Model.Clients, client))
	}
}

func (c *Controller) SendGroupMessage(text string, group model.Group) {
	c.Model.GroupsLocks[group].Lock()
	vectorClock := c.Model.GroupsVectorClocks[group]
	vectorClock.Clock[c.Model.Myself.Proc_id]++
	textMessage := model.TextMessage{
		BaseMessage: model.BaseMessage{MessageType: model.TEXT},
		Content:     model.UniqueMessage{Text: text, UUID: uuid.New().String()}, Group: group, VectorClock: vectorClock}

	if c.Model.GroupsConsistency[group] != model.GLOBAL {
		//c.Model.StableMessages[group] = append(c.Model.StableMessages[group], model.StableMessage{Content: textMessage.Content, Client: c.Model.Myself})
		maps.Store(&c.Model.StableMessages, group, append(maps.Load(&c.Model.StableMessages, group), model.StableMessage{Content: textMessage.Content, Client: c.Model.Myself}))
		c.Notifier.Notify(group)
	} else {
		c.appendMsgToSortedPending(textMessage, c.Model.Myself)
	}

	c.Model.GroupsLocks[group].Unlock()
	c.multicastMessage(textMessage, c.Model.Groups[group])
}
