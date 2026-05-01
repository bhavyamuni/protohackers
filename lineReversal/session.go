package linereversal

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

func (s *LCRPSession) handleClose(m *CloseMessage) {
	SendMessage(m, s.conn, s.addr)
}

func (s *LCRPSession) handleAck(m *AckMessage) {
	if m.Length > int64(len(s.dataSent)) {
		close := &CloseMessage{Session: s.id}
		SendMessage(close, s.conn, s.addr)
	}

	if m.Length < int64(len(s.dataSent)) {
		retransMsg := &DataMessage{
			Session: s.id,
			Pos:     m.Length,
			Data:    s.dataSent[m.Length:],
		}
		SendMessage(retransMsg, s.conn, s.addr)
	}
}

func (s *LCRPSession) handleData(m *DataMessage) {
	ackMsg := &AckMessage{
		Session: s.id,
		Length:  int64(s.pos),
	}

	if m.Pos != s.pos {
		SendMessage(ackMsg, s.conn, s.addr)
		return
	}

	sz := int64(len(m.Data))
	ackMsg.Length = sz
	dataToSend := ReverseAllLines(m.Data)
	dataMsg := &DataMessage{
		Data:    dataToSend,
		Pos:     sz,
		Session: s.id,
	}

	s.pos += sz
	s.dataSent += dataToSend
	SendMessage(ackMsg, s.conn, s.addr)
	SendMessage(dataMsg, s.conn, s.addr)
}
