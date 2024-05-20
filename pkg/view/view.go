package view

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"

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
	extPort := displayInsertExtPort(port)
	_controller.StartServer(strconv.Itoa(port), strconv.Itoa(extPort))
	log.Infoln("Server started on port ", port, "-> ", extPort)
	username := displayInsertUsername()
	Proc_id := username + "-" + _controller.GenerateUniqueID()
	_controller.Model.Myself = model.Client{Proc_id: Proc_id}
	log.Infoln("Username: ", Proc_id)
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

func displayInsertExtPort(port int) int {
	fmt.Print("Enter external port (default: ", port, "): ")
	val, err := ReadInt()
	if err != nil {
		val = port
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
	for {
		DisplayMenu([]MenuOption{
			{"Add new connections", displayAddNewConnectionsMenu},
			{"Create new Group", displayCreateNewGroup},
			{"Open Group", displayOpenGroup},
		})
	}
}

func displayAddNewConnectionsMenu() {
	inMenu := true
	for inMenu {
		DisplayMenu([]MenuOption{
			{"Add Connection From File", displayAddConnectionFromFile},
			{"Add Connection Manually", displayAddConnectionManually},
			{"Back", func() { inMenu = false }},
		})
	}
}

func displayAddConnectionFromFile() {
	MoveScreenUp()
	fmt.Print("Enter the file path (\"q\" to go back): ")
	var connections []string
	for {
		filePath := ReadStringTrimmed()
		if filePath == "q" {
			return
		}

		// call the function to add the connections from the file
		var err error
		connections, err = utils.ReadConnectionsFromFile(filePath)
		if err != nil {
			_controller.AddNewConnections(connections)
			break
		}
		fmt.Println("Error while trying to read the file. Please try again. (\"q\" to go back)")
	}
	log.Infoln("Connections added successfully")
	log.Debugln(connections)
}

func displayAddConnectionManually() {
	MoveScreenUp()
	fmt.Print("Enter Connection (default: ", model.DEFAULT_CONNECTION, ")(\"q\" to go back): ")
	connection := ReadStringTrimmed()
	if connection == "q" {
		return
	}
	if connection == "" {
		connection = model.DEFAULT_CONNECTION
	}

	// call the function to add the connection
	log.Debugln(connection)
	pendingClient, err := _controller.AddNewConnection(connection)
	if err != nil {
		log.Errorln("Error adding new connection", err)
		return
	}
	_controller.WaitForConnection(pendingClient)
	color.Green("Connected")
}
