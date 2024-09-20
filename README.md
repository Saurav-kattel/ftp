# Basic ftp implementation in go
## How to run?
    Needs go version 1.23.0 
    To host a server run go run main.go -h HOST
      it will log the ip of the server.
      Listens of port 4000
    To connect as a client run go run main.go -ip [Ip of the server] -p [Port]
    
  ##  Currently available commands
  --For more information run HELP
  
      LIST: List the content of directory
      HELP: List all the available commands and their description
      CWD: Changes the directory
      SEND: Sends data over a network
      REN: Rename the file or directory
