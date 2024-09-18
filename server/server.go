package server

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sauravkattel/ftp/util"
	"sync"
)

const (
	BUFFER_SIZE = 1024
)

type DataStruct struct {
	CmdName   string
	FlagCount int
	Flags     map[string]string
}

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

func WriteToClient(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		// Read the command from stdin
		fmt.Print(">>")
		cmd, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading command: %v\n", err)
			return
		}

		parsedInput := util.ParseUserInput(cmd)
		// Send the command to the client
		data := util.ConvertIntoBytes(parsedInput)
		util.WriteBytes(conn, data)

	}
}

func execCmd(cmd util.DataStruct) {
	switch cmd.CmdName {
	case "SEND", "send":

	}
}

func ReadFromClient(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	// Use a buffered reader to read commands from stdin
	// Read response from client

	for {
		response := util.ReadBytes(conn)
		fmt.Println(string(response))
	}
}

// HandleConn manages the connection and keeps it open
func HandleServerConn(conn net.Conn) {
	defer conn.Close()
	var wg sync.WaitGroup
	wg.Add(2)

	go ReadFromClient(conn, &wg)
	go WriteToClient(conn, &wg)
	wg.Wait()
	fmt.Println("Connected with", conn.RemoteAddr())
}

func HandleClientConn(serverIp, port *string) {
	conn, err := net.Dial("tcp", *serverIp+":"+*port)
	if err != nil {
		log.Fatalf("Failed establishing the connection: %v", err)
	}

	defer conn.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine to handle server messages
	go ReadFromHost(conn, &wg)
	go WriteToHost(conn, &wg)
	// Wait for the goroutine to finish when the connection closes
	wg.Wait()
	fmt.Println("Client connection closed")

}

func WriteToHost(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	inputReader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf(">>")
		userInput, _ := inputReader.ReadString('\n')
		parsedInput := util.ParseUserInput(userInput)

		ftLen := make([]byte, 4)
		fnLen := make([]byte, 4)
		actLen := make([]byte, 4)
		filePath := ""
		dataBuffer := make([]byte, BUFFER_SIZE)

		switch parsedInput.CmdName {
		case "SEND":
			{
				for key, value := range parsedInput.Flags {
					switch key {
					case "t":
						{
							binary.BigEndian.PutUint32(ftLen, uint32(len(value)))
							dataBuffer = append(dataBuffer, ftLen...)
							dataBuffer = append(dataBuffer, []byte(value)...)
						}

					case "n":
						{
							binary.BigEndian.PutUint32(fnLen, uint32(len(value)))
							dataBuffer = append(dataBuffer, fnLen...)
							dataBuffer = append(dataBuffer, []byte(value)...)

						}
					case "p":
						{
							fileStat, err := os.Stat(value)
							if err != nil {
								fmt.Printf("could not read file: %+v\n", err)
								os.Exit(1)
							}

							fileUnit32Size := uint32(fileStat.Size())
							binary.BigEndian.AppendUint32(actLen, fileUnit32Size)
							dataBuffer = append(dataBuffer, actLen...)
							filePath = value
						}
					}

				}
			}
		}
		fmt.Println(filePath)
		file, err := os.Open(filePath)

		if err != nil {
			log.Fatalf("unable to open file %+v", err)
		}
		uintFtLen := binary.BigEndian.Uint32(ftLen)
		uintFnLen := binary.BigEndian.Uint32(fnLen)
		occupiedBufferSpace := uintFnLen + uintFtLen + 12 // 12 for sotring  file metadata
		availableSpace := BUFFER_SIZE - occupiedBufferSpace
		tmpBuff := make([]byte, availableSpace)
		for {
			bytesRead, err := file.Read(tmpBuff)
			if err != nil && err != io.EOF {
				fmt.Printf("failed reading file %+v", err)
				return
			}
			if bytesRead == 0 {
				break
			}
			util.WriteBytes(conn, tmpBuff)
			tmpBuff = make([]byte, BUFFER_SIZE)
		}
	}
}

func ReadFromHost(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		n := util.ReadBytes(conn)
		fmt.Printf("\n<< %v", util.ConvertIntoDataStruct(n))
	}
}
