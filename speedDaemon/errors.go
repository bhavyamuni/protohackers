package speeddaemon

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
)

type ErrorMessage struct {
	MessageType
	Msg string
}

func (ssd *SpeedDaemonServer) SendError(conn *net.Conn, message string) {
	messageType := ErrorMessageType
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, messageType)
	binary.Write(buf, binary.BigEndian, uint8(len(message)))
	buf.WriteString(message)
	err := binary.Write(*conn, binary.BigEndian, buf.Bytes())
	if err != nil {
		log.Println("Error sending error: ", err)
		return
	}
	log.Println("Sent error: ", message, "to", (*conn).RemoteAddr())
}
