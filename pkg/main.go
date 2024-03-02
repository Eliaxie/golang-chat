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

	view.Start(&_controller)

}
