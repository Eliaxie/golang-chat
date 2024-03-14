package main

import (
	"golang-chat/pkg/controller"
	"golang-chat/pkg/model"
	"golang-chat/pkg/notify"
	"golang-chat/pkg/utils"
	"os"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"golang-chat/pkg/view"
)

// create a variable model of type Model

func main() {

	level := 0
	if len(os.Args) > 1 {
		if strings.Contains(os.Args[1], "-v") {
			level = strings.Count(os.Args[1], "v")
		}
	}
	utils.LogInit(log.Level(level))

	// initialize model
	globModel := &model.Model{
		Proc_id:            "",
		PendingClients:     make(map[model.Client]bool),
		Clients:            make(map[model.Client]bool),
		PendingMessages:    make(map[model.Group][]model.PendingMessage),
		StableMessages:     make(map[model.Group][]model.StableMessages),
		Groups:             make(map[model.Group][]model.Client),
		GroupsVectorClocks: make(map[model.Group]model.VectorClock),
		GroupsConsistency:  make(map[model.Group]model.ConsistencyModel),
		GroupsLocks:        make(map[model.Group]*sync.Mutex),
	}

	// initialize notifier
	notifier := notify.NewNotifier()

	_controller := controller.Controller{Model: globModel, Notifier: notifier}


	// starts view
	view.Start(&_controller)

}
