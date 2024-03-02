package view

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"

	"golang-chat/pkg/controller"
	"golang-chat/pkg/model"
	"golang-chat/pkg/utils"

	"github.com/fatih/color"
)

var reader = bufio.NewReader(os.Stdin)
var _controller *controller.Controller

func Start(c *controller.Controller) {
	_controller = c
	color.Green("Welcome to the chat app")
	port := DisplayInsertPort()
	_controller.StartServer(strconv.Itoa(port))
	// clear the console
	MoveScreenUp()
	log.Println("Server started on port ", port)
	username := DisplayInsertUsername()
	log.Println("Username entered: ", username)
	MoveScreenUp()
	DisplayMainMenu()
}

func DisplayInsertPort() int {
	fmt.Print("Enter port number to start the server on (default ", model.DEFAULT_PORT, "): ")
	val, err := ReadInt()
	if err != nil {
		val = model.DEFAULT_PORT
	}
	return val
}

func DisplayInsertUsername() string {
	fmt.Print("Enter username: ")
	return ReadStringTrimmed()
}

// Function to display the main menu
func DisplayMainMenu() {
	DisplayMenu([]MenuOption{
		{"Add new connections", DisplayAddNewConnectionsMenu},
		{"Create new Group", DisplayCreateNewGroup},
		{"Open Group", DisplayOpenGroup},
	})
}

func DisplayAddNewConnectionsMenu() {
	// use DisplayMenu to display the options
	DisplayMenu([]MenuOption{
		{"Add Connection From File", DisplayAddConnectionFromFile},
		{"Add Connection Manually", DisplayAddConnectionManually},
		{"Back", DisplayMainMenu},
	})
}

func DisplayAddConnectionFromFile() {
	fmt.Print("Enter the file path (\"q\" to go back): ")
	var connections []string
	for {
		filePath := ReadStringTrimmed()
		if filePath == "q" {
			DisplayAddNewConnectionsMenu()
			break
		}

		// call the function to add the connections from the file
		var err error
		connections, err = utils.ReadConnectionsFromFile(filePath)
		if err == nil {
			_controller.AddNewConnections(connections)
			break
		}
		fmt.Println("Error while trying to read the file. Please try again. (\"q\" to go back)")
	}
	log.Println("Connections added successfully")
	log.Print(connections)
	DisplayMainMenu()
}

func DisplayAddConnectionManually() {
	for {
		fmt.Print("Enter Connection (default: ", model.DEFAULT_CONNECTION, ")(\"q\" to go back): ")
		connection := ReadStringTrimmed()
		if connection == "q" {
			DisplayAddNewConnectionsMenu()
			break
		}
		if connection == "" {
			connection = model.DEFAULT_CONNECTION
		}

		// call the function to add the connection
		log.Print(connection)
		_controller.AddNewConnection(connection)
		// ask the user if they want to add another connection
		fmt.Println("Do you want to add another connection? (y/n)")
		choice := ReadStringTrimmed()
		MoveScreenUp()
		if choice == "n" {
			break
		}
	}
	DisplayMainMenu()
}

func DisplayCreateNewGroup() {
	fmt.Print("Enter the group name: ")
	groupName := ReadStringTrimmed()
	// call the function to create a new group
	log.Println("Group" + groupName + " created successfully")
	DisplayMainMenu()
}

func DisplayOpenGroup() {
	fmt.Print("Enter the group name: ")
	groupName := ReadStringTrimmed()
	// call the function to open the group
	log.Println("Opening group " + groupName)
	// start reading messages from a certain group
	ClearScreen()
	DisplayRoom(groupName)
}




