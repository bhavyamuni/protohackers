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
	retransTimer   *time.Timer
	expiryTimer    *time.Timer
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
		retransTimer:   time.NewTimer(RetransmissionTimeout),
		expiryTimer:    time.NewTimer(ExpiryTimeout),
	}
	go s.run()
	go s.send()
	go s.application()
	return s
}

func (s *LCRPSession) send() {
	var lastMsg Message = nil
	for {
		select {
		case msg := <-s.outgoing:
			//max 1000 bytes
			switch msg.(type) {
			case *DataMessage:

				lastMsg = msg

				s.retransTimer.Reset(RetransmissionTimeout)
			case *CloseMessage:
				s.retransTimer.Stop()
				s.expiryTimer.Stop()
			}
			SendMessage(msg, s.conn, s.addr)
		case <-s.retransTimer.C:
			// Retransmit message
			if lastMsg != nil {
				SendMessage(lastMsg, s.conn, s.addr)
				s.retransTimer.Reset(RetransmissionTimeout)
			} else {
				s.retransTimer.Stop()
			}
		}
	}
}
func (s *LCRPSession) application() {
	currBuf := &strings.Builder{}
	for data := range s.dataRecvBuffer {
		currBuf.WriteString(data)
		for i := strings.Index(currBuf.String(), "\n"); i != -1; i = strings.Index(currBuf.String(), "\n") {
			dataToSend := ReverseAllLines(currBuf.String()[:i+1])
			remainingData := currBuf.String()[i+1:]

			s.batchDataMsg(dataToSend, s.dataSent)
			s.dataSent += int64(len(dataToSend))
			s.dataSentBuffer += dataToSend
			currBuf.Reset()
			currBuf.WriteString(remainingData)
			i = strings.Index(currBuf.String(), "\n")
		}
	}
}

func (s *LCRPSession) run() {
	for {
		select {
		case msg := <-s.incoming:
			fmt.Printf("Handling message: %q from %v\n", msg.String(), s.addr)
			s.expiryTimer.Reset(ExpiryTimeout)
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
		case <-s.expiryTimer.C:
			s.expiryTimer.Stop()
			s.outgoing <- &CloseMessage{Session: s.id}
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
		s.retransTimer.Stop()
		// s.dataSentBuffer = ""
	} else if m.Length > s.dataSent {
		close := &CloseMessage{Session: s.id}
		s.outgoing <- close
	} else if m.Length < s.dataSent {
		s.dataAcked += m.Length
		s.batchDataMsg(s.dataSentBuffer[m.Length:], m.Length)
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
	s.dataRecvd += sz
	ackMsg.Length = s.dataRecvd
	s.outgoing <- ackMsg
	s.dataRecvBuffer <- m.Data
}

func (s *LCRPSession) batchDataMsg(msg string, pos int64) {
	if len(msg) > 950 {
		dataMsg := &DataMessage{
			Session: s.id,
			Pos:     pos,
			Data:    msg[:950],
		}
		s.outgoing <- dataMsg
		s.batchDataMsg(msg[950:], pos+950)
	} else {
		dataMsg := &DataMessage{
			Session: s.id,
			Pos:     pos,
			Data:    msg,
		}
		s.outgoing <- dataMsg
	}

}
