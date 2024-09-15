package server

import (
	"bufio"
	"fmt"
	"net"
	"os"
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

// HandleConn manages the connection and keeps it open
func HandleConn(conn net.Conn) {
	defer conn.Close()

	fmt.Println("Connected with", conn.RemoteAddr())

	// Use a buffered reader to read commands from stdin
	reader := bufio.NewReader(conn)
	for {
		// Read the command from stdin
		fmt.Print("Enter the command: ")
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

		// Read response from client
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading response: %v\n", err)
			return
		}
		fmt.Println("Response from client:", response)
	}
}
