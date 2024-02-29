package main

import "fmt"

type Controller struct {
	model Model
}

func HandleTextMessage(c *Controller, textMsg TextMessage, client Client) {
	fmt.Println("Received text message:", textMsg.Content)
	model.GroupsBuffers[textMsg.Group] = append(model.GroupsBuffers[textMsg.Group], PendingMessage{textMsg.Content, client, textMsg.VectorClock})
}
