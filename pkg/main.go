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
		Proc_id:            "",
		PendingClients:     make(map[model.Client]bool),
		Clients:            make(map[model.Client]bool),
		GroupsBuffers:      make(map[model.Group][]model.PendingMessage),
		Groups:             make(map[model.Group][]model.Client),
		GroupsVectorClocks: make(map[model.Group]model.VectorClock),
		GroupsConsistency:  make(map[model.Group]model.ConsistencyModel),
	}

	_controller := controller.Controller{Model: globModel}

	view.Start(&_controller)

}
