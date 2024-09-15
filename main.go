package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"sauravkattel/ftp/server"
	"sync"
)

var isConnectionEstablished bool
var mu sync.Mutex

func main() {
	// Define flags
	stateFlag := flag.String("h", "HOST", "-h [CLIENT | HOST] \n CLIENT to connect to the host server \n HOST to init a server\n")
	serverIp := flag.String("ip", "", "-ip [Server's IP] (for CLIENT mode only)")

	port := flag.String("p", "", "-p [PORT] (for CLIENT mode only)")

	// Parse the command-line flags
	flag.Parse()

	switch *stateFlag {
	case "HOST":
		// Get the IP address for the HOST mode
		ip, err := server.GetIp()
		if err != nil {
			log.Fatalf("Error getting IP: %v", err)
		}
		fmt.Println("ip is ", ip)
		listener, err := server.InitServer(ip, "4000")
		if err != nil {
			log.Fatalf("Error initializing server: %v", err)
		}
		defer listener.Close()

		isConnectionEstablished = false
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Error accepting connection: %v", err)
				continue
			}
			go server.HandleConn(conn)
		}

	case "CLIENT":
		// Ensure -ip and -p are provided for CLIENT mode
		if *serverIp == "" {
			log.Fatal("Server IP is required for CLIENT mode")
		}
		if *port == "" {
			log.Fatal("Port is required for CLIENT mode")
		}
		fmt.Println("server ip is ", *serverIp)
		conn, err := net.Dial("tcp", *serverIp+":"+*port)
		defer conn.Close()
		if err != nil {
			log.Fatalf("Failed establishing the connection: %v", err)
		}

		_, err = conn.Write([]byte("Hello server, I am a client"))
		if err != nil {
			log.Fatalf("Error sending message: %v", err)
		}
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			wg.Done()
			buffer := make([]byte, 1024)
			for {
				// Read data from the connection
				n, err := conn.Read(buffer)
				if err != nil {
					fmt.Printf("Error reading from server: %v\n", err)
					break
				}
				// Print the received message
				fmt.Println("Message from server:", string(buffer[:n]))
			}
		}()
		wg.Wait()

	default:
		log.Fatalf("Unknown value for -h flag, use CLIENT or HOST as the flag value")
	}
}
