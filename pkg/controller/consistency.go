package controller

import (
	"golang-chat/pkg/model"
)

func (c *Controller) tryAcceptCasualMessages(group model.Group) bool {

	ownVectorClown := c.Model.GroupsVectorClocks[group]
	for _, pendingMessage := range c.Model.PendingMessages[group] {
		for proc_id, clientsClock := range pendingMessage.VectorClock.Clock {
			if ownVectorClown.Clock[proc_id] == clientsClock+1 {
				everyOtherIsLower := true
				for otherProc_id, otherClientsClock := range ownVectorClown.Clock {
					if otherProc_id != proc_id && otherClientsClock <= clientsClock {
						everyOtherIsLower = false
						break
					}
				}
				if everyOtherIsLower {
					c.Model.StableMessages[group] = append(c.Model.StableMessages[group],
						model.StableMessages{Content: pendingMessage.Content})

					newPendingMessages := []model.PendingMessage{}
					for _, otherPendingMessage := range c.Model.PendingMessages[group] {
						if otherPendingMessage.Content.UUID != pendingMessage.Content.UUID {
							newPendingMessages = append(newPendingMessages, otherPendingMessage)
						}
					}
					c.Model.PendingMessages[group] = newPendingMessages
					c.Model.GroupsVectorClocks[group].Clock[proc_id]++
					c.tryAcceptCasualMessages(group)
					return true
				}
			}
		}
	}
	return false
}

func (c *Controller) acceptCasualMessage(message model.TextMessage, client model.Client) bool {

	return false
}

func (c *Controller) tryAcceptGlobalMessages(message model.TextMessage, client model.Client) bool {

	return false
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
	c.Model.PendingMessages[message.Group] = nil // setting to nil doesn't mean it needs to be reinitialized
	return newMessage
}
