package primetime

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net"
	"strconv"

	"github.com/BhavyaMuni/protohackers/server"
)

type primeTimeInt struct {
	big.Int
}

type primeTimeRequest struct {
	Method *string       `json:"method"`
	Number *primeTimeInt `json:"number"`
}

func (b *primeTimeInt) UnmarshalJSON(data []byte) error {
	var z big.Int
	if _, ok := z.SetString(string(data[:]), 10); ok {
		b.Int = z
	} else if _, err := strconv.ParseFloat(string(data[:]), 64); err == nil {
		b.Int = *big.NewInt(0)
	} else {
		return fmt.Errorf("Invalid number format")
	}
	return nil
}

type primeTimeResponse struct {
	Method string `json:"method"`
	Prime  bool   `json:"prime"`
}

type PrimeTimeServer struct {
	server.BaseServer
}

func (r *primeTimeRequest) validRequest() bool {
	return r.Method != nil && *r.Method == "isPrime" && r.Number != nil
}

func NewPrimeTimeServer() *PrimeTimeServer {
	s := &PrimeTimeServer{}
	s.HandleConnectionFunc = s.handleConnection
	return s
}

func isPrime(n *big.Int) bool {
	return n.ProbablyPrime(20)
}

func (PrimeTimeServer) handleConnection(conn net.Conn) {
	log.Println("Connected with...")
	log.Println(conn.RemoteAddr())

	connReader := bufio.NewReader(conn)
	connWriter := bufio.NewWriter(conn)

	for {
		line, err := connReader.ReadBytes('\n')
		if err != nil {
			break
		}
		req := primeTimeRequest{}
		err = json.Unmarshal(line, &req)
		if err != nil || !req.validRequest() {
			fmt.Println(req)
			connWriter.Write([]byte("💩\n"))
			log.Println("Invalid request, recieved: " + string(line))
			break
		}
		log.Println("Request: ", req)
		resp := primeTimeResponse{Method: "isPrime", Prime: isPrime(&req.Number.Int)}
		respBytes, err := json.Marshal(resp)
		respBytes = append(respBytes, '\n')
		if err != nil {
			break
		}
		connWriter.Write(respBytes)
		connWriter.Flush()
	}
}
