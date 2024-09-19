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

func ConvertByteToUint32(data []byte) uint32 {
	length := binary.BigEndian.Uint32(data)
	return length
}

func ReadFromClient(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	fileName := ""
	var fileSize uint32
	// Use a bufferedreader to read commands from stdin
	// Read response from client
	readMetaData := true
	for {
		response := util.ReadBytes(conn)

		if readMetaData && len(response) > 0 {

			//extract the length of filename
			stIdx := 0
			endIdx := 4
			fileNameLenBuffer := response[stIdx:endIdx]
			fileNameLen := ConvertByteToUint32(fileNameLenBuffer)

			fmt.Println(fileNameLen, response)
			// extract the filename
			stIdx = endIdx
			endIdx = stIdx + int(fileNameLen)
			fileName = string(response[stIdx:endIdx])

			//extract the filesize
			stIdx = endIdx
			endIdx = stIdx + 4
			fileSizeBuffer := response[stIdx:endIdx]
			fileSize = ConvertByteToUint32(fileSizeBuffer)

			fmt.Println(fileSize, fileName)
			readMetaData = false
		} else if !readMetaData && len(response) > 0 {
			fname := "./fs" + fileName
			file, err := os.OpenFile(fname, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Println(err)
			}
			defer file.Close()
			file.Write(response)
			fmt.Println("[done]")

		}

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

		actLen := make([]byte, 4)
		fileLen := make([]byte, 4)
		filePath := ""
		dataBuffer := []byte{}
		switch parsedInput.CmdName {
		case "SEND":
			{
				for key, value := range parsedInput.Flags {
					switch key {
					case "p":
						{
							fileStat, err := os.Stat(value)
							if err != nil {
								fmt.Printf("could not read file: %+v\n", err)
								os.Exit(1)
							}

							// to store encode filename inside data buffer
							fileName := fileStat.Name()

							fileNameSize := uint32(len(fileName))
							binary.BigEndian.PutUint32(fileLen, fileNameSize)
							dataBuffer = append(dataBuffer, fileLen...)
							// encode filename
							dataBuffer = append(dataBuffer, []byte(fileName)...)
							// to encode filesize inside data buffer
							fileUnit32Size := uint32(fileStat.Size())
							binary.BigEndian.PutUint32(actLen, fileUnit32Size)

							dataBuffer = append(dataBuffer, actLen...)
							filePath = value
						}
					}

				}

				util.WriteBytes(conn, dataBuffer)

				file, err := os.Open(filePath)
				contentBuffer := make([]byte, 1024)

				if err != nil {
					log.Fatalf("unable to open file %+v", err)
				}

				for {
					bytesRead, err := file.Read(contentBuffer)
					if err != nil && err != io.EOF {
						fmt.Printf("failed reading file %+v", err)
						return
					}
					if bytesRead == 0 {
						break
					}
					util.WriteBytes(conn, contentBuffer[:bytesRead])
				}

			}
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
