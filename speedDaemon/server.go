package speeddaemon

import (
	"bufio"
	"io"
	"log"
	"net"
	"sync"

	"github.com/BhavyaMuni/protohackers/server"
)

type SpeedDaemonServer struct {
	server.BaseServer
	mu           sync.Mutex
	observations map[string][]Observation
	cameras      map[*net.Conn]Camera
	dispatchers  map[*net.Conn]Dispatcher
	tickets      map[uint16]chan *Ticket
	heartbeats   map[*net.Conn]bool
	ticketDays   map[uint32]map[string]bool
}

func NewSpeedDaemonServer() *SpeedDaemonServer {
	ssd := &SpeedDaemonServer{
		observations: make(map[string][]Observation),
		cameras:      make(map[*net.Conn]Camera),
		dispatchers:  make(map[*net.Conn]Dispatcher),
		tickets:      make(map[uint16]chan *Ticket),
		heartbeats:   make(map[*net.Conn]bool),
		ticketDays:   make(map[uint32]map[string]bool),
	}
	ssd.HandleConnectionFunc = ssd.handleConnection
	return ssd
}

func (ssd *SpeedDaemonServer) handleConnection(conn net.Conn) {
	buf := bufio.NewReader(conn)
	defer ssd.disconnectClient(&conn)
	for {
		message, err := ParseMessage(buf)
		if err != nil {
			if err != io.EOF {
				log.Println("Error reading from client: ", err)
				ssd.SendError(&conn, "Error reading from client")
			}
			break
		}
		log.Println("Received message: ", message, "from", conn.RemoteAddr(), "conn", &conn)
		message.Handle(ssd, &conn)
	}
}

func (ssd *SpeedDaemonServer) disconnectClient(conn *net.Conn) {
	ssd.mu.Lock()
	defer ssd.mu.Unlock()
	defer (*conn).Close()

	delete(ssd.cameras, conn)

	delete(ssd.dispatchers, conn)

	delete(ssd.heartbeats, conn)

	log.Println("Disconnected client: ", (*conn).RemoteAddr())

}
