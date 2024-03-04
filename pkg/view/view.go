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
	port := displayInsertPort()
	_controller.StartServer(strconv.Itoa(port))
	log.Println("Server started on port ", port)
	username := displayInsertUsername()
	_controller.Model.Proc_id = username + "-" + _controller.GenerateUniqueID()
	log.Println("Username: ", _controller.Model.Proc_id)
	displayMainMenu()
}

func displayInsertPort() int {
	fmt.Print("Enter port number to start the server on (default ", model.DEFAULT_PORT, "): ")
	val, err := ReadInt()
	if err != nil {
		val = model.DEFAULT_PORT
	}
	return val
}

func displayInsertUsername() string {
	MoveScreenUp()
	fmt.Print("Enter username: ")
	return ReadStringTrimmed()
}

// Function to display the main menu
func displayMainMenu() {

	DisplayMenu([]MenuOption{
		{"Add new connections", displayAddNewConnectionsMenu},
		{"Create new Group", displayCreateNewGroup},
		{"Open Group", displayOpenGroup},
	})
}

func displayAddNewConnectionsMenu() {
	DisplayMenu([]MenuOption{
		{"Add Connection From File", displayAddConnectionFromFile},
		{"Add Connection Manually", displayAddConnectionManually},
		{"Back", displayMainMenu},
	})
}

func displayAddConnectionFromFile() {
	MoveScreenUp()
	fmt.Print("Enter the file path (\"q\" to go back): ")
	var connections []string
	for {
		filePath := ReadStringTrimmed()
		if filePath == "q" {
			displayAddNewConnectionsMenu()
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
	displayMainMenu()
}

func displayAddConnectionManually() {
	for {
		MoveScreenUp()
		fmt.Print("Enter Connection (default: ", model.DEFAULT_CONNECTION, ")(\"q\" to go back): ")
		connection := ReadStringTrimmed()
		if connection == "q" {
			displayAddNewConnectionsMenu()
			break
		}
		if connection == "" {
			connection = model.DEFAULT_CONNECTION
		}

		// call the function to add the connection
		log.Print(connection)
		pendingClient := _controller.AddNewConnection(connection)
		_controller.WaitForConnection(pendingClient)
		color.Green("Connected")

		// ask the user if they want to add another connection
		fmt.Println("Do you want to add another connection? (y/n)")
		choice := ReadStringTrimmed()
		if choice == "n" {
			break
		}
	}
	displayMainMenu()
}

func displayCreateNewGroup() {
	MoveScreenUp()
	fmt.Print("Enter the group name: ")
	groupName := ReadStringTrimmed()
	displayAddClientsToGroup(groupName)
}
