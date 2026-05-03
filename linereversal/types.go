package linereversal

import (
	"fmt"
	"time"
)

type MessageType string

const RetransmissionTimeout = 3 * time.Second

const (
	ConnectMessageType MessageType = "connect"
	CloseMessageType   MessageType = "close"
	DataMessageType    MessageType = "data"
	AckMessageType     MessageType = "ack"
)

type Message interface {
	String() string
}

type ConnectMessage struct {
	MessageType
	Session int64
}

type CloseMessage struct {
	MessageType
	Session int64
}

type DataMessage struct {
	MessageType
	Session int64
	Pos     int64
	Data    string
}

type AckMessage struct {
	MessageType
	Session int64
	Length  int64
}

func (m *AckMessage) String() string {
	return fmt.Sprintf("/ack/%d/%d/", m.Session, m.Length)
}

func (m *ConnectMessage) String() string {
	return fmt.Sprintf("/connect/%d/", m.Session)
}

func (m *CloseMessage) String() string {
	return fmt.Sprintf("/close/%d/", m.Session)
}

func (m *DataMessage) String() string {
	return fmt.Sprintf("/data/%d/%d/%s/", m.Session, m.Pos, encodeData(m.Data))
}
