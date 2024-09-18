package main

import (
	"flag"
	"log"
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
		listener, err := server.InitServer(ip, "4000")
		if err != nil {
			log.Fatalf("Error initializing server: %v", err)
		}
		defer listener.Close()

		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Error accepting connection: %v", err)
				continue
			}
			go server.HandleServerConn(conn)
		}

	case "CLIENT":
		// Ensure -ip and -p are provided for CLIENT mode
		if *serverIp == "" {
			log.Fatal("Server IP is required for CLIENT mode")
		}
		if *port == "" {
			log.Fatal("Port is required for CLIENT mode")
		}
		server.HandleClientConn(serverIp, port)
	default:
		log.Fatalf("Unknown value for -h flag, use CLIENT or HOST as the flag value")
	}
}
