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
		GroupsBuffers:      make(map[model.Group][]model.PendingMessage),
		Groups:             make(map[model.Group][]model.Client),
		GroupsVectorClocks: make(map[model.Group]model.VectorClock),
	}

	_controller := controller.Controller{Model: globModel}

	// _controller.SendMessage(model.TextMessage{Content: "Hello", Group: model.Group{Name: "default", Madeby: "default"}, VectorClock: model.VectorClock{}}, model.Client{})

	view.Start(&_controller)

}
