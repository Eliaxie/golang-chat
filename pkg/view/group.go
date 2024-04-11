package view

import (
	"fmt"
	"golang-chat/pkg/model"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/fatih/color"
)

var currentMessage int

type GroupCreationInfo struct {
	GroupName        string
	ConsistencyModel model.ConsistencyModel
}

func displayCreateNewGroup() {
	MoveScreenUp()
	fmt.Print("Enter the group name: ")
	groupName := ReadStringTrimmed()
	groupInfo := GroupCreationInfo{GroupName: groupName}

	MoveScreenUp()
	color.Green("Choose consistency model for group:")

	selectModel := func(consistencyModel *model.ConsistencyModel, selectedModel model.ConsistencyModel) {
		*consistencyModel = selectedModel
	}

	DisplayMenu([]MenuOption{
		{"FIFO", func() {
			selectModel(&groupInfo.ConsistencyModel, model.FIFO)
		}},
		{"CAUSAL", func() {
			selectModel(&groupInfo.ConsistencyModel, model.CAUSAL)
		}},
		{"GLOBAL", func() {
			selectModel(&groupInfo.ConsistencyModel, model.GLOBAL)
		}},
		{"LINEARIZABLE", func() {
			selectModel(&groupInfo.ConsistencyModel, model.LINEARIZABLE)
		}},
	})

	displayAddClientsToGroup(groupInfo)

}

func displayAddClientsToGroup(groupInfo GroupCreationInfo) {
	MoveScreenUp()
	color.Green("Add clients to group " + groupInfo.GroupName)

	// add a menu option for each client

	adding := true
	var clientsToAdd []model.Client
	for adding {
		var menuOptions []MenuOption
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

			// check if client is connected
			if !_controller.Model.Clients[client] {
				continue
			}

			menuOptions = append(menuOptions, MenuOption{client.Proc_id, func() {
				clientsToAdd = append(clientsToAdd, client)
			}})
		}
		// add done option
		menuOptions = append(menuOptions, MenuOption{"[Create group]", func() {
			_controller.CreateGroup(groupInfo.GroupName, groupInfo.ConsistencyModel, clientsToAdd)
			log.Infoln("Group " + groupInfo.GroupName + " created successfully")
			adding = false
		}})
		// add back option
		menuOptions = append(menuOptions, MenuOption{"[BACK]", func() { adding = false }})

		DisplayMenu(menuOptions)
	}
}

func displayOpenGroup() {
	MoveScreenUp()
	color.Green("Select group to open:")

	// list of groups as MenuOptions
	var groups []MenuOption
	for group := range _controller.Model.Groups {
		groups = append(groups, MenuOption{group.Name, func() { displayGroup(group) }})
	}
	groups = append(groups, MenuOption{"Back", displayMainMenu})
	DisplayMenu(groups)
}

func displayGroup(group model.Group) {
	MoveScreenUp()
	color.Green("Previous messages in group:")
	for _, message := range _controller.Model.StableMessages[group] {
		color.Yellow(message.Content.Text)
	}
	color.Green("Entering room %s ( type '/exit' to leave the room '/list' to see other members )", group)
	_controller.Notifier.Listen(group, UpdateGroup)
	inputLoop(group)
}

// function that loops and waits for input from the user
func inputLoop(group model.Group) {
	for {
		value := ReadStringTrimmed()
		if value == "/exit" {
			_controller.Notifier.Remove(group)
			currentMessage = 0
			displayMainMenu()
			break
		}
		if value == "/list" {
			displayGroupMembers(group)
			continue
		}

		_controller.SendGroupMessage(value, group)
	}
}

func displayGroupMembers(group model.Group) {
	color.Green("Members of group %s", group)
	for _, client := range _controller.Model.Groups[group] {
		color.White("- %s", client.Proc_id)
	}
}

func UpdateGroup(group model.Group) {
	log.Debugln("Updating group " + group.Name + " made by " + group.Madeby)
	var stableMessages = _controller.Model.StableMessages[group]
	for i := currentMessage; i < len(stableMessages); i++ {
		fmt.Print(color.RedString(strings.Split(stableMessages[i].Client.Proc_id, "-")[0] + ": "))
		fmt.Println(color.YellowString(stableMessages[i].Content.Text))
		currentMessage++
	}
}

func isolateNameFromProcId(proc_id string) string {
	return strings.Split(proc_id, "-")[0]
}
