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
	_controller.Notifier.ListenView(DisplayString)
	port := displayInsertPort()
	extIp := displayInsertExtPort("ws://localhost:" + strconv.Itoa(port) + "/ws")
	username := displayInsertUsername()
	Proc_id := username + "-" + _controller.GenerateUniqueID()
	_controller.Model.Myself.Proc_id = Proc_id
	log.Infoln("Username: ", Proc_id)
	_controller.StartServer(strconv.Itoa(port), extIp)
	log.Infoln("Server started on port ", port, "-> ", extIp)
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

func displayInsertExtPort(localIp string) string {
	defIp, ok := os.LookupEnv("GOLANGCHAT_DEFAULT_EXTERNALIP")
	if !ok {
		defIp = localIp
	}
	fmt.Print("Enter external ip (default: ", defIp, "): ")
	val := ReadStringTrimmed()
	if val == "" {
		val = defIp
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
	defConnection, ok := os.LookupEnv("GOLANGCHAT_DEFAULT_INITIALCONNECTION")
	if !ok {
		// Handle the case where the environment variable is not set
		defConnection = model.DEFAULT_CONNECTION
	}

	fmt.Print("Enter Connection (default: ", defConnection, ")(\"q\" to go back): ")
	connection := ReadStringTrimmed()
	if connection == "q" {
		return
	}
	if connection == "" {
		connection = defConnection
	}

	// call the function to add the connection
	pendingClient, err := _controller.AddNewConnection(connection)
	if err != nil {
		log.Errorln("Error adding new connection", err)
		return
	}
	_controller.WaitForConnection(pendingClient)
	DisplayString("Connection added successfully", color.BgGreen, color.FgHiWhite)
}

func DisplayString(str string, colors ...color.Attribute) {
	// print black string with green background
	backgroundColor := color.BgYellow
	textColor := color.FgHiWhite
	if len(colors) >= 1 {
		backgroundColor = colors[0]
		if len(colors) >= 2 {
			textColor = colors[1]
		}
	}
	color.New(backgroundColor, textColor).Println(str)
}
