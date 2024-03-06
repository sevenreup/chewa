package lexer

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"unicode"

	"github.com/sevenreup/chewa/src/token"
)

type Lexer struct {
	r   *bufio.Reader
	pos token.Position
}

type TokenInfo struct {
	Position token.Position
	Token    token.TokenType
	Literal  string
}

func New(path []byte) *Lexer {
	return &Lexer{
		r:   bufio.NewReader(bytes.NewReader(path)),
		pos: token.Position{Line: 1, Column: 0},
	}
}

func (l *Lexer) NextToken() token.Token {
	for {
		r, _, err := l.r.ReadRune()
		if err != nil {
			if err == io.EOF {
				return newToken(l.pos, token.EOF, "")
			}
			panic(err)
		}

		l.pos.Column++
		switch r {
		case '\n':
			l.resetPosition()
		case ';':
			return newToken(l.pos, token.SEMICOLON, ";")
		case '"', '\'':
			return l.ReadString()
		case '+':
			return newToken(l.pos, token.PLUS, "+")
		case '-':
			return newToken(l.pos, token.MINUS, "-")
		case ':':
			return newToken(l.pos, token.COLON, ":")
		case '/':
			nextRune := l.Peek()
			if nextRune == '/' {
				l.Next()
				return l.ReadComment()
			} else if nextRune == '*' {
				var builder strings.Builder
				builder.WriteString("/")
				for {
					r, _, err := l.r.ReadRune()
					if err != nil {
						if err == io.EOF {
							return newToken(l.pos, token.EOF, "")
						}
						panic(err)
					}
					l.pos.Column++
					if r == '*' {
						nextRune = l.Peek()
						if nextRune == '/' {
							l.Next()
							break
						}
					}
					builder.WriteString(string(r))
				}
				builder.WriteString("*/")
				return newToken(l.pos, token.MULTILINE_COMMENT, builder.String())
			}
			return newToken(l.pos, token.DIVIDE, "/")
		case '{':
			return newToken(l.pos, token.OPENING_BRACE, "{")
		case '}':
			return newToken(l.pos, token.CLOSING_BRACE, "}")
		case '(':
			return newToken(l.pos, token.OPENING_PAREN, "(")
		case ')':
			return newToken(l.pos, token.CLOSING_PAREN, ")")
		case '.':
			return newToken(l.pos, token.FULL_STOP, ".")
		case ',':
			return newToken(l.pos, token.COMMA, ",")
		case '=':
			nextRune := l.Peek()
			if nextRune == '=' {
				l.Next()
				return newToken(l.pos, token.EQUAL_TO, "==")
			}
			return newToken(l.pos, token.ASSIGN, "=")
		case '>':
			nextRune := l.Peek()
			if nextRune == '=' {
				l.Next()
				return newToken(l.pos, token.GREATER_THAN_OR_EQUAL_TO, ">=")
			}
			return newToken(l.pos, token.GREATER_THAN, ">")
		case '<':
			nextRune := l.Peek()
			if nextRune == '=' {
				l.Next()
				return newToken(l.pos, token.LESS_THAN_OR_EQUAL_TO, "<=")
			}
			return newToken(l.pos, token.LESS_THAN, "<")
		case '!':
			nextRune := l.Peek()
			if nextRune == '=' {
				l.Next()
				return newToken(l.pos, token.NOT_EQUAL_TO, "!=")
			}
		default:
			if unicode.IsSpace(r) {
				continue
			} else if unicode.IsLetter(r) {
				return l.ReadIdentifier(r)
			} else if unicode.IsDigit(r) {
				return l.ReadNumber(r)
			}
			return newToken(l.pos, token.STRING, string(r))
		}
	}
}

func (l *Lexer) Next() rune {
	r, _, err := l.r.ReadRune()
	if err != nil {
		return rune(0)
	}
	l.pos.Column++
	return r
}

func (l *Lexer) Peek() rune {
	r, _, _ := l.r.ReadRune()
	l.r.UnreadRune()
	return r
}

func (l *Lexer) ReadComment() token.Token {
	rawString := ""
	var newPos token.Position
	for {
		r, _, err := l.r.ReadRune()
		if err != nil {
			if err == io.EOF {
				return newToken(l.pos, token.EOF, "")
			}
			panic(err)
		}
		l.pos.Column++
		if r == '\n' {
			newPos = l.pos
			l.resetPosition()
			break
		} else {
			rawString += string(r)
		}
	}
	return newToken(newPos, token.COMMENT, rawString)
}

func (l *Lexer) ReadString() token.Token {
	rawString := ""
	for {
		r, _, err := l.r.ReadRune()
		if err != nil {
			if err == io.EOF {
				return newToken(l.pos, token.EOF, "")
			}
			panic(err)
		}
		l.pos.Column++
		if r == '"' || r == '\'' {
			break
		} else {
			rawString += string(r)
		}
	}
	return newToken(l.pos, token.STRING, rawString)
}

func (l *Lexer) ReadNumber(current rune) token.Token {
	number := string(current)
	for {
		r, _, err := l.r.ReadRune()
		if err != nil {
			if err == io.EOF {
				return newToken(l.pos, token.EOF, "")
			}
			panic(err)
		}
		l.pos.Column++
		if unicode.IsDigit(r) {
			number += string(r)
		} else {
			l.r.UnreadRune()
			break
		}
	}
	return newToken(l.pos, token.INT, number)
}

func (l *Lexer) ReadIdentifier(current rune) token.Token {
	identifier := string(current)
	for {
		r, _, err := l.r.ReadRune()
		if err != nil {
			if err == io.EOF {
				return newToken(l.pos, token.EOF, "")
			}
			panic(err)
		}
		l.pos.Column++
		if validIdentifierSymbol(r) {
			identifier += string(r)
		} else {
			l.r.UnreadRune()
			break
		}
	}
	return newToken(l.pos, token.LookupIdent(identifier), identifier)
}

func validIdentifierSymbol(symbol rune) bool {
	return unicode.IsLetter(symbol) || unicode.IsDigit(symbol) || symbol == '_'
}

func (l *Lexer) resetPosition() {
	l.pos.Line++
	l.pos.Column = 0
}

func newToken(pos token.Position, tokenType token.TokenType, ch string) token.Token {
	return token.Token{Type: tokenType, Literal: ch, Pos: pos}
}

func (l *Lexer) AccumTokens() []TokenInfo {
	var tokens []TokenInfo
	for {
		tok := l.NextToken()
		if tok.Type == token.EOF {
			break
		}
		info := TokenInfo{
			Position: tok.Pos,
			Token:    tok.Type,
			Literal:  tok.Literal,
		}
		tokens = append(tokens, info)
	}

	return tokens
}
