	package lexer

	type TokenType string

	const (
		LBRACE   TokenType = "LBRACE"
		RBRACE   TokenType = "RBRACE"
		LBRACKET TokenType = "LBRACKET"
		RBRACKET TokenType = "RBRACKET"
		COLON    TokenType = "COLON"
		COMMA    TokenType = "COMMA"
		STRING   TokenType = "STRING"
		NUMBER   TokenType = "NUMBER"
		TRUE     TokenType = "TRUE"
		FALSE    TokenType = "FALSE"
		NULL     TokenType = "NULL"
		EOF      TokenType = "EOF"
		ILLEGAL  TokenType = "ILLEGAL"
	)

	type Token struct {
		Type    TokenType
		Literal string
		Pos     int
	}

	type Lexer struct {
		pos   int
		input string
	}

	func NewLexer(input string) *Lexer {
		return &Lexer{
			input: input,
		}
	}

	func isJSONWhitespace(b byte) bool {
		return b == 0x20 || b == 0x09 || b == 0x0A || b == 0x0D
	}

	func (l *Lexer) skipWhitespaces() {
		for l.pos < len(l.input) && isJSONWhitespace(l.input[l.pos]) {
			l.pos++
		}
	}

	func getTokenType(b byte) TokenType {
		switch b {
		case '{':
			return LBRACE
		case '}':
			return RBRACE
		case '[':
			return LBRACKET
		case ']':
			return RBRACKET
		case ':':
			return COLON
		case ',':
			return COMMA
		default:
			return ILLEGAL
		}
	}

	func (l *Lexer) NextToken() *Token {
		l.skipWhitespaces()

		if l.pos >= len(l.input) {
			return &Token{Type: EOF}
		}

		currentPos := l.pos
		ch := l.input[l.pos]
		l.pos++

		return &Token{
			Type:    getTokenType(ch),
			Literal: string(ch),
			Pos:     currentPos,
		}
	}
