package linereversal

import (
	"net"
)

// type Session interface{}

type LCRPSession struct {
	id        int64
	dataSent  string
	dataAcked int64
	pos       int64
	addr      *net.UDPAddr
	conn      *net.UDPConn
	incoming  chan Message
	outgoing  chan Message
}

func NewSession(id int64, addr *net.UDPAddr, conn *net.UDPConn) *LCRPSession {
	s := &LCRPSession{
		id:        id,
		pos:       0,
		addr:      addr,
		conn:      conn,
		dataSent:  "",
		dataAcked: 0,
		incoming:  make(chan Message, 16),
		outgoing:  make(chan Message),
	}
	go s.run()
	go s.send()
	return s
}

func (s *LCRPSession) send() {
	for msg := range s.outgoing {
		SendMessage(msg, s.conn, s.addr)
	}
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
	ackMsg := &AckMessage{Session: s.id, Length: 0}
	s.outgoing <- ackMsg
}

func (s *LCRPSession) handleClose(m *CloseMessage) {
	s.outgoing <- m
}

func (s *LCRPSession) handleAck(m *AckMessage) {
	if m.Length < s.dataAcked {
		return
	}

	if m.Length == int64(len(s.dataSent)) {
		s.dataAcked += int64(len(s.dataSent))
		s.dataSent = ""
	}

	if m.Length > int64(len(s.dataSent)) {
		close := &CloseMessage{Session: s.id}
		s.outgoing <- close
	}

	if m.Length < int64(len(s.dataSent)) {
		retransMsg := &DataMessage{
			Session: s.id,
			Pos:     m.Length,
			Data:    s.dataSent[m.Length:],
		}
		s.outgoing <- retransMsg
	}
}

func (s *LCRPSession) handleData(m *DataMessage) {
	ackMsg := &AckMessage{
		Session: s.id,
		Length:  int64(s.pos),
	}

	if m.Pos != s.pos {
		s.outgoing <- ackMsg
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
	s.outgoing <- ackMsg
	s.outgoing <- dataMsg
}
