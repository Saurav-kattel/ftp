package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
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

		conn, err := net.Dial("tcp", *serverIp+":"+*port)
		if err != nil {
			log.Fatalf("Failed establishing the connection: %v", err)
		}
		defer conn.Close()

		var wg sync.WaitGroup
		wg.Add(2)

		// Goroutine to handle server messages
		go func() {
			defer wg.Done()
			buffer := make([]byte, 1024)
			for {
				n, err := conn.Read(buffer)
				if err != nil {
					if err == io.EOF {
						log.Panic("Connection closed by server")
						break
					}
					fmt.Printf("Error reading from server: %v\n", err)
					continue
				}
				fmt.Printf("\n<< %v", string(buffer[:n]))
			}
		}()

		go func() {
			defer wg.Done()
			inputReader := bufio.NewReader(os.Stdin)
			fmt.Println("say hi")
			for {
				fmt.Printf(">>")
				userInput, _ := inputReader.ReadString('\n')

				// Write user input to server
				_, err := conn.Write([]byte(userInput))
				if err != nil {
					fmt.Printf("Error writing to server: %v\n", err)
					break
				}
			}
		}()
		// Wait for the goroutine to finish when the connection closes
		wg.Wait()
		fmt.Println("Client connection closed")
	default:
		log.Fatalf("Unknown value for -h flag, use CLIENT or HOST as the flag value")
	}
}
