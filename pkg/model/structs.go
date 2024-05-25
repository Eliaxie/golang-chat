package model

import (
	"sync"

	"github.com/gorilla/websocket"
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
	MESSAGE_ACK
	CLIENT_DISC
	DISC_ACK
)

type ConnectionFlow int

const (
	FirstConnection ConnectionFlow = iota
	ReconnectionNetwork
	ReconnectionPeerCrashed
	ReconnectionSelfCrashed
)

func (m MessageType) String() string {
	names := [...]string{
		"BASE",
		"TEXT",
		"CONN_INIT",
		"CONN_INIT_RESPONSE",
		"CONN_RESTORE",
		"CONN_RESTORE_RESPONSE",
		"SYNC_PEERS",
		"SYNC_PEERS_RESPONSE",
		"GROUP_CREATE",
		"MESSAGE_ACK",
		"CLIENT_DISC",
		"DISC_ACK",
	}

	if m < BASE || m > DISC_ACK {
		return "UNKNOWN"
	}

	return names[m]
}

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
	ClientID     string `json:"clientId"`
	ServerIp     string `json:"serverPort"`
	Reconnection bool   `json:"reconnect"`
}

type ConnectionInitResponseMessage struct {
	BaseMessage
	Refused  bool   `json:"refused"`
	ClientID string `json:"clientId"`
}

type MessageAck struct {
	BaseMessage
	Group     Group               `json:"group"`
	Reference ScalarClockToProcId `json:"reference"`
}

type ClientDisconnectMessage struct {
	BaseMessage
	Group           Group            `json:"group"`    // group from which client disconnected
	Client          SerializedClient `json:"clientId"` // client that disconnected
	PendingMessages []PendingMessage `json:"pendingMessages"`
}

type DisconnectAckMessage struct {
	BaseMessage
	Group    Group  `json:"group"`
	ClientID string `json:"clientId"`
}

type ConnectionRestoreMessage struct {
	BaseMessage
	StableMessages            [][]StableMessage    `json:"stableMessages"`
	PendingMessages           [][]PendingMessage   `json:"pendingMessages"`
	Groups                    []Group              `json:"group"`
	SerializedClientsInGroups [][]SerializedClient `json:"serializedClientsInGroups"`
	ConsistencyModel          []ConsistencyModel   `json:"consistencyModel"`

	//Causal
	GroupsVectorClocks []VectorClock `json:"groupsVectorClocks"`
}

type ConnectionRestoreResponseMessage struct {
	BaseMessage
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

type ScalarClockToProcId struct {
	Clock   int    `json:"scalarClock"`
	Proc_id string `json:"proc_id"`
}

type PendingMessage struct {
	Content     UniqueMessage
	Client      Client
	VectorClock VectorClock
	ScalarClock ScalarClockToProcId
}

type StableMessage struct {
	Client  Client
	Content UniqueMessage
}

// Model
type Model struct {
	Myself     Client
	ServerPort string

	// map client_endpoint -> ws
	ClientWs map[string]*websocket.Conn
	// map client_endpoint -> client (before client init)
	PendingClients  map[string]struct{}
	Clients         map[Client]bool // map client -> bool (false if client is disconnected or not in the map)
	PendingMessages map[Group][]PendingMessage
	// group -> scalarClock -> array proc_id from which acks were received
	MessageAcks    map[Group]map[ScalarClockToProcId]map[string]bool
	StableMessages map[Group][]StableMessage

	DisconnectionAcks  map[Group]map[string]struct{} // maps group -> clients that sent back an ack
	DisconnectionLocks map[Group]*sync.Mutex         // locks for when we are waiting for acks and accesing Client Array
	Groups             map[Group][]Client
	GroupsConsistency  map[Group]ConsistencyModel
	GroupsVectorClocks map[Group]VectorClock
	GroupsLocks        map[Group]*sync.Mutex // groups need to be locked when we are modifying Groups, Clients, VectorClocks, PendingMessages, StableMessages. Groups are also locked when a disconnection is happening or if the group is not in a majority partition. Groups are also locked when a reconnection is happening.

	MessageExitBuffer     map[Client][][]byte
	MessageExitBufferLock *sync.Mutex
}

const (
	DEFAULT_PORT       = 8080
	DEFAULT_CONNECTION = "ws://localhost:8080/ws"
)
