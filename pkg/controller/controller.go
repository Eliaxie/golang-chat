package main

import (
	"fmt"
	"golang-chat/pkg/model"
)

type Controller struct {
	model model.Model
}

func HandleTextMessage(c *Controller, textMsg model.TextMessage, client model.Client) {
	fmt.Println("Received text message:", textMsg.Content)
	c.model.GroupsBuffers[textMsg.Group] =
		append(c.model.GroupsBuffers[textMsg.Group], model.PendingMessage{Content: textMsg.Content, Client: client, VectorClock: textMsg.VectorClock})
}
