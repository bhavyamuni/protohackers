package lineReversal

import (
	"fmt"
	"net"
)

type LineReversalServer struct {
	sessions map[string]*LCRPSession
}

func NewLineReversalServer() *LineReversalServer {
	return &LineReversalServer{
		sessions: make(map[string]*LCRPSession),
	}
}

func (s *LineReversalServer) Start(port string) error {
	// udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("fly-global-services%s", port))
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1%s", port))
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
	key := fmt.Sprintf("%s-%d", addr, message.Session)
	sesh := NewSession(message.Session, addr, conn)
	s.sessions[key] = sesh
	sesh.incoming <- message
}

func (s *LineReversalServer) closeMessage(message *CloseMessage, conn *net.UDPConn, addr *net.UDPAddr) {
	key := fmt.Sprintf("%s-%d", addr, message.Session)
	sesh := s.sessions[key]
	sesh.incoming <- message
}

func (s *LineReversalServer) dataMessage(message *DataMessage, conn *net.UDPConn, addr *net.UDPAddr) {
	key := fmt.Sprintf("%s-%d", addr, message.Session)
	sesh, exists := s.sessions[key]
	if !exists {
		close := &CloseMessage{
			Session: message.Session,
		}
		SendMessage(close, conn, addr)
		return
	}
	sesh.incoming <- message
}

func (s *LineReversalServer) ackMessage(message *AckMessage, conn *net.UDPConn, addr *net.UDPAddr) {
	key := fmt.Sprintf("%s-%d", addr, message.Session)
	sesh, exists := s.sessions[key]
	if !exists {
		close := &CloseMessage{
			Session: message.Session,
		}
		SendMessage(close, conn, addr)
		return
	}
	sesh.incoming <- message
}

func CloseSession(session int64, conn *net.UDPConn, addr *net.UDPAddr) {
	closeMessage := &CloseMessage{
		Session: session,
	}
	closeMessageString := closeMessage.String()
	conn.WriteToUDP([]byte(closeMessageString), addr)
}

func SendMessage(m Message, conn *net.UDPConn, addr *net.UDPAddr) error {
	_, err := conn.WriteToUDP([]byte(m.String()), addr)
	if err != nil {
		return fmt.Errorf("Error writing message %s to conn %s", m.String(), addr)
	}
	return nil
}
