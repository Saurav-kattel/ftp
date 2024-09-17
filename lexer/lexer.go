package lexer

import (
	"unicode"
)

type TokenType int

const (
	MINUS TokenType = iota
	EQUAL
	STRING
	EOF
	ERR
	DOT
	SLASH
)

type Lexer struct {
	cmd string
	pos int // current position in the string
}

type Token struct {
	TokenType TokenType
	Value     string
}

func (l *Lexer) LoadLexer(src string) {
	l.cmd = src
	l.pos = 0 // reset position
}

func (l *Lexer) GetNextToken() Token {
	// Skip whitespace
	for l.pos < len(l.cmd) && unicode.IsSpace(rune(l.cmd[l.pos])) {
		l.pos++
	}

	if l.pos >= len(l.cmd) {
		return Token{
			TokenType: EOF,
			Value:     "eof",
		}
	}

	ch := l.cmd[l.pos]

	switch ch {
	case '-':
		l.pos++
		return Token{
			TokenType: MINUS,
			Value:     "-",
		}
	case '+':
		l.pos++
		return Token{
			TokenType: EQUAL,
			Value:     "=",
		}
	}

	if unicode.IsLetter(rune(ch)) || rune(ch) == '.' || rune(ch) == '/' {
		buffer := ""
		for l.pos < len(l.cmd) && unicode.IsLetter(rune(l.cmd[l.pos])) && (rune(ch) == '.' || rune(ch) == '/') {
			buffer += string(l.cmd[l.pos])
			l.pos++
		}
		return Token{
			TokenType: STRING,
			Value:     buffer,
		}
	}

	// Default case for unknown characters
	l.pos++
	return Token{
		TokenType: ERR,
		Value:     string(ch),
	}
}
