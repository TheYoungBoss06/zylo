package lexer

import (
	"fmt"
	"testing"
)

func TestNextToken(t *testing.T) {
	input := `
var five = 5;
const ten = 10.5;

func add(x, y) {
  return x + y;
}

"hello"
'world'
"""
multi
line
"""
// Prueba de escapes: \n \t \" \\ \u0041
"linea1\nlinea2"
"Unicode: \u0041\u0042\u0043"
`
	// Nota: El lexer actual ignora los NEWLINEs en skipWhitespace.
	// Para probarlos, necesitaríamos una configuración o modificar esa lógica.
	// Por ahora, los tests no incluirán NEWLINE, asumiendo que se omiten.

	tests := []struct {
		expectedType      TokenType
		expectedLexeme    string
		expectedLiteral   interface{}
		expectedStartLine int
		expectedStartCol  int
		expectedEndLine   int
		expectedEndCol    int
	}{
		{NEWLINE, "\n", nil, 1, 1, 1, 1},
		{VAR, "var", nil, 2, 1, 2, 3},
		{IDENTIFIER, "five", nil, 2, 5, 2, 8},
		{EQUAL, "=", nil, 2, 10, 2, 10},
		{NUMBER, "5", int64(5), 2, 12, 2, 12},
		{SEMICOLON, ";", nil, 2, 13, 2, 13},
		{NEWLINE, "\n", nil, 2, 14, 2, 14},
		{CONST, "const", nil, 3, 1, 3, 5},
		{IDENTIFIER, "ten", nil, 3, 7, 3, 9},
		{EQUAL, "=", nil, 3, 11, 3, 11},
		{NUMBER, "10.5", float64(10.5), 3, 13, 3, 16},
		{SEMICOLON, ";", nil, 3, 17, 3, 17},
		{NEWLINE, "\n", nil, 3, 18, 3, 18},
		{NEWLINE, "\n", nil, 4, 1, 4, 1},
		{FUNC, "func", nil, 5, 1, 5, 4},
		{IDENTIFIER, "add", nil, 5, 6, 5, 8},
		{LEFT_PAREN, "(", nil, 5, 9, 5, 9},
		{IDENTIFIER, "x", nil, 5, 10, 5, 10},
		{COMMA, ",", nil, 5, 11, 5, 11},
		{IDENTIFIER, "y", nil, 5, 13, 5, 13},
		{RIGHT_PAREN, ")", nil, 5, 14, 5, 14},
		{LEFT_BRACE, "{", nil, 5, 16, 5, 16},
		{NEWLINE, "\n", nil, 5, 17, 5, 17},
		{RETURN, "return", nil, 6, 3, 6, 8},
		{IDENTIFIER, "x", nil, 6, 10, 6, 10},
		{PLUS, "+", nil, 6, 12, 6, 12},
		{IDENTIFIER, "y", nil, 6, 14, 6, 14},
		{SEMICOLON, ";", nil, 6, 15, 6, 15},
		{NEWLINE, "\n", nil, 6, 16, 6, 16},
		{RIGHT_BRACE, "}", nil, 7, 1, 7, 1},
		{NEWLINE, "\n", nil, 7, 2, 7, 2},
		{NEWLINE, "\n", nil, 8, 1, 8, 1},
		{STRING, `"hello"`, "hello", 9, 1, 9, 7},
		{NEWLINE, "\n", nil, 9, 8, 9, 8},
		{STRING, "'world'", "world", 10, 1, 10, 7},
		{NEWLINE, "\n", nil, 10, 8, 10, 8},
		{STRING, "\"\"\"\nmulti\nline\n\"\"\"", "multi\nline\n", 11, 1, 14, 3},
		{NEWLINE, "\n", nil, 14, 4, 14, 4},
		{NEWLINE, "\n", nil, 15, 41, 15, 41}, // El comentario de línea termina en la columna 40, el newline empieza en 41.
		{STRING, `"linea1\nlinea2"`, "linea1\nlinea2", 16, 1, 16, 16},
		{NEWLINE, "\n", nil, 16, 17, 16, 17},
		{STRING, `"Unicode: \u0041\u0042\u0043"`, "Unicode: ABC", 17, 1, 17, 29},
		{NEWLINE, "\n", nil, 17, 30, 17, 30},
		{EOF, "", nil, 18, 1, 18, 0},
	}

	l := New(input)

	for i, tt := range tests {
		t.Run(fmt.Sprintf("Test %d: %s", i, tt.expectedLexeme), func(t *testing.T) {
			tok := l.NextToken()

			if tok.Type != tt.expectedType {
				t.Errorf("wrong tokentype. expected=%q, got=%q (lexeme: '%s')",
					tt.expectedType, tok.Type, tok.Lexeme)
			}

			if tok.Lexeme != tt.expectedLexeme {
				t.Errorf("wrong lexeme. expected=%q, got=%q (type: %s)",
					tt.expectedLexeme, tok.Lexeme, tok.Type)
			}

			if tok.Literal != tt.expectedLiteral {
				t.Errorf("wrong literal. expected=%v(%T), got=%v(%T)",
					tt.expectedLiteral, tt.expectedLiteral, tok.Literal, tok.Literal)
			}

			if tok.StartLine != tt.expectedStartLine {
				t.Errorf("wrong start line. expected=%d, got=%d", tt.expectedStartLine, tok.StartLine)
			}
			if tok.StartCol != tt.expectedStartCol {
				t.Errorf("wrong start col. expected=%d, got=%d", tt.expectedStartCol, tok.StartCol)
			}
			if tok.EndLine != tt.expectedEndLine {
				t.Errorf("wrong end line. expected=%d, got=%d", tt.expectedEndLine, tok.EndLine)
			}
			if tok.EndCol != tt.expectedEndCol {
				t.Errorf("wrong end col. expected=%d, got=%d", tt.expectedEndCol, tok.EndCol)
			}
		})
	}
}
func BenchmarkLex(b *testing.B) {
	input := `var five = 5;
const ten = 10.5;
func add(x, y) {
  return x + y;
}
"hello"
'world'
"""multi
line"""
// Unicode: \u0041\u0042\u0043
"linea1\nlinea2"
`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := New(input)
		for {
			tok := l.NextToken()
			if tok.Type == EOF {
				break
			}
		}
	}
}
