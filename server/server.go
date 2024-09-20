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
		ParseCmd(conn, parsedInput)
	}
}

func ConvertByteToUint32(data []byte) uint32 {
	length := binary.BigEndian.Uint32(data)
	return length
}

func handleDataWrite(conn net.Conn, fileName string, fileSize uint32, previousResSize *int, response []byte) {

	_ = conn
	fname := "./fs" + fileName
	file, err := os.OpenFile(fname, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	progress := (len(response) + (*previousResSize)) * 100 / int(fileSize)
	str := fmt.Sprintf("Done %d%%", progress)
	fmt.Println(str)
	*previousResSize += len(response)
	fmt.Println()
	if err != nil {
		fmt.Println(err)
	}

	defer file.Close()
	file.Write(response)
}

func handleHelp() {
	type Cmd struct {
		Name  string
		Desc  string
		Usage string
		Flags map[string]string
	}

	commands := []Cmd{
		{
			Name:  "HELP",
			Desc:  "list all available commands",
			Usage: "HELP ",
		},
		{
			Name:  "SEND",
			Desc:  "send's file over a network",
			Usage: "SEND [flags]",
			Flags: map[string]string{
				"-p": "Path of the file",
			},
		}, {
			Name:  "LIST",
			Desc:  "List the file and directory of the current dir",
			Usage: "LIST ",
		}, {
			Name:  "CWD",
			Desc:  "Changes the current dir to the provided path's dir",
			Usage: "CWD [Flags]",
			Flags: map[string]string{
				"-p": "Path to the directory",
			},
		}, {
			Name:  "REN",
			Desc:  "Rename the file or directory",
			Usage: "REN [Flag] [Flag]",
			Flags: map[string]string{
				"-f": "Name of the file to rename",
				"-n": "New name",
			},
		}, {
			Name:  "DEL",
			Desc:  "Removes file or empty directory",
			Usage: "DEL [FLAG]",
			Flags: map[string]string{
				"-p": "Path to the file or directory",
			},
		},
	}

	for _, cmds := range commands {
		fmt.Printf("Name: %s\t Usage:%s\t", cmds.Name, cmds.Usage)
		fmt.Printf("Desc: %s\t", cmds.Desc)
		if len(cmds.Flags) > 0 {
			fmt.Printf("\n\t\tFlags: \n")
			for key, value := range cmds.Flags {
				fmt.Printf("\t\t%s : %s \n", key, value)
			}
		}
		fmt.Println()
	}
}

func ReadFromClient(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	fileName := ""
	cmdName := ""

	previousResSize := 0
	var fileSize uint32
	// Use a bufferedreader to read commands from stdin
	// Read response from client
	readMetaData := true
	for {
		response, err := util.ReadBytes(conn)
		if err != nil {
			fmt.Println(err)
			return
		}
		if readMetaData && len(response) > 0 {

			//extract the length of filename
			stIdx := 0
			endIdx := 4

			cmdNameLenBuffer := response[stIdx:endIdx]
			cmdNameLen := ConvertByteToUint32(cmdNameLenBuffer)

			stIdx = endIdx
			endIdx = stIdx + int(cmdNameLen)

			cmdName = string(response[stIdx:endIdx])
			stIdx = endIdx
			endIdx = stIdx + 4

			fileNameLenBuffer := response[stIdx:endIdx]
			fileNameLen := ConvertByteToUint32(fileNameLenBuffer)

			// extract the filename
			stIdx = endIdx
			endIdx = stIdx + int(fileNameLen)
			fileName = string(response[stIdx:endIdx])

			//extract the filesize
			stIdx = endIdx
			endIdx = stIdx + 4
			fileSizeBuffer := response[stIdx:endIdx]
			fileSize = ConvertByteToUint32(fileSizeBuffer)
			readMetaData = false
		} else if !readMetaData && len(response) > 0 {
			switch cmdName {
			case "SEND":
				handleDataWrite(conn, fileName, fileSize, &previousResSize, response)
				break

			}
		} else {
			previousResSize = 0
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

func ParseCmd(conn net.Conn, parsedInput util.DataStruct) {

	actLen := make([]byte, 4)
	fileLen := make([]byte, 4)
	cmdNameLenBuffer := make([]byte, 4)
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
							return

						}

						// to store the cmd name
						cmdNameLen := len(parsedInput.CmdName)
						binary.BigEndian.PutUint32(cmdNameLenBuffer, uint32(cmdNameLen))
						dataBuffer = append(dataBuffer, cmdNameLenBuffer...)
						dataBuffer = append(dataBuffer, []byte(parsedInput.CmdName)...)

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
					return
				}
				if bytesRead == 0 {
					break
				}
				util.WriteBytes(conn, contentBuffer[:bytesRead])
			}

		}
	case "HELP":
		handleHelp()
	case "LIST":
		handleList()
	case "CWD":
		handleChw(parsedInput)
	case "REN":
		handleRename(parsedInput)
	case "DEL":
		handleDelete(parsedInput)
	}
}

func handleDelete(parsedInput util.DataStruct) {
	if parsedInput.FlagCount <= 0 {
		return
	}
	fileName, ok := parsedInput.Flags["p"]

	if !ok {
		fmt.Println("Expected flag -p")
	}

	err := os.Remove(fileName)
	if err != nil {
		fmt.Println("error occured ", err)
	}
}

func handleRename(parsedInput util.DataStruct) {
	if parsedInput.FlagCount <= 0 {
		return
	}
	fileName, ok := parsedInput.Flags["f"]
	if !ok {
		fmt.Println("Expected flag -f")
		return
	}
	newName, ok := parsedInput.Flags["n"]
	if !ok {
		fmt.Println("Please provide new name with -n flag")
		return
	}

	err := os.Rename(fileName, newName)
	if err != nil {
		fmt.Printf("Cannot rename the file. Error: %+v\n", err)
		return
	}
}

func handleChw(parsedInput util.DataStruct) {
	if parsedInput.FlagCount <= 0 {
		return
	}

	for keys, value := range parsedInput.Flags {
		switch keys {
		case "p":
			fStat, err := os.Stat(value)
			if err != nil {
				fmt.Printf("%s is not a dir or  doesnot exist\n", value)
				return
			}

			if !fStat.IsDir() {
				fmt.Printf("%s is not a dir\n", value)
				return
			}

			err = os.Chdir(value)
			if err != nil {
				fmt.Printf("Error occured: %+v\n", err)
				return
			}
		default:
			fmt.Println("CWD has no such flags")
		}
	}
}
func handleList() {
	dirs, err := os.ReadDir(".")
	if err != nil {
		fmt.Println("Cannot read current dir")
		return
	}
	fmt.Printf("\n")
	for i, entry := range dirs {

		fmt.Printf("\t%d.%-30s %5s\n", i+1, entry.Name(), entry.Type())
	}
}

func WriteToHost(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	inputReader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf(">>")
		userInput, _ := inputReader.ReadString('\n')
		parsedInput := util.ParseUserInput(userInput)
		ParseCmd(conn, parsedInput)
	}
}

func ReadFromHost(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	fileName := ""
	cmdName := ""

	previousResSize := 0
	var fileSize uint32
	// Use a bufferedreader to read commands from stdin
	// Read response from client
	readMetaData := true
	for {
		response, err := util.ReadBytes(conn)
		if err != nil {
			if err.Error() == "Connection Closed" {
				fmt.Println("Connection closed")
				os.Exit(0)
			}
			fmt.Println(err)
		}
		if readMetaData && len(response) > 0 {

			//extract the length of filename
			stIdx := 0
			endIdx := 4

			cmdNameLenBuffer := response[stIdx:endIdx]
			cmdNameLen := ConvertByteToUint32(cmdNameLenBuffer)

			fmt.Println(cmdNameLen, " ", string(cmdNameLenBuffer))
			stIdx = endIdx
			endIdx = stIdx + int(cmdNameLen)

			cmdName = string(response[stIdx:endIdx])
			stIdx = endIdx
			endIdx = stIdx + 4

			fileNameLenBuffer := response[stIdx:endIdx]
			fileNameLen := ConvertByteToUint32(fileNameLenBuffer)

			// extract the filename
			stIdx = endIdx
			endIdx = stIdx + int(fileNameLen)
			fileName = string(response[stIdx:endIdx])

			//extract the filesize
			stIdx = endIdx
			endIdx = stIdx + 4
			fileSizeBuffer := response[stIdx:endIdx]
			fileSize = ConvertByteToUint32(fileSizeBuffer)
			readMetaData = false
		} else if !readMetaData && len(response) > 0 {
			switch cmdName {
			case "SEND":
				readMetaData = false
				handleDataWrite(conn, fileName, fileSize, &previousResSize, response)

			}
		} else {
			previousResSize = 0
			readMetaData = false
		}
	}

}
