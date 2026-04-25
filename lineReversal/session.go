package lineReversal

import (
	"net"
)

// type Session interface{}

type LCRPSession struct {
	id       int64
	dataSent string
	pos      int64
	addr     *net.UDPAddr
	conn     *net.UDPConn
	incoming chan Message
}

func NewSession(id int64, addr *net.UDPAddr, conn *net.UDPConn) *LCRPSession {
	s := &LCRPSession{
		id:       id,
		pos:      0,
		addr:     addr,
		conn:     conn,
		dataSent: "",
		incoming: make(chan Message, 16),
	}
	go s.run()
	return s
}

func (s *LCRPSession) run() {
	for msg := range s.incoming {
		switch msg := msg.(type) {
		case *ConnectMessage:
			s.handleConnect(msg)
		case *CloseMessage:
			s.handleClose(msg)
		case *DataMessage:
			s.handleData(msg)
		case *AckMessage:
			s.handleAck(msg)
		}
	}
}

func (s *LCRPSession) handleConnect(_ *ConnectMessage) {
	println("Connect message")
}

func (s *LCRPSession) handleClose(_ *CloseMessage) {
	println("Close message")
}

func (s *LCRPSession) handleAck(_ *AckMessage) {
	println("ACK message")
}

func (s *LCRPSession) handleData(m *DataMessage) {
	ack := &AckMessage{
		Session: s.id,
		Length:  int64(s.pos),
	}

	if m.Pos != s.pos {
		SendMessage(ack, s.conn, s.addr)
		return
	}

	if m.Pos == s.pos {
		sz := int64(len(m.Data))
		ack.Length = sz
		dataToSend := ReverseAllLines(m.Data)
		dataMsg := &DataMessage{
			Data:    dataToSend,
			Pos:     sz,
			Session: s.id,
		}

		s.pos += sz
		s.dataSent += dataToSend
		SendMessage(dataMsg, s.conn, s.addr)
	}
}
