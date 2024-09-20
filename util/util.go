package util

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sauravkattel/ftp/lexer"
)

type DataStruct struct {
	CmdName   string
	FlagCount int
	Flags     map[string]string
}

func consume(lex *lexer.Lexer, tkn *lexer.Token) {
	*tkn = lex.GetNextToken()
}

func ParseUserInput(str string) DataStruct {
	cmdName := ""
	flagCount := 0
	flags := map[string]string{}

	lex := lexer.Lexer{}
	lex.LoadLexer(str)

	tkn := lex.GetNextToken()

	for tkn.TokenType != lexer.EOF {

		if tkn.TokenType == lexer.STRING {
			cmdName = tkn.Value
			consume(&lex, &tkn)
		} else if tkn.TokenType == lexer.MINUS {
			flaValue := ""
			consume(&lex, &tkn) // takes -

			flagName := tkn.Value
			consume(&lex, &tkn) // stores the value after -,
			consume(&lex, &tkn) // consumes =
			for tkn.TokenType != lexer.MINUS && tkn.TokenType != lexer.EOF {
				flaValue += tkn.Value
				consume(&lex, &tkn)
			}
			flags[flagName] = flaValue
			flagCount++
		} else {
			consume(&lex, &tkn)
		}

	}

	return DataStruct{
		CmdName:   cmdName,
		FlagCount: flagCount,
		Flags:     flags,
	}
}

func ConvertIntoBytes(parsedInput DataStruct) []byte {
	data, jsonErr := json.Marshal(parsedInput)
	if jsonErr != nil {
		fmt.Printf("Error converting parsedData into json byte %v\n", jsonErr)
		return []byte{}
	}
	return data
}

func ReadBytes(conn net.Conn) ([]byte, error) {
	lengthBytes := make([]byte, 4)
	_, err := io.ReadFull(conn, lengthBytes)

	if err != nil {
		if err == io.EOF {
			return []byte{}, fmt.Errorf("Connection Closed")
		}
		return []byte{}, err
	}

	length := binary.BigEndian.Uint32(lengthBytes)
	buffer := make([]byte, length)
	_, err = io.ReadFull(conn, buffer)
	if err != nil {
		return []byte{}, err

	}
	return buffer, nil
}

func WriteBytes(conn net.Conn, dataBytes []byte) {
	length := uint32(len(dataBytes))
	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, length)
	messageByte := append(lengthBytes, dataBytes...)
	_, err := conn.Write(messageByte)
	if err != nil {
		fmt.Printf("Error sending command: %v\n", err)
		return
	}
}
