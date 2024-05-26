package controller

import (
	"golang-chat/pkg/maps"
	"golang-chat/pkg/model"
	"sort"
)

func (c *Controller) tryAcceptCasualMessages(group model.Group) bool {

	ownVectorClock := c.Model.GroupsVectorClocks[group]
	// for each pending message
	for pending_index, pendingMessage := range maps.Load(&c.Model.PendingMessages, group) {
		// for each client clock in the pending message
		for proc_id, clientsClock := range pendingMessage.VectorClock.Clock {

			if ownVectorClock.Clock[proc_id]+1 == clientsClock {
				everyOtherIsLower := true
				// ownVectorClock: 1 2 3 3
				// pendingMessage.VectorClock.Clock: 1 3 1 1
				for otherProc_id, ownClientsClock := range ownVectorClock.Clock {
					// proc_id = b, otherProc_id = c, otherClientsClock = 3, VectorClock.Clock[otherProc_id] = 1
					if otherProc_id != proc_id && ownClientsClock < pendingMessage.VectorClock.Clock[otherProc_id] {
						everyOtherIsLower = false
						break
					}
				}
				if everyOtherIsLower {
					//c.Model.StableMessages[group] = append(c.Model.StableMessages[group], model.StableMessage{Content: pendingMessage.Content, Client: pendingMessage.Client})
					maps.Store(&c.Model.StableMessages, group, append(maps.Load(&c.Model.StableMessages, group), model.StableMessage{Content: pendingMessage.Content, Client: pendingMessage.Client}))

					//c.Model.PendingMessages[group] = removeAtIndex(c.Model.PendingMessages[group], pending_index)
					maps.Store(&c.Model.PendingMessages, group, removeAtIndex(maps.Load(&c.Model.PendingMessages, group), pending_index))
					c.Model.GroupsVectorClocks[group].Clock[proc_id]++
					c.tryAcceptCasualMessages(group)
					return true
				}
			}
		}
	}
	return false
}

func removeAtIndex[T any](s []T, index int) []T {
	return append(s[:index], s[index+1:]...)
}

func (c *Controller) tryAcceptGlobalMessages(message model.TextMessage, client model.Client) bool {
	c.appendMsgToSortedPending(message, client)
	// increment own clock
	c.Model.GroupsVectorClocks[message.Group].Clock[c.Model.Myself.Proc_id]++

	// multicast message ack to all group members
	scalarClock := model.ScalarClockToProcId{Clock: message.VectorClock.Clock[client.Proc_id], Proc_id: client.Proc_id}

	activeClients := make([]model.Client, 0)
	for _, groupMember := range c.Model.Groups[message.Group] {
		if maps.Load(&c.Model.Clients, groupMember) {
			activeClients = append(activeClients, groupMember)
		}
	}

	c.multicastMessage(model.MessageAck{
		BaseMessage: model.BaseMessage{MessageType: model.MESSAGE_ACK},
		Group:       message.Group, Reference: scalarClock}, activeClients)

	// ensure the message ack map is initialized
	_group := maps.Load(&c.Model.MessageAcks, message.Group)
	_scalarClock := maps.Load(&_group, scalarClock)
	if _scalarClock == nil {
		//c.Model.MessageAcks[message.Group][scalarClock] = map[string]bool{}
		maps.Store(&_group, scalarClock, map[string]bool{})
	}
	// mark the message sender as acked
	//c.Model.MessageAcks[message.Group][scalarClock][client.Proc_id] = true
	maps.Store(&_scalarClock, client.Proc_id, true)
	return c.tryAcceptTopGlobals(message.Group)
}

func (c *Controller) appendMsgToSortedPending(message model.TextMessage, client model.Client) model.PendingMessage {
	ScalarClock := model.ScalarClockToProcId{Clock: message.VectorClock.Clock[client.Proc_id], Proc_id: client.Proc_id}
	pendingMessage := model.PendingMessage{Content: message.Content, Client: client,
		ScalarClock: ScalarClock}

	c.appendSortedPending(pendingMessage, message.Group)
	return pendingMessage
}

func (c *Controller) appendSortedPending(message model.PendingMessage, group model.Group) {
	// find the index where to insert the message
	pendingMessages := maps.Load(&c.Model.PendingMessages, group)
	index := sort.Search(len(pendingMessages), func(i int) bool {
		if pendingMessages[i].ScalarClock.Clock == message.ScalarClock.Clock {
			return pendingMessages[i].ScalarClock.Proc_id > message.ScalarClock.Proc_id
		}
		return pendingMessages[i].ScalarClock.Clock > message.ScalarClock.Clock
	})

	// insert the message
	//c.Model.PendingMessages[group] = append(pendingMessages[:index], append([]model.PendingMessage{message}, pendingMessages[index:]...)...)
	maps.Store(&c.Model.PendingMessages, group, append(pendingMessages[:index], append([]model.PendingMessage{message}, pendingMessages[index:]...)...)) // lol ok
}

func (c *Controller) tryAcceptTopGlobals(group model.Group) bool {
	hasNewMessages := false

	// inner functions that checks if all acks for a message have been received
	checkAcks := func(pendingMessage model.PendingMessage) bool {
		groupMembers := c.Model.Groups[group]
		for _, groupMember := range groupMembers {
			if maps.Load(&c.Model.Clients, groupMember) {
				// group member is active
				if groupMember.Proc_id == c.Model.Myself.Proc_id {
					continue
				}
				_group := maps.Load(&c.Model.MessageAcks, group)
				_scalarClock := maps.Load(&_group, pendingMessage.ScalarClock)
				if _, ok := maps.LoadAndCheck(&_scalarClock, groupMember.Proc_id); !ok {
					return false
				}
			}
		}
		return true
	}

	// try accepting pending messages until one is not accepted
	for _, pendingMessage := range maps.Load(&c.Model.PendingMessages, group) {
		isAccepted := checkAcks(pendingMessage)
		if !isAccepted {
			break
		}
		c.Model.StableMessages[group] = append(c.Model.StableMessages[group], model.StableMessage{Content: pendingMessage.Content, Client: pendingMessage.Client})
		maps.Store(&c.Model.StableMessages, group, append(maps.Load(&c.Model.StableMessages, group), model.StableMessage{Content: pendingMessage.Content, Client: pendingMessage.Client}))
		//c.Model.PendingMessages[group] = removeAtIndex(c.Model.PendingMessages[group], 0)
		maps.Store(&c.Model.PendingMessages, group, removeAtIndex(maps.Load(&c.Model.PendingMessages, group), 0))
		//c.Model.MessageAcks[group][pendingMessage.ScalarClock] = map[string]bool{}
		_group := maps.Load(&c.Model.MessageAcks, group)
		maps.Store(&_group, pendingMessage.ScalarClock, map[string]bool{})
		hasNewMessages = isAccepted || hasNewMessages
	}
	return hasNewMessages
}

func (c *Controller) tryAcceptFIFOMessages(message model.TextMessage, client model.Client) bool {

	newMessage := false
	// adds all pending messages to the stable buffer
	for _, pendingMessage := range maps.Load(&c.Model.PendingMessages, message.Group) {
		newMessage = true
		// c.Model.StableMessages[message.Group] = append(c.Model.StableMessages[message.Group], model.StableMessage{Content: pendingMessage.Content})
		maps.Store(&c.Model.StableMessages, message.Group, append(maps.Load(&c.Model.StableMessages, message.Group), model.StableMessage{Content: pendingMessage.Content}))
	}
	// empties the pending buffer
	//c.Model.PendingMessages[message.Group] = []model.PendingMessage{}
	maps.Store(&c.Model.PendingMessages, message.Group, []model.PendingMessage{})
	return newMessage
}
