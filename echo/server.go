package echo

import (
	"io"
	"log"
	"net"

	"github.com/BhavyaMuni/protohackers/server"
)

type EchoServer struct {
	server.BaseServer
}

func NewEchoServer() *EchoServer {
	es := &EchoServer{}
	es.HandleConnectionFunc = es.handleConnection
	return es
}

func (es EchoServer) handleConnection(conn net.Conn) {
	log.Println("Connected with...")
	log.Println(conn.RemoteAddr())
	io.Copy(conn, conn)
}
