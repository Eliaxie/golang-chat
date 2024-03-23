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

	// initialize logger level
	level := 0
	if len(os.Args) > 1 {
		if strings.Contains(os.Args[1], "-v") {
			level = strings.Count(os.Args[1], "v")
		}
	}
	utils.LogInit(log.Level(level))

	// initialize model
	globModel := &model.Model{
		// name of yourself
		Proc_id: "",

		// clients before the handshake
		PendingClients: make(map[model.Client]bool),

		// clients after the handshake
		Clients: make(map[model.Client]bool),

		// messages that have been received or sent but not yet trasmitted to the app level
		PendingMessages: make(map[model.Group][]model.PendingMessage),
		// messages shown to the users
		StableMessages: make(map[model.Group][]model.StableMessages),

		// groups
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
