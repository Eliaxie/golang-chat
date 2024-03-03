package view

import (
	"golang-chat/pkg/model"
	"log"

	"github.com/fatih/color"
)

var clientsToAdd []model.Client

func displayAddClientsToGroup(groupName string) {
	MoveScreenUp()
	color.Green("Add clients to group " + groupName)

	// add a menu option for each client
	var menuOptions []MenuOption
	//
	for client := range _controller.Model.Clients {

		// check if the client is already in the list of clients to add
		found := func() bool {
			for _, val := range clientsToAdd {
				if val == client {
					return true
				}
			}
			return false
		}()

		if found {
			continue
		}

		menuOptions = append(menuOptions, MenuOption{client.Proc_id, func() {
			clientsToAdd = append(clientsToAdd, client)
			log.Println("Client " + client.Proc_id + " added to group " + groupName)
			displayAddClientsToGroup(groupName)
		}})
	}
	menuOptions = append(menuOptions, MenuOption{"Back", func() {
		_controller.CreateGroup(groupName, clientsToAdd)
		log.Println("Group " + groupName + " created successfully")
		DisplayMainMenu()
	}})

	DisplayMenu(menuOptions)
}

func displayOpenGroup() {
	MoveScreenUp()
	color.Green("Select group to open:")

	// list of groups as MenuOptions
	var groups []MenuOption
	for group := range _controller.Model.Groups {
		groups = append(groups, MenuOption{group.Name, func() { DisplayGroup(group) }})
	}
	groups = append(groups, MenuOption{"Back", DisplayMainMenu})
	DisplayMenu(groups)
}

func DisplayGroup(group model.Group) {
	MoveScreenUp()
	color.Green("Entering room %s ( type '/exit' to leave the room )", group)
	inputLoop(group)
}

// function that loops and waits for input from the user
func inputLoop(group model.Group) {
	for {
		value := ReadStringTrimmed()
		if value == "/exit" {
			DisplayMainMenu()
			break
		}
		_controller.SendGroupMessage(value, group)
	}
}
