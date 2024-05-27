package view

import (
	"fmt"
	"golang-chat/pkg/maps"
	"golang-chat/pkg/model"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/fatih/color"
)

var currentMessage int
var groupUserColors = make(map[model.Group]map[string]*color.Color)

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

			// check if client is connected
			if !maps.Load(&_controller.Model.Clients, client) {
				continue
			}

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
			if !maps.Load(&_controller.Model.Clients, client) {
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
			DisplayString("Group "+groupInfo.GroupName+" created successfully", color.BgGreen)
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
	for _, group := range maps.Keys(&_controller.Model.Groups) {
		groups = append(groups, MenuOption{group.Name, func() { displayGroup(group) }})
	}
	groups = append(groups, MenuOption{"Back", displayMainMenu})
	DisplayMenu(groups)
}

func displayGroup(group model.Group) {
	MoveScreenUp()
	if groupUserColors[group] == nil {
		initializeGroupColors(group)
	}
	color.Green("Previous messages in group:")
	UpdateGroup(group)
	color.Green("Entering room %s ( type '/exit' to leave the room '/list' to see other members )", group)
	_controller.Notifier.Listen(group, UpdateGroup)
	inputLoop(group)
}

func initializeGroupColors(group model.Group) {
	groupUserColors[group] = make(map[string]*color.Color)
	// clients := _controller.Model.Groups[group]
	clients := maps.Load(&_controller.Model.Groups, group)
	for _, client := range clients {
		groupUserColors[group][client.Proc_id] = color.New(RandomColor())
	}
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
	// for _, client := range _controller.Model.Groups[group] {
	for _, client := range maps.Load(&_controller.Model.Groups, group) {
		if maps.Load(&_controller.Model.Clients, client) || client == _controller.Model.Myself {
			color.White("- %s", client.Proc_id)
		} else {
			color.Red("disc - %s", client.Proc_id)
		}
	}
}

func UpdateGroup(group model.Group) {
	log.Debugln("Updating group " + group.Name + " made by " + group.Madeby)
	var stableMessages = maps.Load(&_controller.Model.StableMessages, group)
	for _, message := range stableMessages[currentMessage:] {
		userName := strings.Split(message.Client.Proc_id, "-")[0] + ": "
		fmt.Print(groupUserColors[group][message.Client.Proc_id].Sprint(userName))
		fmt.Println(message.Content.Text)
		currentMessage++
	}
}
