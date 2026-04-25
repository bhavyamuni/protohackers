package mobinthemiddle

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"regexp"
	"strings"

	"github.com/BhavyaMuni/protohackers/server"
)

const upstreamHost = "chat.protohackers.com"
const upstreamPort = "16963"
const tonyAddress = "7YWHMfk9JZe0LM0g1ZauHuiSxhI"

type MobInTheMiddleServer struct {
	server.BaseServer
}

func NewMobInTheMiddleServer() *MobInTheMiddleServer {
	ms := &MobInTheMiddleServer{}
	ms.HandleConnectionFunc = ms.handleConnection
	return ms
}

func (ms MobInTheMiddleServer) handleConnection(conn net.Conn) {
	log.Println("Connected with client:", conn.RemoteAddr())

	upstreamConn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", upstreamHost, upstreamPort))
	if err != nil {
		log.Println("Error connecting to upstream server:", err)
		return
	}
	defer upstreamConn.Close()
	log.Println("Connected to upstream server:", upstreamConn.RemoteAddr())

	go handleUpstream(upstreamConn, conn)

	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Println("Error reading from client:", err)
			}
			break
		}
		message = strings.TrimSuffix(message, "\n")
		_, err = upstreamConn.Write([]byte(modifyMessage(message) + "\n"))
		if err != nil {
			log.Println("Error writing to upstream:", err)
			return
		}
	}
}

func handleUpstream(upstreamConn net.Conn, downstreamConn net.Conn) {
	upstreamScanner := bufio.NewScanner(upstreamConn)
	for upstreamScanner.Scan() {
		message := upstreamScanner.Text()
		_, err := downstreamConn.Write([]byte(modifyMessage(message) + "\n"))
		if err != nil {
			log.Println("Error writing to client:", err)
			return
		}
	}
}

func modifyMessage(message string) string {
	log.Println("Modifying Message:", message)
	rx := regexp.MustCompile(`^7[a-zA-Z0-9]{25,34}$`)

	var actualMatches []string
	for _, word := range strings.Split(message, " ") {
		if rx.MatchString(word) {
			actualMatches = append(actualMatches, word)
		}
	}
	for _, match := range actualMatches {
		message = strings.Replace(message, match, tonyAddress, 1)
	}
	return message
}
