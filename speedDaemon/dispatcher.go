package speeddaemon

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
)

type IAmDispatcherMessage struct {
	MessageType
	NumRoads uint8
	Roads    []uint16
}

type Dispatcher struct {
	NumRoads uint8
	Roads    []uint16
	Conn     *net.Conn
}

type TicketMessage struct {
	MessageType
	Plate      string
	Road       uint16
	Mile1      uint16
	Timestamp1 uint32
	Mile2      uint16
	Timestamp2 uint32
	Speed      uint16
}

type Ticket struct {
	Plate      string
	Road       uint16
	Mile1      uint16
	Timestamp1 uint32
	Mile2      uint16
	Timestamp2 uint32
	Speed      uint16
}

func (m *IAmDispatcherMessage) Handle(s *SpeedDaemonServer, conn *net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.dispatchers[conn]; ok {
		s.SendError(conn, "Dispatcher already registered")
		return
	}
	if _, ok := s.cameras[conn]; ok {
		s.SendError(conn, "Camera already registered")
		return
	}

	dispatcher := Dispatcher{NumRoads: m.NumRoads, Roads: m.Roads, Conn: conn}
	s.dispatchers[conn] = dispatcher
	for _, road := range m.Roads {
		if _, ok := s.tickets[road]; !ok {
			s.tickets[road] = make(chan *Ticket)
		}
		go dispatcher.MonitorTicketQueue(s.tickets[road], &s.ticketDays)
	}
}

func (d *Dispatcher) SendTicket(ticket Ticket) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, TicketMessageType)
	binary.Write(buf, binary.BigEndian, uint8(len(ticket.Plate)))
	buf.WriteString(ticket.Plate)
	binary.Write(buf, binary.BigEndian, ticket.Road)
	binary.Write(buf, binary.BigEndian, ticket.Mile1)
	binary.Write(buf, binary.BigEndian, ticket.Timestamp1)
	binary.Write(buf, binary.BigEndian, ticket.Mile2)
	binary.Write(buf, binary.BigEndian, ticket.Timestamp2)
	binary.Write(buf, binary.BigEndian, ticket.Speed)
	binary.Write(*d.Conn, binary.BigEndian, buf.Bytes())
	log.Println("Sent ticket: ", ticket)
}

func (d *Dispatcher) MonitorTicketQueue(tickets <-chan *Ticket, ticketDays *map[uint32]map[string]bool) {
	for ticket := range tickets {
		day1 := ticket.Timestamp1 / 86400
		day2 := ticket.Timestamp2 / 86400
		if _, ok := (*ticketDays)[day1]; !ok {
			(*ticketDays)[day1] = make(map[string]bool)
		}
		if _, ok := (*ticketDays)[day2]; !ok {
			(*ticketDays)[day2] = make(map[string]bool)
		}
		_, d1ok := (*ticketDays)[day1][ticket.Plate]
		_, d2ok := (*ticketDays)[day2][ticket.Plate]
		if !d1ok && !d2ok {
			(*ticketDays)[day1][ticket.Plate] = true
			(*ticketDays)[day2][ticket.Plate] = true
			go d.SendTicket(*ticket)
		}
	}
}
