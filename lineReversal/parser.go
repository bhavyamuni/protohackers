package lineReversal

import (
	"fmt"
	"strconv"
	"strings"
)

type MessageType string

const (
	ConnectMessageType MessageType = "connect"
	CloseMessageType   MessageType = "close"
	DataMessageType    MessageType = "data"
	AckMessageType     MessageType = "ack"
)

type Message interface{}

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

func (m *CloseMessage) String() string {
	return fmt.Sprintf("/close/%d/", m.Session)
}

func (m *DataMessage) String() string {
	return fmt.Sprintf("/data/%d/%s/", m.Session, m.Data)
}

func ParseMessage(message string) (Message, error) {
	messageChunks := SplitCommand(message)
	fmt.Println("Message chunks:", messageChunks)
	if len(messageChunks) < 2 {
		return nil, fmt.Errorf("invalid message: %s", message)
	}
	messageType := messageChunks[1]
	switch messageType {
	case string(ConnectMessageType):
		session, err := strconv.ParseInt(messageChunks[2], 10, 64)
		if err != nil {
			return nil, err
		}
		return &ConnectMessage{
			MessageType: ConnectMessageType,
			Session:     session,
		}, nil
	case string(CloseMessageType):
		session, err := strconv.ParseInt(messageChunks[2], 10, 64)
		if err != nil {
			return nil, err
		}
		return &CloseMessage{
			MessageType: CloseMessageType,
			Session:     session,
		}, nil
	case string(DataMessageType):
		session, err := strconv.ParseInt(messageChunks[2], 10, 64)
		if err != nil {
			return nil, err
		}

		fmt.Println("chunks: ", messageChunks[4])
		return &DataMessage{
			MessageType: DataMessageType,
			Session:     session,
			Data:        messageChunks[4],
		}, nil
	case string(AckMessageType):
		session, err := strconv.ParseInt(messageChunks[2], 10, 64)
		if err != nil {
			return nil, err
		}
		length, err := strconv.ParseInt(messageChunks[3], 10, 64)
		if err != nil {
			return nil, err
		}
		return &AckMessage{
			MessageType: AckMessageType,
			Session:     session,
			Length:      length,
		}, nil
	default:
		return nil, fmt.Errorf("unknown message type: %s", messageChunks[0])
	}
}

// SplitCommand splits a command string by slashes while handling escaped slashes.
// For example: "/connect/12345/" -> ["", "connect", "12345", ""]
// And: "/connect/123\/45/" -> ["", "connect", "123/45", ""]
func SplitCommand(cmd string) []string {
	var result []string
	var current strings.Builder
	escaped := false

	for i := 0; i < len(cmd); i++ {
		c := cmd[i]
		if escaped {
			current.WriteByte(c)
			escaped = false
			continue
		}
		if c == '\\' {
			escaped = true
			continue
		}
		if c == '/' {
			result = append(result, current.String())
			current.Reset()
			continue
		}
		current.WriteByte(c)
	}
	// Add the last segment
	result = append(result, current.String())

	fmt.Println("Split: ", result)
	return result
}

func ReverseAllLines(s string) string {
	lines := strings.Split(s, "\n")

	for ix, line := range lines {
		lines[ix] = ReverseString(line)
	}

	return strings.Join(lines, "\n")
}

func ReverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
