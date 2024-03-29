package controller

import (
	"debug/pe"
	"golang-chat/pkg/model"
	"sort"
	"strings"
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
	
	pendingMessages := c.Model.PendingMessages[message.Group]
	
	// order pending messages by vector clock and procID
	sort.Slice(pendingMessages, func(i, j int) bool {
		if pendingMessages[i].VectorClock.Clock[client.Proc_id] == pendingMessages[j].VectorClock.Clock[client.Proc_id] {
				return pendingMessages[i].ProcID < pendingMessages[j].ProcID
		}
		return pendingMessages[i].VectorClock < pendingMessages[j].VectorClock
})




  // if pending messages
	if (len(c.Model.PendingMessages[message.Group]) == 0) {
		c.Model.PendingMessages[message.Group] = append(c.Model.PendingMessages[message.Group], model.PendingMessage{Content: message.Content, Client: client, VectorClock: message.VectorClock})
	} 

	c.Model.PendingMessages[message.Group] = 
	append(c.Model.PendingMessages[message.Group], 
		model.PendingMessage{Content: message.Content, Client: client, VectorClock: message.VectorClock})

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
	c.Model.PendingMessages[message.Group] = []model.PendingMessage{}
	return newMessage
}
