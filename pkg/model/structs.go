package model

import "golang.org/x/net/websocket"

type MessageType int

const (
	TEXT MessageType = iota
	BASE
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
	MessageType MessageType `json:"messageType"`
}

func (m BaseMessage) GetMessageType() MessageType {
	return BASE
}

type ConnectionRestoreResponseMessage struct {
	// tbd
}

func (m ConnectionRestoreResponseMessage) GetMessageType() MessageType {
	return CONN_RESTORE_RESPONSE
}

type SyncPeersMessage struct {
	BaseMessage
	PeerIDs []string `json:"peerIds"`
}

func (m SyncPeersMessage) GetMessageType() MessageType { return SYNC_PEERS }

type SyncPeersResponseMessage struct {
	// tbd
}

func (m SyncPeersResponseMessage) GetMessageType() MessageType {
	return SYNC_PEERS_RESPONSE
}

type GroupCreateMessage struct {
	BaseMessage
	Group   Group              `json:"group"`
	Clients []SerializedClient `json:"clients"`
}

func (m GroupCreateMessage) GetMessageType() MessageType { return GROUP_CREATE }

type TextMessage struct {
	BaseMessage
	Content     string      `json:"content"`
	Group       Group       `json:"group"`
	VectorClock VectorClock `json:"vectorClock"`
}

func (m TextMessage) GetMessageType() MessageType { return TEXT }

type ConnectionInitMessage struct {
	BaseMessage
	ClientID string `json:"clientId"`
}

func (m ConnectionInitMessage) GetMessageType() MessageType { return CONN_INIT }

type ConnectionInitResponseMessage struct {
	BaseMessage
	ClientID string `json:"clientId"`
}

func (m ConnectionInitResponseMessage) GetMessageType() MessageType { return CONN_INIT_RESP }

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
	Clients            map[*Client]bool
	Groups             map[Group][]Client
	GroupsBuffers      map[Group][]PendingMessage
	GroupsVectorClocks map[Group]VectorClock
}

const (
	DEFAULT_PORT       = 8080
	DEFAULT_CONNECTION = "ws://localhost:8080/ws"
)
