package main

import (
	"golang-chat/pkg/controller"
	"golang-chat/pkg/model"

	//"os"

	"golang-chat/pkg/view"
)

// create a variable model of type Model
var globModel *model.Model = &model.Model{}

func main() {

	// initialize model
	globModel = &model.Model{
		Name:               "",
		Clients:            make(map[*model.Client]bool),
		GroupsBuffers:      make(map[model.GroupName][]model.PendingMessage),
		Groups:             make(map[model.GroupName][]model.Client),
		GroupsVectorClocks: make(map[model.GroupName]model.VectorClock),
	}

	_controller := controller.Controller{Model: globModel}

	// reader := bufio.NewReader(os.Stdin)
	// fmt.Print("Enter port number to start the server on (e.g., 8080): ")
	// port, _ := reader.ReadString('\n')
	// port = strings.TrimSpace(port)

	// fmt.Printf("Starting WebSocket server on port %s\n", port)
	// go startServer(port)

	// connectAndCommunicate(reader)

	view.Start(&_controller)

}
