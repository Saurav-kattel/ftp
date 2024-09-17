package server

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sauravkattel/ftp/lexer"
	"sync"
)

type ServerType int

const (
	HOST ServerType = iota
	CLIENT
)

func GetIp() (string, error) {

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return " ", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String(), nil
		}
	}
	return "", fmt.Errorf("ip not found, try again!")
}

func InitServer(ip string, port string) (net.Listener, error) {
	conn, err := net.Listen("tcp", ip+":"+port)

	if err != nil {
		return nil, err
	}
	return conn, nil
}

func WriteClientResponse(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		// Read the command from stdin
		fmt.Print(">>")
		cmd, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading command: %v\n", err)
			return
		}
		// Send the command to the client
		_, err = conn.Write([]byte(cmd))
		if err != nil {
			fmt.Printf("Error sending command: %v\n", err)
			return
		}
	}
}

func ReadFromClient(conn net.Conn, wg *sync.WaitGroup) {
	reader := bufio.NewReader(conn)
	defer wg.Done()
	// Use a buffered reader to read commands from stdin
	// Read response from client

	for {

		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading response: %v\n", err)
			return
		}

		var Tokens []lexer.Token

		lex := lexer.Lexer{}
		lex.LoadLexer(response)
		tkn := lex.GetNextToken()
		for {
			if tkn.TokenType == lexer.EOF {
				Tokens = append(Tokens, tkn)
				break
			}
			Tokens = append(Tokens, tkn)
			tkn = lex.GetNextToken()
		}

	}
}

// HandleConn manages the connection and keeps it open
func HandleServerConn(conn net.Conn) {
	defer conn.Close()
	var wg sync.WaitGroup
	wg.Add(2)

	go ReadFromClient(conn, &wg)
	go WriteClientResponse(conn, &wg)
	wg.Wait()
	fmt.Println("Connected with", conn.RemoteAddr())
}
