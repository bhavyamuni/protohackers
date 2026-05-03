package linereversal

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// type Session interface{}

type LCRPSession struct {
	id             int64
	dataSentBuffer string
	dataSent       int64
	dataAcked      int64
	dataRecvd      int64
	addr           *net.UDPAddr
	conn           *net.UDPConn
	dataRecvBuffer chan string
	incoming       chan Message
	outgoing       chan Message
}

func NewSession(id int64, addr *net.UDPAddr, conn *net.UDPConn) *LCRPSession {
	s := &LCRPSession{
		id:             id,
		addr:           addr,
		conn:           conn,
		dataRecvBuffer: make(chan string),
		dataSentBuffer: "",
		dataSent:       0,
		dataAcked:      0,
		dataRecvd:      0,
		incoming:       make(chan Message),
		outgoing:       make(chan Message),
	}
	go s.run()
	go s.send()
	go s.application()
	return s
}

func (s *LCRPSession) send() {
	timer := time.NewTimer(60 * time.Second)
	var lastMsg Message = nil
	for {
		select {
		case msg := <-s.outgoing:
			SendMessage(msg, s.conn, s.addr)
			lastMsg = msg
			timer.Reset(RetransmissionTimeout)
		case <-timer.C:
			SendMessage(lastMsg, s.conn, s.addr)
			timer.Reset(RetransmissionTimeout)
		}
	}
}
func (s *LCRPSession) application() {
	currBuf := ""
	for data := range s.dataRecvBuffer {
		for i := strings.Index(data, "\n"); i != -1; {
			dataToSend := ReverseAllLines(currBuf + data[:i+1])
			dataMsg := &DataMessage{
				Data:    dataToSend,
				Pos:     s.dataSent,
				Session: s.id,
			}

			s.dataSent += int64(len(dataToSend))
			s.dataSentBuffer += dataToSend
			s.outgoing <- dataMsg
			currBuf = ""
			data = data[i+1:]
			i = strings.Index(data, "\n")
		}
		currBuf += string(data)
	}
}

func (s *LCRPSession) run() {
	for msg := range s.incoming {
		fmt.Printf("Handling message: %q from %v\n", msg.String(), s.addr)
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

	if m.Length == s.dataSent {
		s.dataAcked += m.Length
		// s.dataSentBuffer = ""
	} else if m.Length > s.dataSent {
		close := &CloseMessage{Session: s.id}
		s.outgoing <- close
	} else if m.Length < s.dataSent {
		s.dataAcked += m.Length
		retransMsg := &DataMessage{
			Session: s.id,
			Pos:     m.Length,
			Data:    s.dataSentBuffer[m.Length:],
		}
		s.outgoing <- retransMsg
	}
}

func (s *LCRPSession) handleData(m *DataMessage) {
	ackMsg := &AckMessage{
		Session: s.id,
		Length:  int64(s.dataRecvd),
	}

	if m.Pos != s.dataRecvd {
		s.outgoing <- ackMsg
		return
	}

	sz := int64(len(m.Data))
	ackMsg.Length = sz
	s.dataRecvd += sz
	s.outgoing <- ackMsg
	s.dataRecvBuffer <- m.Data
}
