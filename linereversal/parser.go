package linereversal

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseMessage(message string) (Message, error) {
	messageChunks := strings.Split(message, "/")
	if len(messageChunks) < 4 {
		return nil, fmt.Errorf("invalid message: %s", message)
	}
	messageChunks = messageChunks[1 : len(messageChunks)-1]
	messageType := messageChunks[0]
	session, err := strconv.ParseInt(messageChunks[1], 10, 32)
	if err != nil {
		return nil, err
	}

	switch messageType {
	case string(ConnectMessageType):
		if len(messageChunks) != 2 {
			return nil, fmt.Errorf("Validation failed for message %s", message)
		}
		return &ConnectMessage{
			MessageType: ConnectMessageType,
			Session:     session,
		}, nil
	case string(CloseMessageType):
		if len(messageChunks) != 2 {
			return nil, fmt.Errorf("Validation failed for message %s", message)
		}
		return &CloseMessage{
			MessageType: CloseMessageType,
			Session:     session,
		}, nil
	case string(DataMessageType):
		if len(messageChunks) < 4 {
			return nil, fmt.Errorf("invalid data message: %s", message)
		}
		pos, err := strconv.ParseInt(messageChunks[2], 10, 64)
		if err != nil {
			return nil, err
		}
		rawData := strings.Join(messageChunks[3:], "/")
		decoded, err := decodeData(rawData)
		if err != nil {
			return nil, fmt.Errorf("Validation failed for message %s", message)
		}
		return &DataMessage{
			MessageType: DataMessageType,
			Session:     session,
			Pos:         pos,
			Data:        decoded,
		}, nil
	case string(AckMessageType):
		if len(messageChunks) != 3 {
			return nil, fmt.Errorf("invalid data message: %s", message)
		}
		length, err := strconv.ParseInt(messageChunks[2], 10, 64)
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

func encodeData(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' || s[i] == '/' {
			b.WriteByte('\\')
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func decodeData(s string) (string, error) {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			return b.String(), fmt.Errorf("Extra characters exist, this is a malformed packet")
		}
		if s[i] == '\\' {
			if i+1 >= len(s) {
				return b.String(), fmt.Errorf("dangling \\ in data, malformed packet")
			}
			if s[i+1] != '\\' && s[i+1] != '/' {
				return b.String(), fmt.Errorf("invalid escape \\%c in data, malformed packet", s[i+1])
			}
			i++
		}
		b.WriteByte(s[i])
	}
	return b.String(), nil
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
