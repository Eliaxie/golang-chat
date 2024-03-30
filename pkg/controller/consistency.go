package controller

import (
	"golang-chat/pkg/model"
	"sort"
)

func (c *Controller) tryAcceptCasualMessages(group model.Group) bool {

	ownVectorClock := c.Model.GroupsVectorClocks[group]
	// for each pending message
	for pending_index, pendingMessage := range c.Model.PendingMessages[group] {
		// for each client clock in the pending message
		for proc_id, clientsClock := range pendingMessage.VectorClock.Clock {

			if ownVectorClock.Clock[proc_id]+1 == clientsClock {
				everyOtherIsLower := true
				for otherProc_id, otherClientsClock := range ownVectorClock.Clock {
					if otherProc_id != proc_id && otherClientsClock > pendingMessage.VectorClock.Clock[proc_id] {
						everyOtherIsLower = false
						break
					}
				}
				if everyOtherIsLower {
					c.Model.StableMessages[group] = append(c.Model.StableMessages[group],
						model.StableMessages{Content: pendingMessage.Content})

					c.Model.PendingMessages[group] = removeAtIndex(c.Model.PendingMessages[group], pending_index)
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

func (c *Controller) acceptCasualMessage(message model.TextMessage, client model.Client) bool {

	return false
}

func (c *Controller) tryAcceptGlobalMessages(message model.TextMessage, client model.Client) bool {
	ScalarClock := model.ScalarClockToProcId{Clock: message.VectorClock.Clock[client.Proc_id], Proc_id: client.Proc_id}

	// multicast message ack to all group members
	c.multicastMessage(model.MessageAck{Group: message.Group, Reference: ScalarClock}, c.Model.Groups[message.Group])

	c.Model.MessageAcks[message.Group][ScalarClock][client.Proc_id] = true

	pendingMessage := model.PendingMessage{Content: message.Content, Client: client,
		ScalarClock: ScalarClock}

	// find the index where to insert the message
	pendingMessages := c.Model.PendingMessages[message.Group]
	index := sort.Search(len(pendingMessages), func(i int) bool {
		if pendingMessages[i].ScalarClock.Clock == pendingMessage.ScalarClock.Clock {
			return pendingMessages[i].ScalarClock.Proc_id > pendingMessage.ScalarClock.Proc_id
		}
		return pendingMessages[i].ScalarClock.Clock > message.VectorClock.Clock[client.Proc_id]
	})

	// insert the message
	c.Model.PendingMessages[message.Group] = append(pendingMessages[:index], append([]model.PendingMessage{pendingMessage}, pendingMessages[index:]...)...)

	return c.tryAcceptTopGlobals(message.Group)
}

func (c *Controller) tryAcceptTopGlobals(group model.Group) bool {
	hasNewMessages := false

	// inner functions that checks if all acks for a message have been received
	checkAcks := func(pendingMessage model.PendingMessage) bool {
		groupMembers := c.Model.Groups[group]
		if len(c.Model.MessageAcks[group][pendingMessage.ScalarClock]) == len(groupMembers) {
			// for each group member check if an ack has been received
			for _, groupMember := range groupMembers {
				if _, ok := c.Model.MessageAcks[group][pendingMessage.ScalarClock][groupMember.Proc_id]; !ok {
					return false
				}
			}
			return true
		}
		return false
	}

	// try accepting pending messages until one is not accepted
	for pending_index, pendingMessage := range c.Model.PendingMessages[group] {
		isAccepted := checkAcks(pendingMessage)
		if !isAccepted {
			break
		}
		c.Model.StableMessages[group] = append(c.Model.StableMessages[group],
			model.StableMessages{Content: pendingMessage.Content})
		c.Model.PendingMessages[group] = removeAtIndex(c.Model.PendingMessages[group], pending_index)
		hasNewMessages = isAccepted || hasNewMessages
	}
	return hasNewMessages
}

func (c *Controller) tryAcceptLinearizableMessages(message model.TextMessage, client model.Client) bool {

	return false
}

func (c *Controller) tryAcceptFIFOMessages(message model.TextMessage, client model.Client) bool {

	newMessage := false
	// adds all pending messages to the stable buffer
	for _, pendingMessage := range c.Model.PendingMessages[message.Group] {
		newMessage = true
		c.Model.StableMessages[message.Group] = append(c.Model.StableMessages[message.Group],
			model.StableMessages{Content: pendingMessage.Content})

	}
	// empties the pending buffer
	c.Model.PendingMessages[message.Group] = []model.PendingMessage{}
	return newMessage
}
