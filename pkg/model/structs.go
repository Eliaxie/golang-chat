package model

import (
	"sync"

	"golang.org/x/net/websocket"
)

type ConsistencyModel int

const (
	FIFO ConsistencyModel = iota
	CAUSAL
	GLOBAL
	LINEARIZABLE
)

type MessageType int

const (
	BASE MessageType = iota
	TEXT
	CONN_INIT
	CONN_INIT_RESPONSE
	CONN_RESTORE
	CONN_RESTORE_RESPONSE
	SYNC_PEERS
	SYNC_PEERS_RESPONSE
	GROUP_CREATE
)

type Message interface {
	GetMessageType() MessageType
}

type BaseMessage struct {
	MessageType MessageType `json:"messageType"`
}

func (m BaseMessage) GetMessageType() MessageType {
	return m.MessageType
}

type SyncPeersMessage struct {
	BaseMessage
}

type SyncPeersResponseMessage struct {
	BaseMessage
	Peers []SerializedClient `json:"peers"`
}

type GroupCreateMessage struct {
	BaseMessage
	Group            Group              `json:"group"`
	ConsistencyModel ConsistencyModel   `json:"consistencyModel"`
	Clients          []SerializedClient `json:"clients"`
}

type UniqueMessage struct {
	Text string `json:"content"`
	UUID string `json:"uuid"`
}

type TextMessage struct {
	BaseMessage
	Content     UniqueMessage `json:"content"`
	Group       Group         `json:"group"`
	VectorClock VectorClock   `json:"vectorClock"`
}

type ConnectionInitMessage struct {
	BaseMessage
	ClientID   string `json:"clientId"`
	ServerPort string `json:"serverPort"`
}

type ConnectionInitResponseMessage struct {
	BaseMessage
	ClientID string `json:"clientId"`
}

type ConnectionRestoreMessage struct {
	BaseMessage
	ClientID string `json:"clientId"`
}

type ConnectionRestoreResponseMessage struct {
	BaseMessage
	ClientID string `json:"clientId"`
}

type Client struct {
	Proc_id          string
	ConnectionString string
}

type SerializedClient struct {
	Proc_id  string `json:"proc_id"`
	HostName string `json:"hostName"`
}

type Group struct {
	Name   string
	Madeby string
}

type VectorClock struct {
	//Map proc_id to clock
	Clock map[string]int `json:"clock"`
}

type PendingMessage struct {
	Content     UniqueMessage
	Client      Client
	VectorClock VectorClock
}

type StableMessages struct {
	Content UniqueMessage
}

// Model
type Model struct {
	Myself     Client
	ServerPort string

	// map client_endpoint -> ws
	ClientWs map[string]*websocket.Conn
	// map client_endpoint -> client (before client init)
	PendingClients  map[string]*Client
	Clients         map[Client]bool
	PendingMessages map[Group][]PendingMessage
	StableMessages  map[Group][]StableMessages

	Groups             map[Group][]Client
	GroupsConsistency  map[Group]ConsistencyModel
	GroupsVectorClocks map[Group]VectorClock
	GroupsLocks        map[Group]*sync.Mutex
}

const (
	DEFAULT_PORT       = 8080
	DEFAULT_CONNECTION = "ws://localhost:8080/ws"
)
