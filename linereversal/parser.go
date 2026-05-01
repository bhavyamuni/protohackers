package linereversal

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseMessage(message string) (Message, error) {
	messageChunks := SplitCommand(message)
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
		pos, err := strconv.ParseInt(messageChunks[3], 10, 64)
		if err != nil {
			return nil, err
		}
		return &DataMessage{
			MessageType: DataMessageType,
			Session:     session,
			Pos:         pos,
			Data:        decodeData(messageChunks[4]),
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

// SplitCommand splits a command string by unescaped slashes, preserving escape sequences in segments.
func SplitCommand(cmd string) []string {
	var result []string
	var current strings.Builder

	for i := 0; i < len(cmd); i++ {
		c := cmd[i]
		if c == '\\' && i+1 < len(cmd) {
			current.WriteByte(c)
			i++
			current.WriteByte(cmd[i])
			continue
		}
		if c == '/' {
			result = append(result, current.String())
			current.Reset()
			continue
		}
		current.WriteByte(c)
	}
	result = append(result, current.String())
	return result
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

func decodeData(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			i++
		}
		b.WriteByte(s[i])
	}
	return b.String()
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
