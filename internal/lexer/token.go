package lexer

import "fmt"

// TokenType es un string que representa el tipo de un token.
type TokenType string

// Token representa una unidad léxica del lenguaje Zylo.
type Token struct {
	Type      TokenType // El tipo del token (e.g., IDENTIFIER, NUMBER).
	Lexeme    string    // El substring del código fuente que representa el token.
	Literal   interface{} // El valor literal del token, si aplica (e.g., 123, "hello").
	StartLine int       // La línea donde comienza el token.
	StartCol  int       // La columna donde comienza el token.
	EndLine   int       // La línea donde termina el token.
	EndCol    int       // La columna donde termina el token.
}

// String devuelve una representación legible del token, útil para debugging.
func (t Token) String() string {
	return fmt.Sprintf("Token(Type: %s, Lexeme: '%s', Literal: %v, Pos: %d:%d-%d:%d)",
		t.Type, t.Lexeme, t.Literal, t.StartLine, t.StartCol, t.EndLine, t.EndCol)
}

// Constantes para los tipos de token.
const (
	// Tokens de un solo carácter
	LEFT_PAREN  TokenType = "LEFT_PAREN"
	RIGHT_PAREN TokenType = "RIGHT_PAREN"
	LEFT_BRACE  TokenType = "LEFT_BRACE"
	RIGHT_BRACE TokenType = "RIGHT_BRACE"
	LEFT_BRACKET TokenType = "LEFT_BRACKET"
	RIGHT_BRACKET TokenType = "RIGHT_BRACKET"
	COMMA       TokenType = "COMMA"
	DOT         TokenType = "DOT"
	MINUS       TokenType = "MINUS"
	PLUS        TokenType = "PLUS"
	SEMICOLON   TokenType = "SEMICOLON"
	SLASH       TokenType = "SLASH"
	STAR        TokenType = "STAR"
	PERCENT     TokenType = "PERCENT"
	COLON       TokenType = "COLON"

	// Tokens de uno o dos caracteres
	BANG          TokenType = "BANG"
	BANG_EQUAL    TokenType = "BANG_EQUAL"
	EQUAL         TokenType = "EQUAL"
	EQUAL_EQUAL   TokenType = "EQUAL_EQUAL"
	GREATER       TokenType = "GREATER"
	GREATER_EQUAL TokenType = "GREATER_EQUAL"
	LESS          TokenType = "LESS"
	LESS_EQUAL    TokenType = "LESS_EQUAL"
	ARROW         TokenType = "ARROW" // =>

	// Literales
	IDENTIFIER TokenType = "IDENTIFIER"
	STRING     TokenType = "STRING"
	NUMBER     TokenType = "NUMBER"

	// Palabras clave
	AND      TokenType = "AND"
	CLASS    TokenType = "CLASS"
	ELSE     TokenType = "ELSE"
	ELIF     TokenType = "ELIF"
	FALSE    TokenType = "FALSE"
	FOR      TokenType = "FOR"
	FUNC     TokenType = "FUNC"
	IF       TokenType = "IF"
	NIL      TokenType = "NIL"
	OR       TokenType = "OR"
	RETURN   TokenType = "RETURN"
	SUPER    TokenType = "SUPER" // Placeholder, puede cambiar
	THIS     TokenType = "THIS"    // Placeholder, puede cambiar
	TRUE     TokenType = "TRUE"
	VAR      TokenType = "VAR"
	CONST    TokenType = "CONST"
	WHILE    TokenType = "WHILE"
	BREAK    TokenType = "BREAK"
	CONTINUE TokenType = "CONTINUE"
	SHOW     TokenType = "SHOW"
	LOG      TokenType = "LOG"
	IMPORT   TokenType = "IMPORT"
	FROM     TokenType = "FROM"
	TRY      TokenType = "TRY"
	CATCH    TokenType = "CATCH"
	THROW    TokenType = "THROW"
	FINALLY  TokenType = "FINALLY"
	ASYNC    TokenType = "ASYNC"
	AWAIT    TokenType = "AWAIT"
	SPAWN    TokenType = "SPAWN"
	IN       TokenType = "IN"

	// Control
	NEWLINE TokenType = "NEWLINE"
	EOF     TokenType = "EOF"
)
