package speeddaemon

import (
	"bufio"
	"encoding/binary"
	"errors"
	"net"
)

const (
	ErrorMessageType         MessageType = 0x10
	PlateMessageType         MessageType = 0x20
	TicketMessageType        MessageType = 0x21
	WantHeartbeatMessageType MessageType = 0x40
	HeartbeatMessageType     MessageType = 0x41
	IAmCameraMessageType     MessageType = 0x80
	IAmDispatcherMessageType MessageType = 0x81
)

type MessageType byte

type Message interface {
	Handle(s *SpeedDaemonServer, conn *net.Conn)
}

func ParseMessage(buf *bufio.Reader) (Message, error) {
	bufType, err := buf.Peek(1)
	if err != nil {
		return nil, err
	}
	mType := MessageType(bufType[0])
	switch mType {
	case PlateMessageType:
		plateMsg := &PlateMessage{MessageType: mType}
		buf.ReadByte()
		numLength, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}
		plateBytes := make([]byte, numLength)
		_, err = buf.Read(plateBytes)
		if err != nil {
			return nil, err
		}
		plateMsg.Plate = string(plateBytes)
		err = binary.Read(buf, binary.BigEndian, &plateMsg.Timestamp)
		if err != nil {
			return nil, err
		}
		return plateMsg, nil
	case WantHeartbeatMessageType:
		wantHeartbeatMsg := &WantHeartbeatMessage{}
		err := binary.Read(buf, binary.BigEndian, wantHeartbeatMsg)
		if err != nil {
			return nil, err
		}
		return wantHeartbeatMsg, nil
	case IAmCameraMessageType:
		iAmCameraMsg := &IAmCameraMessage{}
		err := binary.Read(buf, binary.BigEndian, iAmCameraMsg)
		if err != nil {
			return nil, err
		}
		return iAmCameraMsg, nil
	case IAmDispatcherMessageType:
		iAmDispatcherMsg := &IAmDispatcherMessage{
			MessageType: IAmDispatcherMessageType,
			Roads:       make([]uint16, 0),
		}
		buf.ReadByte()
		numRoads, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}
		iAmDispatcherMsg.NumRoads = numRoads
		roads := make([]uint16, numRoads)
		err = binary.Read(buf, binary.BigEndian, roads)
		if err != nil {
			return nil, err
		}
		iAmDispatcherMsg.Roads = roads
		return iAmDispatcherMsg, nil
	default:
		return nil, errors.New("unknown message type")
	}
}
