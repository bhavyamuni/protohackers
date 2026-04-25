package budgetchat

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"

	"github.com/BhavyaMuni/protohackers/server"
)

type BudgetChatServer struct {
	server.BaseServer
	Users []user
}

type user struct {
	connection *net.Conn
	username   string
}

func NewBudgetChatServer() *BudgetChatServer {
	bcs := &BudgetChatServer{}
	bcs.HandleConnectionFunc = bcs.handleConnection
	bcs.Users = make([]user, 0)
	return bcs
}

func (s *BudgetChatServer) handleConnection(conn net.Conn) {
	log.Println("Connected with...")
	log.Println(conn.RemoteAddr())
	scanner := bufio.NewScanner(conn)
	conn.Write([]byte("Welcome to budgetchat! What shall I call you?\n"))
	scanner.Scan()
	name := scanner.Text()
	if !validUsername(name) {
		conn.Write([]byte("Invalid username. Please try again.\n"))
		conn.Close()
		return
	}
	currUsers := s.listUsers()
	s.Users = append(s.Users, user{connection: &conn, username: name})
	conn.Write([]byte("* The room contains: " + strings.Join(currUsers, ", ") + "\n"))
	s.broadcast("* "+name+" has entered the room\n", &conn)

	for scanner.Scan() {
		message := scanner.Text()
		s.broadcast(fmt.Sprintf("[%s] %s\n", name, message), &conn)
	}

	defer s.userLeft(&conn)
}

func (s *BudgetChatServer) broadcast(message string, currConn *net.Conn) {
	for _, u := range s.Users {
		if u.connection != currConn {
			(*u.connection).Write([]byte(message))
		}
	}
}

func (s *BudgetChatServer) listUsers() []string {
	var users []string
	for _, u := range s.Users {
		users = append(users, u.username)
	}
	return users
}

func (s *BudgetChatServer) userLeft(conn *net.Conn) {
	for i, u := range s.Users {
		if u.connection == conn {
			s.broadcast("* "+u.username+" has left the room\n", conn)
			s.Users = append(s.Users[:i], s.Users[i+1:]...)
			break
		}
	}
}

func validUsername(name string) bool {
	rx := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	return len(name) >= 1 && rx.MatchString(name)
}
