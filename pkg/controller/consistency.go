package controller

import "golang-chat/pkg/model"

func (c *Controller) tryAcceptCasualMessages(pendingMessages *[]model.PendingMessage, stableMessages *[]model.StableMessages, message model.TextMessage, client model.Client) bool {

	return false
}

func (c *Controller) tryAcceptGlobalMessages(pendingMessages *[]model.PendingMessage,
	stableMessages *[]model.StableMessages, message model.TextMessage, client model.Client) bool {

	return false
}

func (c *Controller) tryAcceptLinearizableMessages(pendingMessages *[]model.PendingMessage,
	stableMessages *[]model.StableMessages, message model.TextMessage, client model.Client) bool {

	return false
}

func (c *Controller) tryAcceptFIFOMessages(pendingMessages *[]model.PendingMessage,
	stableMessages *[]model.StableMessages, message model.TextMessage, client model.Client) bool {
	changes := false
	for _, message := range *pendingMessages {
		changes = true
		a := append(*stableMessages, model.StableMessages{Content: message.Content})
		stableMessages = &a
	}
	return changes
}
