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
	"golang.org/x/net/websocket"

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
	model := &model.Model{
		// name of yourself
		Myself: model.Client{},

		// clients endpoint to ws
		ClientWs: make(map[string]*websocket.Conn),
		// clients before the handshake
		PendingClients: make(map[string]struct{}),
		// clients after the handshake
		Clients: make(map[model.Client]bool),

		// messages that have been received or sent but not yet trasmitted to the app level
		PendingMessages: make(map[model.Group][]model.PendingMessage),
		// messages shown to the users
		StableMessages: make(map[model.Group][]model.StableMessage),
		// list of acks for each message
		MessageAcks: make(map[model.Group]map[model.ScalarClockToProcId]map[string]bool),

		// groups
		Groups:             make(map[model.Group][]model.Client),
		DisconnectionAcks:  make(map[model.Group]map[string]struct{}),
		DisconnectionLocks: make(map[model.Group]*sync.Mutex),
		GroupsConsistency:  make(map[model.Group]model.ConsistencyModel),
		GroupsVectorClocks: make(map[model.Group]model.VectorClock),
		GroupsLocks:        make(map[model.Group]*sync.Mutex),

		MessageExitBuffer:     make(map[model.Client][][]byte),
		MessageExitBufferLock: &sync.Mutex{},
	}

	// initialize notifier
	notifier := notify.NewNotifier()

	_controller := controller.Controller{Model: model, Notifier: notifier}

	go _controller.StartRetryConnections()
	go _controller.StartRetryMessages()

	// starts view
	view.Start(&_controller)

}
