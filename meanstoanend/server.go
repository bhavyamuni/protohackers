package meanstoanend

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"

	"github.com/BhavyaMuni/protohackers/server"
)

type MeansToAnEndServer struct {
	server.BaseServer
}

type meansToAnEndMessage struct {
	Type byte
	P1   int32
	P2   int32
}

func NewMeansToAnEndServer() *MeansToAnEndServer {
	s := &MeansToAnEndServer{}
	s.HandleConnectionFunc = s.handleConnection
	return s
}

func findMean(queries map[int32]int32, minTime int32, maxTime int32) int32 {
	var sum int64
	var cnt int32
	for i := minTime; i <= maxTime; i++ {
		if v, ok := queries[i]; ok {
			sum += int64(v)
			cnt++
		}
	}
	if cnt == 0 {
		return 0
	}
	return int32(sum / int64(cnt))
}

func (MeansToAnEndServer) handleConnection(conn net.Conn) {
	log.Println("Connected with...")
	log.Println(conn.RemoteAddr())
	queries := make(map[int32]int32)
	for {
		message := meansToAnEndMessage{}
		err := binary.Read(conn, binary.BigEndian, &message)
		if err != nil {
			fmt.Println("binary.Read failed:", err)
			return
		}

		if message.Type == byte('I') {
			log.Println("Insert message received")
			queries[message.P1] = message.P2
		} else if message.Type == byte('Q') {
			log.Println("Query message received")
			mean := findMean(queries, message.P1, message.P2)
			binary.Write(conn, binary.BigEndian, mean)
		} else {
			binary.Write(conn, binary.BigEndian, "Invalid message received")
			log.Println("Invalid message received")
		}

		log.Printf("Received message: %v", message)
	}
}
