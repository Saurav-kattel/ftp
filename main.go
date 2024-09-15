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

	stateFlag := flag.String("h", "HOST", "-h [ CLIENT | HOST ] \n CLINET to connect to the host server \n HOST to init a server\n")
	serverIp := flag.String("ip", "null", "-ip [Server's Ip] (for CLIENT mode only)")
	port := flag.String("p", "null", "-p [PORT]  (for CLIENT mode only)")

	flag.Parse()

	if *stateFlag == "HOST" {
		ip, err := server.GetIp()

		if err != nil {
			fmt.Printf("error: ", err.Error())
		}
		conn, err := server.InitServer(ip, "4000")
		if err != nil {
			log.Panic("Error initalizing connection", err.Error())
		}
		defer conn.Close()
		isConnectionEstablished = false
		for {

			if isConnectionEstablished {
				break
			}

			con, err := conn.Accept()

			if err != nil {
				fmt.Println("Fuck")
			}

			isConnectionEstablished = true
			go server.HandleConn(con)
		}

	} else if *stateFlag == "CLIENT" {

		if *serverIp == "null" {
			log.Fatal("expected server's ip but got null")
		}

		if *port == "null" {
			log.Fatal("expected  port address but got null")
		}
		conn, err := net.Dial("tcp", *serverIp+":"+*port)
		if err != nil {
			log.Fatalf("failed establishing the connection %+v", err.Error())
		}
		conn.Write([]byte("Hello server I'am Client"))
		defer conn.Close()
	} else {
		log.Panic("unknown value for -h flag, use CLINET or HOST as the flag value")
	}
}
