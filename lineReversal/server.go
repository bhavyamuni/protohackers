package lineReversal

import (
	"fmt"
	"net"

	"github.com/BhavyaMuni/protohackers/server"
)

type LineReversalServer struct {
	server.BaseServer
	connections            map[int64]net.Conn
	messagesSent           map[int64]string
	totalPayloadMaxPayload map[int64][2]int64
}

func NewLineReversalServer() *LineReversalServer {
	return &LineReversalServer{
		connections:            make(map[int64]net.Conn),
		messagesSent:           make(map[int64]string),
		totalPayloadMaxPayload: make(map[int64][2]int64),
	}
}

func (s *LineReversalServer) Start(port string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("fly-global-services%s", port))
	// udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1%s", port))
	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		return err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Error listening on UDP:", err)
		return err
	}
	defer conn.Close()

	fmt.Println("UDP server listening on", udpAddr)

	buffer := make([]byte, 1024)

	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading UDP packet:", err)
			continue
		}
		message, err := ParseMessage(string(buffer[:n]))
		if err != nil {
			fmt.Println("Error parsing message:", err)
			continue
		}

		s.handleMessage(message, conn, addr)
	}
}

func (s *LineReversalServer) handleMessage(message Message, conn *net.UDPConn, addr *net.UDPAddr) {
	fmt.Println("Received message:", message, "from", addr)
	switch message := message.(type) {
	case *ConnectMessage:
		s.connectMessage(message, conn, addr)
	case *CloseMessage:
		s.closeMessage(message, conn, addr)
	case *DataMessage:
		s.dataMessage(message, conn, addr)
	case *AckMessage:
		s.ackMessage(message, conn, addr)
	}
}

func (s *LineReversalServer) connectMessage(message *ConnectMessage, conn *net.UDPConn, addr *net.UDPAddr) {
	fmt.Println("Received connect message:", message)
	s.connections[message.Session] = conn
	ackMessage := &AckMessage{
		Session: message.Session,
		Length:  0,
	}
	ackMessageString := ackMessage.String()
	conn.WriteToUDP([]byte(ackMessageString), addr)
}

func (s *LineReversalServer) closeMessage(message *CloseMessage, conn *net.UDPConn, addr *net.UDPAddr) {
	fmt.Println("Received close message:", message)
	CloseSession(message.Session, conn, addr)
	delete(s.connections, message.Session)
}

func (s *LineReversalServer) dataMessage(message *DataMessage, conn *net.UDPConn, addr *net.UDPAddr) {
	fmt.Println("Received data message:", message)
	data := message.Data

	ackMessage := &AckMessage{
		Session: message.Session,
		Length:  int64(len(data)),
	}

	ackMessageString := ackMessage.String()
	conn.WriteToUDP([]byte(ackMessageString+"\n"), addr)

	reversedData := ReverseAllLines(data)
	messageSent := SendDataMessage(message.Session, reversedData, conn, addr)
	s.messagesSent[message.Session] += messageSent
}

func (s *LineReversalServer) ackMessage(message *AckMessage, conn *net.UDPConn, addr *net.UDPAddr) {
	fmt.Println("Received ack message:", message)
	if _, ok := s.connections[message.Session]; !ok {
		fmt.Println("Connection not found for session:", message.Session)
		CloseSession(message.Session, conn, addr)
		return
	}

	totalPayload := s.totalPayloadMaxPayload[message.Session][0]
	if message.Length > totalPayload {
		CloseSession(message.Session, conn, addr)
	}
	if message.Length < totalPayload {
		// get the messages starting from message.length and send them back
		messageSent, ok := s.messagesSent[message.Session]
		if !ok {
			fmt.Println("No messages sent for session:", message.Session)
			return
		}
		messageSent = messageSent[message.Length:]
		SendDataMessage(message.Session, messageSent, conn, addr)

	}
}

func CloseSession(session int64, conn *net.UDPConn, addr *net.UDPAddr) {
	closeMessage := &CloseMessage{
		Session: session,
	}
	closeMessageString := closeMessage.String()
	conn.WriteToUDP([]byte(closeMessageString), addr)
}

func SendDataMessage(session int64, data string, conn *net.UDPConn, addr *net.UDPAddr) string {
	dataMessage := &DataMessage{
		Session: session,
		Data:    data,
	}
	dataMessageString := dataMessage.String()
	_, err := conn.WriteToUDP([]byte(dataMessageString+"\n"), addr)
	fmt.Println("Sending data message ", dataMessageString, addr)
	if err != nil {
		fmt.Println("Error sending data message:", err)
		return ""
	}
	return dataMessageString
}
