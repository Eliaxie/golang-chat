package view

import (
	"github.com/fatih/color"
)

func DisplayRoom(roomName string) {
	color.Yellow("Entering room %s", roomName)
	ListenForMessages()
}

// function that loops and listens for messages every second
func ListenForMessages() {
	for {
		value := ReadStringTrimmed()
		_controller.BroadcastMessage(value)
	}
}
