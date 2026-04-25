package server

import (
	"net"
)

type BaseServer struct {
	HandleConnectionFunc func(conn net.Conn)
}

func (s *BaseServer) Start(port string) error {
	ln, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go s.HandleConnection(conn)
	}
}

func (s *BaseServer) HandleConnection(conn net.Conn) {
	if s.HandleConnectionFunc != nil {
		s.HandleConnectionFunc(conn)
	}
	conn.Close()
}
