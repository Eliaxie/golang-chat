package main

import (
	"fmt"
	"golang-chat/pkg/utils"
)

type Controller struct {
	model utils.Model
}

func HandleTextMessage(c *Controller, textMsg utils.TextMessage, client utils.Client) {
	fmt.Println("Received text message:", textMsg.Content)
	c.model.GroupsBuffers[textMsg.Group] =
		append(c.model.GroupsBuffers[textMsg.Group], utils.PendingMessage{Content: textMsg.Content, Client: client, VectorClock: textMsg.VectorClock})
}
