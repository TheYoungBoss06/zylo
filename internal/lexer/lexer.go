package lexer

import (
	"strconv"
	"strings"
	"unicode"
)

// Lexer se encarga de convertir el código fuente en una secuencia de tokens.
type Lexer struct {
	source      []rune // El código fuente como un slice de runas para soportar Unicode.
	start       int    // Posición de inicio del token actual.
	current     int    // Posición actual en el slice de runas.
	line        int    // Línea actual.
	column      int    // Columna actual en la línea.
	startLine   int    // Línea de inicio del token actual.
	startColumn int    // Columna de inicio del token actual.
}

// New crea un nuevo Lexer para el código fuente proporcionado.
func New(source string) *Lexer {
	return &Lexer{
		source: []rune(source),
		line:   1,
		column: 1,
	}
}

// isAtEnd comprueba si hemos llegado al final del código fuente.
func (l *Lexer) isAtEnd() bool {
	return l.current >= len(l.source)
}

// advance consume la runa actual y avanza la posición.
func (l *Lexer) advance() rune {
	if l.isAtEnd() {
		return 0 // EOF
	}
	r := l.source[l.current]
	l.current++
	if r == '\n' {
		l.line++
		l.column = 0 // Se reinicia y el siguiente caracter será la columna 1
	}
	l.column++
	return r
}

// peek devuelve la runa actual sin consumirla.
func (l *Lexer) peek() rune {
	if l.isAtEnd() {
		return 0
	}
	return l.source[l.current]
}

// peekNext devuelve la siguiente runa sin consumirla.
func (l *Lexer) peekNext() rune {
	if l.current+1 >= len(l.source) {
		return 0
	}
	return l.source[l.current+1]
}

// match comprueba si la runa actual coincide con la esperada. Si es así, la consume.
func (l *Lexer) match(expected rune) bool {
	if l.isAtEnd() || l.source[l.current] != expected {
		return false
	}
	l.advance()
	return true
}

// makeToken crea un nuevo token con la información de posición actual.
func (l *Lexer) makeToken(tokenType TokenType, literal interface{}) Token {
	lexeme := string(l.source[l.start:l.current])

	// Para tokens de newline, la posición final es la misma que la inicial
	if tokenType == NEWLINE {
		return Token{
			Type:      tokenType,
			Lexeme:    lexeme,
			Literal:   literal,
			StartLine: l.startLine,
			StartCol:  l.startColumn,
			EndLine:   l.startLine,
			EndCol:    l.startColumn, // Un newline ocupa solo una columna
		}
	}

	// Calcular la posición final correctamente
	endLine := l.startLine
	endCol := l.startColumn + len(lexeme) - 1

	// Si el token contiene newlines, ajustar la posición final
	if l.startLine != l.line {
		endLine = l.line
		endCol = l.column - 1
		if endCol < 1 {
			endCol = 1
		}
	}

	return Token{
		Type:      tokenType,
		Lexeme:    lexeme,
		Literal:   literal,
		StartLine: l.startLine,
		StartCol:  l.startColumn,
		EndLine:   endLine,
		EndCol:    endCol,
	}
}

// errorToken crea un token de error.
func (l *Lexer) errorToken(message string) Token {
	return Token{
		Type:      "ERROR",
		Lexeme:    message,
		StartLine: l.line,
		StartCol:  l.column,
		EndLine:   l.line,
		EndCol:    l.column,
	}
}

// skipWhitespace consume todos los espacios en blanco y tabulaciones, pero no los newlines.
func (l *Lexer) skipWhitespace() {
	for {
		switch l.peek() {
		case ' ', '\r', '\t':
			l.advance()
		case '\n':
			// No consumir newlines aquí, dejarlos para que sean tokens
			return
		case '/':
			if l.peekNext() == '/' { // Comentario de una línea
				for l.peek() != '\n' && !l.isAtEnd() {
					l.advance()
				}
				// No consumir el newline aquí, dejarlo para que sea un token
			} else if l.peekNext() == '*' { // Comentario multilínea /* ... */
				l.advance() // Consume '/'
				l.advance() // Consume '*'
				l.skipMultiLineComment()
			} else {
				return
			}
		case '#': // Soporte para comentarios de una línea con #
			for l.peek() != '\n' && !l.isAtEnd() {
				l.advance()
			}
			// No consumir el newline aquí, dejarlo para que sea un token
		default:
			return
		}
	}
}

// skipMultiLineComment consume un comentario multilínea, incluyendo anidamiento.
func (l *Lexer) skipMultiLineComment() {
	nestingLevel := 1
	for nestingLevel > 0 && !l.isAtEnd() {
		if l.peek() == '*' && l.peekNext() == '/' {
			l.advance() // Consume '*'
			l.advance() // Consume '/'
			nestingLevel--
		} else if l.peek() == '/' && l.peekNext() == '*' {
			l.advance() // Consume '/'
			l.advance() // Consume '*'
			nestingLevel++
		} else {
			l.advance()
		}
	}
	if nestingLevel > 0 {
		// Error: comentario multilínea sin terminar.
		// Por ahora, no emitimos un token de error aquí, pero el lexer podría hacerlo.
	}
}

// isAlpha comprueba si una runa es una letra o un guion bajo.
func isAlpha(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

// isDigit comprueba si una runa es un dígito.
func isDigit(r rune) bool {
	return unicode.IsDigit(r)
}

// isHexDigit comprueba si una runa es un dígito hexadecimal.
func isHexDigit(r rune) bool {
	return unicode.IsDigit(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

// identifier procesa un identificador o una palabra clave.
func (l *Lexer) identifier() Token {
	for isAlpha(l.peek()) || isDigit(l.peek()) {
		l.advance()
	}
	text := string(l.source[l.start:l.current])
	tokenType, isKeyword := keywords[text]
	if !isKeyword {
		tokenType = IDENTIFIER
	}
	return l.makeToken(tokenType, nil)
}

// number procesa un número literal.
func (l *Lexer) number() Token {
	isFloat := false
	for isDigit(l.peek()) {
		l.advance()
	}
	if l.peek() == '.' && isDigit(l.peekNext()) {
		isFloat = true
		l.advance() // Consume el '.'
		for isDigit(l.peek()) {
			l.advance()
		}
	}

	lexeme := string(l.source[l.start:l.current])
	if isFloat {
		value, err := strconv.ParseFloat(lexeme, 64)
		if err != nil {
			return l.errorToken("Invalid float number.")
		}
		return l.makeToken(NUMBER, value)
	}

	value, err := strconv.ParseInt(lexeme, 10, 64)
	if err != nil {
		return l.errorToken("Invalid integer number.")
	}
	return l.makeToken(NUMBER, value)
}

// stringLiteral procesa una cadena literal entre comillas simples o dobles.
func (l *Lexer) stringLiteral(quote rune) Token {
	var builder strings.Builder
	for {
		if l.peek() == quote || l.isAtEnd() {
			break
		}

		if l.peek() == '\\' { // Secuencia de escape
			l.advance() // consume '\'
			switch l.peek() {
			case 'n':
				builder.WriteRune('\n')
			case 't':
				builder.WriteRune('\t')
			case '"':
				builder.WriteRune('"')
			case '\'':
				builder.WriteRune('\'')
			case '\\':
				builder.WriteRune('\\')
			case 'u':
				l.advance() // consume 'u'
				hex := make([]rune, 4)
				for i := 0; i < 4; i++ {
					if !isHexDigit(l.peek()) {
						return l.errorToken("Invalid Unicode escape sequence: expected 4 hex digits.")
					}
					hex[i] = l.advance()
				}
				hexVal, err := strconv.ParseInt(string(hex), 16, 32)
				if err != nil {
					return l.errorToken("Invalid Unicode escape sequence.")
				}
				builder.WriteRune(rune(hexVal))
				continue // Evitar el l.advance() de abajo
			default:
				builder.WriteRune('\\')
				builder.WriteRune(l.peek())
			}
			l.advance()
		} else {
			// No permitir newlines en strings normales
			if l.peek() == '\n' {
				return l.errorToken("Unterminated string.")
			}
			builder.WriteRune(l.advance())
		}
	}

	if l.isAtEnd() {
		return l.errorToken("Unterminated string.")
	}

	l.advance() // Consume la comilla de cierre.
	return l.makeToken(STRING, builder.String())
}

// tripleQuotedStringLiteral procesa una cadena multilínea.
func (l *Lexer) tripleQuotedStringLiteral() Token {
	l.advance() // consume second "
	l.advance() // consume third "

	var builder strings.Builder
	for {
		if l.isAtEnd() {
			return l.errorToken("Unterminated multi-line string.")
		}
		if l.peek() == '"' && l.peekNext() == '"' && l.peekN(2) == '"' {
			break
		}
		builder.WriteRune(l.advance())
	}

	l.advance() // consume first "
	l.advance() // consume second "
	l.advance() // consume third "

	// Eliminar el newline inicial si existe, como en Python
	content := builder.String()
	if len(content) > 0 && content[0] == '\n' {
		content = content[1:]
	}

	return l.makeToken(STRING, content)
}

// peekN devuelve la runa en la posición current + n.
func (l *Lexer) peekN(n int) rune {
	if l.current+n >= len(l.source) {
		return 0
	}
	return l.source[l.current+n]
}

// NextToken escanea y devuelve el siguiente token del código fuente.
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()
	l.start = l.current
	l.startLine = l.line
	l.startColumn = l.column

	if l.isAtEnd() {
		return l.makeToken(EOF, nil)
	}

	r := l.advance()

	if isAlpha(r) {
		return l.identifier()
	}
	if isDigit(r) {
		return l.number()
	}

	switch r {
	case '(':
		return l.makeToken(LEFT_PAREN, nil)
	case ')':
		return l.makeToken(RIGHT_PAREN, nil)
	case '{':
		return l.makeToken(LEFT_BRACE, nil)
	case '}':
		return l.makeToken(RIGHT_BRACE, nil)
	case '[':
		return l.makeToken(LEFT_BRACKET, nil)
	case ']':
		return l.makeToken(RIGHT_BRACKET, nil)
	case ';':
		return l.makeToken(SEMICOLON, nil)
	case ',':
		return l.makeToken(COMMA, nil)
	case '.':
		return l.makeToken(DOT, nil)
	case '-':
		return l.makeToken(MINUS, nil)
	case '+':
		return l.makeToken(PLUS, nil)
	case '/':
		return l.makeToken(SLASH, nil)
	case '*':
		return l.makeToken(STAR, nil)
	case '%':
		return l.makeToken(PERCENT, nil)
	case ':':
		return l.makeToken(COLON, nil)
	case '!':
		if l.match('=') {
			return l.makeToken(BANG_EQUAL, nil)
		}
		return l.makeToken(BANG, nil)
	case '=':
		if l.match('=') {
			return l.makeToken(EQUAL_EQUAL, nil)
		}
		if l.match('>') {
			return l.makeToken(ARROW, nil)
		}
		return l.makeToken(EQUAL, nil)
	case '<':
		if l.match('=') {
			return l.makeToken(LESS_EQUAL, nil)
		}
		return l.makeToken(LESS, nil)
	case '>':
		if l.match('=') {
			return l.makeToken(GREATER_EQUAL, nil)
		}
		return l.makeToken(GREATER, nil)
	case '&':
		if l.match('&') {
			return l.makeToken(AND, nil)
		}
		return l.errorToken("Unexpected character '&'. Did you mean '&&'?")
	case '|':
		if l.match('|') {
			return l.makeToken(OR, nil)
		}
		return l.errorToken("Unexpected character '|'. Did you mean '||'?")
	case '\n':
		return l.makeToken(NEWLINE, nil)
	case '"':
		if l.peek() == '"' && l.peekNext() == '"' {
			return l.tripleQuotedStringLiteral()
		}
		return l.stringLiteral('"')
	case '\'':
		return l.stringLiteral('\'')
	}

	return l.errorToken("Unexpected character.")
}

var keywords = map[string]TokenType{
	"and":      AND,
	"class":    CLASS,
	"else":     ELSE,
	"false":    FALSE,
	"for":      FOR,
	"func":     FUNC,
	"if":       IF,
	"nil":      NIL,
	"or":       OR,
	"return":   RETURN,
	"super":    SUPER,
	"this":     THIS,
	"true":     TRUE,
	"var":      VAR,
	"const":    CONST,
	"while":    WHILE,
	"break":    BREAK,
	"continue": CONTINUE,
	"elif":     ELIF,
	// "show" and "log" are treated as regular identifiers for member access
	"import":   IMPORT,
	"from":     FROM,
	"try":      TRY,
	"catch":    CATCH,
	"throw":    THROW,
	"finally":  FINALLY,
	"async":    ASYNC,
	"await":    AWAIT,
	"spawn":    SPAWN,
	"in":       IN,
}
