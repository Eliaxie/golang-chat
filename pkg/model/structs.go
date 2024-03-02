package model

import "golang.org/x/net/websocket"

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
	PeerIDs []string `json:"peerIds"`
}

type SyncPeersResponseMessage struct {
	// tbd
}

type GroupCreateMessage struct {
	BaseMessage
	Group            Group              `json:"group"`
	ConsistencyModel ConsistencyModel   `json:"consistencyModel"`
	Clients          []SerializedClient `json:"clients"`
}

type TextMessage struct {
	BaseMessage
	Content     string      `json:"content"`
	Group       Group       `json:"group"`
	VectorClock VectorClock `json:"vectorClock"`
}

type ConnectionInitMessage struct {
	BaseMessage
	ClientID string `json:"clientId"`
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
	Ws      *websocket.Conn
	Proc_id string
}

type SerializedClient struct {
	Proc_id  string `json:"proc_id"`
	HostName string `json:"hostName"`
}

func NewClient(ws *websocket.Conn) Client {
	return Client{ws, ""}
}

type Group struct {
	Name   string
	Madeby string
}

type VectorClock struct {
	Clock map[string]int `json:"clock"`
}

type PendingMessage struct {
	Content     string
	Client      Client
	VectorClock VectorClock
}

// Model
type Model struct {
	Name               string // username-uniqueIdentifier
	PendingClients     map[Client]bool
	Clients            map[Client]bool
	GroupsConsistency  map[Group]ConsistencyModel
	Groups             map[Group][]Client
	GroupsBuffers      map[Group][]PendingMessage
	GroupsVectorClocks map[Group]VectorClock
}

const (
	DEFAULT_PORT       = 8080
	DEFAULT_CONNECTION = "ws://localhost:8080/ws"
)
