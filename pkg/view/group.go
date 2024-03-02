package view

import (
	"log"

	"github.com/fatih/color"
)

func DisplayRoom(roomName string) {
	color.Yellow("Entering room %s ( type '/exit' to leave the room )", roomName)
	ListenForMessages()
}

// function that loops and listens for messages every second
func ListenForMessages() {
	for {
		value := ReadStringTrimmed()
		if value == "/exit" {
			DisplayMainMenu()
			break
		}

		log.Println("Sending message: ", value)
		_controller.BroadcastMessage(value)
	}
}
