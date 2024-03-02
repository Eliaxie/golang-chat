package model

import "golang.org/x/net/websocket"

type MessageType int

const (
	TEXT MessageType = iota
	CONN_INIT
	CONN_INIT_RESP
	CONN_RESTORE
	CONN_RESTORE_RESPONSE
	SYNC_PEERS
	SYNC_PEERS_RESPONSE
	GROUP_CREATE
)

// Generic Message Interface
type Message interface {
	GetMessageType() MessageType
}

type BaseMessage struct {
	MessageType MessageType
}

func (m BaseMessage) GetMessageType() MessageType {
	return CONN_RESTORE_RESPONSE
}

type ConnectionRestoreResponseMessage struct {
	// tbd
}

func (m ConnectionRestoreResponseMessage) GetMessageType() MessageType {
	return CONN_RESTORE_RESPONSE
}

type SyncPeersMessage struct {
	MessageType MessageType `json:"messageType"`
	PeerIDs     []string    `json:"peerIds"`
}

func (m SyncPeersMessage) GetMessageType() MessageType { return SYNC_PEERS }

type SyncPeersResponseMessage struct {
	// tbd
}

func (m SyncPeersResponseMessage) GetMessageType() MessageType {
	return SYNC_PEERS_RESPONSE
}

type GroupCreateMessage struct {
	// tbd
}

func (m GroupCreateMessage) GetMessageType() MessageType { return GROUP_CREATE }

type TextMessage struct {
	MessageType MessageType `json:"messageType"`
	Content     string      `json:"content"`
	Group       GroupName   `json:"group"`
	VectorClock VectorClock `json:"vectorClock"`
}

func (m TextMessage) GetMessageType() MessageType { return TEXT }

type ConnectionInitMessage struct {
	MessageType MessageType `json:"messageType"`
	ClientID    string      `json:"clientId"`
}

func (m ConnectionInitMessage) GetMessageType() MessageType { return CONN_INIT }

type ConnectionInitResponseMessage struct {
	MessageType MessageType `json:"messageType"`
	ClientID    string      `json:"clientId"`
}

func (m ConnectionInitResponseMessage) GetMessageType() MessageType { return CONN_INIT_RESP }

type Client struct {
	Ws      *websocket.Conn
	Proc_id string
}

func NewClient(ws *websocket.Conn) Client {
	return Client{ws, ""}
}

type GroupName struct {
	Name   string
	Madeby string
}

type VectorClock struct {
	Clock map[Client]int
}

type PendingMessage struct {
	Content     string
	Client      Client
	VectorClock VectorClock
}

// Model
type Model struct {
	Name               string
	Clients            map[*Client]bool
	Groups             map[GroupName][]Client
	GroupsBuffers      map[GroupName][]PendingMessage
	GroupsVectorClocks map[GroupName]VectorClock
}

const (
	DEFAULT_PORT       = 8080
	DEFAULT_CONNECTION = "ws://localhost:8080/ws"
)
