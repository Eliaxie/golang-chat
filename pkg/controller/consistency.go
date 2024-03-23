package controller

import (
	"golang-chat/pkg/model"
)

func (c *Controller) tryAcceptCasualMessages(message model.TextMessage, client model.Client) bool {

	// check vector clock
	// see if that vector clock UNLOCKS which message

	// loop on every pending message
	// to accept it give the vector clock
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
